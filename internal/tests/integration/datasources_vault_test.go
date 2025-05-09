//go:build !windows
// +build !windows

package integration

import (
	"context"
	"os"
	"os/user"
	"path"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

const vaultRootToken = "00000000-1111-2222-3333-444455556666"

func setupDatasourcesVaultTest(t *testing.T) *vaultClient {
	t.Helper()

	_, vaultClient := startVault(t)

	err := vaultClient.vc.Sys().PutPolicy("writepol", `path "*" {
  capabilities = ["create","update","delete"]
}`)
	require.NoError(t, err)
	err = vaultClient.vc.Sys().PutPolicy("readpol", `path "*" {
  capabilities = ["read","delete"]
}`)
	require.NoError(t, err)
	err = vaultClient.vc.Sys().PutPolicy("listpol", `path "*" {
  capabilities = ["read","list","delete"]
}`)
	require.NoError(t, err)

	return vaultClient
}

func startVault(t *testing.T) (*fs.Dir, *vaultClient) {
	t.Helper()

	pidDir := fs.NewDir(t, "gomplate-inttests-vaultpid")
	t.Cleanup(pidDir.Remove)

	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile("config.json", `{
		"pid_file": "`+pidDir.Join("vault.pid")+`"
		}`),
	)
	t.Cleanup(tmpDir.Remove)

	// rename any existing token so it doesn't get overridden
	u, _ := user.Current()
	homeDir := u.HomeDir
	tokenFile := path.Join(homeDir, ".vault-token")
	info, err := os.Stat(tokenFile)
	if err == nil && info.Mode().IsRegular() {
		os.Rename(tokenFile, path.Join(homeDir, ".vault-token.bak"))
	}

	_, vaultAddr := freeport(t)
	vault := icmd.Command("vault", "server",
		"-dev",
		"-dev-root-token-id="+vaultRootToken,
		"-dev-kv-v1", // default to v1, so we can test v1 and v2
		"-log-level=info",
		"-dev-listen-address="+vaultAddr,
		"-config="+tmpDir.Join("config.json"),
	)
	result := icmd.StartCmd(vault)

	t.Logf("Fired up Vault: %v", vault)

	err = waitForURL(t, "http://"+vaultAddr+"/v1/sys/health")
	require.NoError(t, err)

	vaultClient, err := createVaultClient(vaultAddr, vaultRootToken)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := result.Cmd.Process.Kill()
		require.NoError(t, err)

		result.Cmd.Wait()

		result.Assert(t, icmd.Expected{ExitCode: 0})

		t.Log(result.Combined())

		// restore old token if it was backed up
		u, _ := user.Current()
		homeDir := u.HomeDir
		tokenFile := path.Join(homeDir, ".vault-token.bak")
		info, err := os.Stat(tokenFile)
		if err == nil && info.Mode().IsRegular() {
			os.Rename(tokenFile, path.Join(homeDir, ".vault-token"))
		}
	})

	return tmpDir, vaultClient
}

