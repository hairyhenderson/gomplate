//+build !windows

package integration

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

const vaultRootToken = "00000000-1111-2222-3333-444455556666"

func setupDatasourcesVaultTest(t *testing.T) *vaultClient {
	_, vaultClient := startVault(t)

	err := vaultClient.vc.Sys().PutPolicy("writepol", `path "*" {
  capabilities = ["create","update","delete"]
}`)
	require.NoError(t, err)
	err = vaultClient.vc.Sys().PutPolicy("readpol", `path "*" {
  capabilities = ["read","delete"]
}`)
	require.NoError(t, err)
	err = vaultClient.vc.Sys().PutPolicy("listPol", `path "*" {
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

	_, vaultAddr := freeport()
	vault := icmd.Command("vault", "server",
		"-dev",
		"-dev-root-token-id="+vaultRootToken,
		"-dev-leased-kv",
		"-log-level=err",
		"-dev-listen-address="+vaultAddr,
		"-config="+tmpDir.Join("config.json"),
	)
	result := icmd.StartCmd(vault)

	t.Logf("Fired up Vault: %v", vault)

	err = waitForURL(t, "http://"+vaultAddr+"/v1/sys/health")
	require.NoError(t, err)

	config := vaultapi.DefaultConfig()
	config.Address = "http://" + vaultAddr
	client, err := vaultapi.NewClient(config)
	require.NoError(t, err)

	client.SetToken(vaultRootToken)
	vaultClient := &vaultClient{client}
	assert.Equal(t, "http://"+vaultAddr, client.Address())

	t.Cleanup(func() {
		err := result.Cmd.Process.Kill()
		require.NoError(t, err)

		err = result.Cmd.Wait()
		require.Error(t, err)
		var exerr *exec.ExitError
		if errors.As(err, &exerr) {
			require.Equal(t, "signal: killed", exerr.Error())
			err = nil
		}
		require.NoError(t, err)
		// should exit with -1 since we killed it!
		require.Equal(t, -1, result.Cmd.ProcessState.ExitCode())

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

type vaultClient struct {
	vc *vaultapi.Client
}

func (v *vaultClient) tokenCreate(policy string, uses int) (string, error) {
	opts := &vaultapi.TokenCreateRequest{
		Policies: []string{policy},
		TTL:      "1m",
		NumUses:  uses,
	}
	token, err := v.vc.Auth().Token().Create(opts)
	if err != nil {
		return "", err
	}
	return token.Auth.ClientToken, nil
}

func TestDatasources_Vault_TokenAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	tok, err := v.tokenCreate("readpol", 5)
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t, "-d", "vault=vault+"+v.vc.Address()+"/secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar")

	_, _, err = cmd(t, "-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "bar")}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_TOKEN", tok).
		run()
	assert.ErrorContains(t, err, "error calling ds: failed to request vault:///secret/bar: no value found for path /secret/bar")

	tokFile := fs.NewFile(t, "test-vault-token", fs.WithContent(tok))
	defer tokFile.Remove()

	o, e, err = cmd(t, "-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_TOKEN_FILE", tokFile.Path()).
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_UserPassAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	err := v.vc.Sys().EnableAuth("userpass", "userpass", "")
	require.NoError(t, err)
	err = v.vc.Sys().EnableAuth("userpass2", "userpass", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("userpass")
	defer v.vc.Sys().DisableAuth("userpass2")
	_, err = v.vc.Logical().Write("auth/userpass/users/dave", map[string]interface{}{
		"password": "foo", "ttl": "10s", "policies": "readpol"})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/userpass2/users/dave", map[string]interface{}{
		"password": "bar", "ttl": "10s", "policies": "readpol"})
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_AUTH_USERNAME", "dave").
		withEnv("VAULT_AUTH_PASSWORD", "foo").
		run()
	assertSuccess(t, o, e, err, "bar")

	userFile := fs.NewFile(t, "test-vault-user", fs.WithContent("dave"))
	passFile := fs.NewFile(t, "test-vault-pass", fs.WithContent("foo"))
	defer userFile.Remove()
	defer passFile.Remove()
	o, e, err = cmd(t,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_AUTH_USERNAME_FILE", userFile.Path()).
		withEnv("VAULT_AUTH_PASSWORD_FILE", passFile.Path()).
		run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_AUTH_USERNAME", "dave").
		withEnv("VAULT_AUTH_PASSWORD", "bar").
		withEnv("VAULT_AUTH_USERPASS_MOUNT", "userpass2").
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_AppRoleAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	err := v.vc.Sys().EnableAuth("approle", "approle", "")
	require.NoError(t, err)
	err = v.vc.Sys().EnableAuth("approle2", "approle", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("approle")
	defer v.vc.Sys().DisableAuth("approle2")
	_, err = v.vc.Logical().Write("auth/approle/role/testrole", map[string]interface{}{
		"secret_id_ttl": "10s", "token_ttl": "20s",
		"secret_id_num_uses": "1", "policies": "readpol",
	})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/approle2/role/testrole", map[string]interface{}{
		"secret_id_ttl": "10s", "token_ttl": "20s",
		"secret_id_num_uses": "1", "policies": "readpol",
	})
	require.NoError(t, err)

	rid, _ := v.vc.Logical().Read("auth/approle/role/testrole/role-id")
	roleID := rid.Data["role_id"].(string)
	sid, _ := v.vc.Logical().Write("auth/approle/role/testrole/secret-id", nil)
	secretID := sid.Data["secret_id"].(string)
	o, e, err := cmd(t,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_ROLE_ID", roleID).
		withEnv("VAULT_SECRET_ID", secretID).
		run()
	assertSuccess(t, o, e, err, "bar")

	rid, _ = v.vc.Logical().Read("auth/approle2/role/testrole/role-id")
	roleID = rid.Data["role_id"].(string)
	sid, _ = v.vc.Logical().Write("auth/approle2/role/testrole/secret-id", nil)
	secretID = sid.Data["secret_id"].(string)
	o, e, err = cmd(t,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_ROLE_ID", roleID).
		withEnv("VAULT_SECRET_ID", secretID).
		withEnv("VAULT_AUTH_APPROLE_MOUNT", "approle2").
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_AppIDAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")
	err := v.vc.Sys().EnableAuth("app-id", "app-id", "")
	require.NoError(t, err)
	err = v.vc.Sys().EnableAuth("app-id2", "app-id", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("app-id")
	defer v.vc.Sys().DisableAuth("app-id2")
	_, err = v.vc.Logical().Write("auth/app-id/map/app-id/testappid", map[string]interface{}{
		"display_name": "test_app_id", "value": "readpol",
	})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/app-id/map/user-id/testuserid", map[string]interface{}{
		"value": "testappid",
	})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/app-id2/map/app-id/testappid", map[string]interface{}{
		"display_name": "test_app_id", "value": "readpol",
	})
	require.NoError(t, err)
	_, err = v.vc.Logical().Write("auth/app-id2/map/user-id/testuserid", map[string]interface{}{
		"value": "testappid",
	})
	require.NoError(t, err)

	o, e, err := cmd(t,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_APP_ID", "testappid").
		withEnv("VAULT_USER_ID", "testuserid").
		run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_APP_ID", "testappid").
		withEnv("VAULT_USER_ID", "testuserid").
		withEnv("VAULT_AUTH_APP_ID_MOUNT", "app-id2").
		run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_Vault_DynamicAuth(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	err := v.vc.Sys().Mount("ssh/", &vaultapi.MountInput{Type: "ssh"})
	require.NoError(t, err)

	t.Cleanup(func() { v.vc.Sys().Unmount("ssh") })

	_, err = v.vc.Logical().Write("ssh/roles/test", map[string]interface{}{
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
	tok, err := v.tokenCreate("writepol", len(testCommands)*2)
	require.NoError(t, err)

	for _, tc := range testCommands {
		tc := tc
		t.Run(tc.ds, func(t *testing.T) {
			o, e, err := cmd(t, "-d", tc.ds, "-i", tc.in).
				withEnv("VAULT_ADDR", v.vc.Address()).
				withEnv("VAULT_TOKEN", tok).
				run()
			assertSuccess(t, o, e, err, "10.1.2.3")
		})
	}
}

func TestDatasources_Vault_List(t *testing.T) {
	v := setupDatasourcesVaultTest(t)

	v.vc.Logical().Write("secret/dir/foo", map[string]interface{}{"value": "one"})
	v.vc.Logical().Write("secret/dir/bar", map[string]interface{}{"value": "two"})
	defer v.vc.Logical().Delete("secret/dir/foo")
	defer v.vc.Logical().Delete("secret/dir/bar")
	tok, err := v.tokenCreate("listpol", 5)
	require.NoError(t, err)

	o, e, err := cmd(t,
		"-d", "vault=vault:///secret/dir/",
		"-i", `{{ range (ds "vault" ) }}{{ . }}: {{ (ds "vault" .).value }} {{end}}`).
		withEnv("VAULT_ADDR", v.vc.Address()).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar: two foo: one ")

	o, e, err = cmd(t,
		"-d", "vault=vault+"+v.vc.Address()+"/secret",
		"-i", `{{ range (ds "vault" "dir/" ) }}{{ . }} {{end}}`).
		withEnv("VAULT_TOKEN", tok).
		run()
	assertSuccess(t, o, e, err, "bar foo ")
}