func TestDatasources_Vault_TokenAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]any{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	tok, err := v.tokenCreate("readpol", 5)
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t, "-d", "vault=vault+http://"+v.addr+"/secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar")

	_, _, err = cmd(t, "-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "bar").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_TOKEN", tok).
		run()
	assert.ErrorContains(t, err, "error calling ds: couldn't read datasource 'vault'")
	assert.ErrorContains(t, err, "stat secret/bar")
	assert.ErrorContains(t, err, "file does not exist")

	tokFile := fs.NewFile(t, "test-vault-token", fs.WithContent(tok))
	defer tokFile.Remove()

	o, e, err = cmd(t, "-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_TOKEN_FILE", tokFile.Path()).
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_UserPassAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]any{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	err := v.vc.Sys().EnableAuth("userpass", "userpass", "")
	require.NoError(t, err)
	err = v.vc.Sys().EnableAuth("userpass2", "userpass", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("userpass")
	defer v.vc.Sys().DisableAuth("userpass2")
	_, err = v.vc.Logical().Write("auth/userpass/users/dave", map[string]any{
		"password": "foo", "ttl": "10s", "policies": "readpol",
	})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/userpass2/users/dave", map[string]any{
		"password": "bar", "ttl": "10s", "policies": "readpol",
	})
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_AUTH_USERNAME", "dave").
		withEnv("VAULT_AUTH_PASSWORD", "foo").
		run()
	assertSuccess(t, o, e, err, "bar")

	userFile := fs.NewFile(t, "test-vault-user", fs.WithContent("dave"))
	passFile := fs.NewFile(t, "test-vault-pass", fs.WithContent("foo"))
	defer userFile.Remove()
	defer passFile.Remove()
	o, e, err = cmd(t,
		"-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_AUTH_USERNAME_FILE", userFile.Path()).
		withEnv("VAULT_AUTH_PASSWORD_FILE", passFile.Path()).
		run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t,
		"-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_AUTH_USERNAME", "dave").
		withEnv("VAULT_AUTH_PASSWORD", "bar").
		withEnv("VAULT_AUTH_USERPASS_MOUNT", "userpass2").
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_AppRoleAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]any{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	err := v.vc.Sys().EnableAuth("approle", "approle", "")
	require.NoError(t, err)
	err = v.vc.Sys().EnableAuth("approle2", "approle", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("approle")
	defer v.vc.Sys().DisableAuth("approle2")
	_, err = v.vc.Logical().Write("auth/approle/role/testrole", map[string]any{
		"secret_id_ttl": "10s", "token_ttl": "20s",
		"secret_id_num_uses": "1", "policies": "readpol",
	})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/approle2/role/testrole", map[string]any{
		"secret_id_ttl": "10s", "token_ttl": "20s",
		"secret_id_num_uses": "1", "policies": "readpol",
	})
	require.NoError(t, err)

	rid, _ := v.vc.Logical().Read("auth/approle/role/testrole/role-id")
	roleID := rid.Data["role_id"].(string)
	sid, _ := v.vc.Logical().Write("auth/approle/role/testrole/secret-id", nil)
	secretID := sid.Data["secret_id"].(string)
	o, e, err := cmd(t,
		"-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_ROLE_ID", roleID).
		withEnv("VAULT_SECRET_ID", secretID).
		run()
	assertSuccess(t, o, e, err, "bar")

	rid, _ = v.vc.Logical().Read("auth/approle2/role/testrole/role-id")
	roleID = rid.Data["role_id"].(string)
	sid, _ = v.vc.Logical().Write("auth/approle2/role/testrole/secret-id", nil)
	secretID = sid.Data["secret_id"].(string)
	o, e, err = cmd(t,
		"-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_ROLE_ID", roleID).
		withEnv("VAULT_SECRET_ID", secretID).
		withEnv("VAULT_AUTH_APPROLE_MOUNT", "approle2").
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_DynamicAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	err := v.vc.Sys().Mount("ssh/", &vaultapi.MountInput{Type: "ssh"})
	require.NoError(t, err)
	defer v.vc.Sys().Unmount("ssh")

	_, err = v.vc.Logical().Write("ssh/roles/test", map[string]any{
		"key_type": "otp", "default_user": "user", "cidr_list": "10.0.0.0/8",
	})
	require.NoError(t, err)
	testCommands := []struct {
		ds, in string
	}{
		{"vault=vault:///", `{{(ds "vault" "ssh/creds/test?ip=10.1.2.3&username=user").ip}}`},
		{"vault=vault:///ssh/creds/test", `{{(ds "vault" "?ip=10.1.2.3&username=user").ip}}`},
		{"vault=vault:///ssh/creds/test?ip=10.1.2.3&username=user", `{{(ds "vault").ip}}`},
		{"vault=vault:///?ip=10.1.2.3&username=user", `{{(ds "vault" "ssh/creds/test").ip}}`},
	}

	tok, err := v.tokenCreate("writepol", len(testCommands)*4)
	require.NoError(t, err)

	for _, tc := range testCommands {
		o, e, err := cmd(t, "-d", tc.ds, "-i", tc.in).
			withEnv("VAULT_ADDR", "http://"+v.addr).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "10.1.2.3")
	}
}

func TestDatasources_Vault_List(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/dir/foo", map[string]any{"value": "one"})
	v.vc.Logical().Write("secret/dir/bar", map[string]any{"value": "two"})
	defer v.vc.Logical().Delete("secret/dir/foo")
	defer v.vc.Logical().Delete("secret/dir/bar")
	tok, err := v.tokenCreate("listpol", 15)
	require.NoError(t, err)

	o, e, err := cmd(t,
		"-d", "vault=vault:///secret/dir/",
		"-i", `{{ range (ds "vault" ) }}{{ . }}: {{ (ds "vault" .).value }} {{end}}`).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar: two foo: one ")

	o, e, err = cmd(t,
		"-d", "vault=vault+http://"+v.addr+"/secret/",
		"-i", `{{ range (ds "vault" "dir/" ) }}{{ . }} {{end}}`).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar foo ")
}

func setupKV2Test(ctx context.Context, t *testing.T, policy string) (string, string) {
	v := setupDatasourcesVaultTest(t)

	err := v.vc.Sys().MountWithContext(ctx, "kv2", &vaultapi.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "2"},
	})
	require.NoError(t, err)

	err = v.vc.Sys().MountWithContext(ctx, "a/b/c", &vaultapi.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "2"},
	})
	require.NoError(t, err)

	s, err := v.vc.KVv2("kv2").Put(ctx, "foo", map[string]any{"first": "one"}, vaultapi.WithCheckAndSet(0))
	require.NoError(t, err)
	require.Equal(t, 1, s.VersionMetadata.Version)

	s, err = v.vc.KVv2("kv2").Put(ctx, "foo", map[string]any{"second": "two"}, vaultapi.WithCheckAndSet(1))
	require.NoError(t, err)
	require.Equal(t, 2, s.VersionMetadata.Version)

	s, err = v.vc.KVv2("a/b/c").Put(ctx, "d/e/f", map[string]any{"e": "f"})
	require.NoError(t, err)
	require.Equal(t, 1, s.VersionMetadata.Version)

	tok, err := v.tokenCreate(policy, 15)
	require.NoError(t, err)
	return v.addr, tok
}

func TestDatasources_Vault_ReadKVv2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr, tok := setupKV2Test(ctx, t, "readpol")

	t.Run("read latest version", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "vault=vault+http://"+addr+"/kv2/",
			"-i", `{{(ds "vault" "foo").second}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "two")

		o, e, err = cmd(t,
			"-c", "data=vault+http://"+addr+"/kv2/foo",
			"-i", `{{ .data.second}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "two")
	})

	t.Run("read earlier version", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "vault=vault+http://"+addr+"/kv2/",
			"-i", `{{(ds "vault" "foo?version=1").first}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "one")

		o, e, err = cmd(t,
			"-d", "vault=vault+http://"+addr+"/kv2/?version=1",
			"-i", `{{(ds "vault" "foo").first}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "one")

		o, e, err = cmd(t,
			"-c", "data=vault+http://"+addr+"/kv2/foo?version=1",
			"-i", `{{ .data.first }}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "one")
	})

	t.Run("read from mount with slashes", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "vault=vault+http://"+addr+"/a/b/c/d/",
			"-i", `{{(ds "vault" "e/f").e}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "f")

		o, e, err = cmd(t,
			"-c", "data=vault+http://"+addr+"/a/b/c/d/e/f",
			"-i", `{{ .data.e }}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "f")
	})
}

func TestDatasources_Vault_ListKVv2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr, tok := setupKV2Test(ctx, t, "listpol")

	t.Run("list latest version", func(t *testing.T) {
		o, e, err := cmd(t,
			"-c", "data=vault+http://"+addr+"/kv2/",
			"-i", `{{ range .data }}{{ . }} {{end}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "foo ")
	})

	t.Run("list from mount with slashes", func(t *testing.T) {
		o, e, err := cmd(t,
			"-c", "data=vault+http://"+addr+"/a/b/c/d/e",
			"-i", `{{ range .data }}{{ . }} {{end}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "f ")

		o, e, err = cmd(t,
			"-d", "vault=vault+http://"+addr+"/a/",
			"-i", `{{ range (ds "vault" "b/c/d/") }}{{ . }} {{end}}`).
			withEnv("VAULT_TOKEN", tok).
			run()
		assertSuccess(t, o, e, err, "e ")
	})
}
