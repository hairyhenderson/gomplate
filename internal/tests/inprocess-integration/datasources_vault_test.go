//+build !windows

package integration

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"

	vaultapi "github.com/hashicorp/vault/api"
	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type VaultDatasourcesSuite struct {
	tmpDir      *fs.Dir
	pidDir      *fs.Dir
	vaultAddr   string
	vaultResult *icmd.Result
	v           *vaultClient
}

var _ = Suite(&VaultDatasourcesSuite{})

const vaultRootToken = "00000000-1111-2222-3333-444455556666"

func (s *VaultDatasourcesSuite) SetUpSuite(c *C) {
	s.pidDir, s.tmpDir, s.vaultAddr, s.vaultResult = startVault(c)

	var err error
	s.v, err = createVaultClient(s.vaultAddr, vaultRootToken)
	handle(c, err)

	err = s.v.vc.Sys().PutPolicy("writepol", `path "*" {
  capabilities = ["create","update","delete"]
}`)
	handle(c, err)
	err = s.v.vc.Sys().PutPolicy("readpol", `path "*" {
  capabilities = ["read","delete"]
}`)
	handle(c, err)
	err = s.v.vc.Sys().PutPolicy("listPol", `path "*" {
  capabilities = ["read","list","delete"]
}`)
	handle(c, err)
}

func startVault(c *C) (pidDir, tmpDir *fs.Dir, vaultAddr string, vaultResult *icmd.Result) {
	pidDir = fs.NewDir(c, "gomplate-inttests-vaultpid")
	tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("config.json", `{
		"pid_file": "`+pidDir.Join("vault.pid")+`"
		}`),
	)

	// rename any existing token so it doesn't get overridden
	u, _ := user.Current()
	homeDir := u.HomeDir
	tokenFile := path.Join(homeDir, ".vault-token")
	info, err := os.Stat(tokenFile)
	if err == nil && info.Mode().IsRegular() {
		os.Rename(tokenFile, path.Join(homeDir, ".vault-token.bak"))
	}

	_, vaultAddr = freeport()
	vault := icmd.Command("vault", "server",
		"-dev",
		"-dev-root-token-id="+vaultRootToken,
		"-dev-leased-kv",
		"-log-level=err",
		"-dev-listen-address="+vaultAddr,
		"-config="+tmpDir.Join("config.json"),
	)
	vaultResult = icmd.StartCmd(vault)

	c.Logf("Fired up Vault: %v", vault)

	err = waitForURL(c, "http://"+vaultAddr+"/v1/sys/health")
	handle(c, err)

	return pidDir, tmpDir, vaultAddr, vaultResult
}

func (s *VaultDatasourcesSuite) TearDownSuite(c *C) {
	defer s.tmpDir.Remove()
	defer s.pidDir.Remove()

	p, err := ioutil.ReadFile(s.pidDir.Join("vault.pid"))
	handle(c, err)
	pid, err := strconv.Atoi(string(p))
	handle(c, err)
	process, err := os.FindProcess(pid)
	handle(c, err)
	err = process.Kill()
	handle(c, err)

	// restore old token if it was backed up
	u, _ := user.Current()
	homeDir := u.HomeDir
	tokenFile := path.Join(homeDir, ".vault-token.bak")
	info, err := os.Stat(tokenFile)
	if err == nil && info.Mode().IsRegular() {
		os.Rename(tokenFile, path.Join(homeDir, ".vault-token"))
	}
}

func (s *VaultDatasourcesSuite) TestTokenAuth(c *C) {
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	tok, err := s.v.tokenCreate("readpol", 5)
	handle(c, err)

	o, e, err := cmdWithEnv(c, []string{"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`},
		map[string]string{
			"VAULT_ADDR":  "http://" + s.v.addr,
			"VAULT_TOKEN": tok,
		})
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdWithEnv(c, []string{"-d", "vault=vault+http://" + s.v.addr + "/secret",
		"-i", `{{(ds "vault" "foo").value}}`},
		map[string]string{
			"VAULT_TOKEN": tok,
		})
	assertSuccess(c, o, e, err, "bar")

	_, _, err = cmdWithEnv(c, []string{"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "bar").value}}`},
		map[string]string{
			"VAULT_ADDR":  "http://" + s.v.addr,
			"VAULT_TOKEN": tok,
		})
	assert.ErrorContains(c, err, "error calling ds: Couldn't read datasource 'vault': no value found for path /secret/bar")

	tokFile := fs.NewFile(c, "test-vault-token", fs.WithContent(tok))
	defer tokFile.Remove()

	o, e, err = cmdWithEnv(c, []string{"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`},
		map[string]string{
			"VAULT_ADDR":       "http://" + s.v.addr,
			"VAULT_TOKEN_FILE": tokFile.Path(),
		})
	assertSuccess(c, o, e, err, "bar")
}

func (s *VaultDatasourcesSuite) TestUserPassAuth(c *C) {
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	err := s.v.vc.Sys().EnableAuth("userpass", "userpass", "")
	handle(c, err)
	err = s.v.vc.Sys().EnableAuth("userpass2", "userpass", "")
	handle(c, err)
	defer s.v.vc.Sys().DisableAuth("userpass")
	defer s.v.vc.Sys().DisableAuth("userpass2")
	_, err = s.v.vc.Logical().Write("auth/userpass/users/dave", map[string]interface{}{
		"password": "foo", "ttl": "10s", "policies": "readpol"})
	handle(c, err)
	_, err = s.v.vc.Logical().Write("auth/userpass2/users/dave", map[string]interface{}{
		"password": "bar", "ttl": "10s", "policies": "readpol"})
	handle(c, err)

	o, e, err := cmdWithEnv(c, []string{"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`},
		map[string]string{
			"VAULT_ADDR":          "http://" + s.v.addr,
			"VAULT_AUTH_USERNAME": "dave",
			"VAULT_AUTH_PASSWORD": "foo",
		})
	assertSuccess(c, o, e, err, "bar")

	userFile := fs.NewFile(c, "test-vault-user", fs.WithContent("dave"))
	passFile := fs.NewFile(c, "test-vault-pass", fs.WithContent("foo"))
	defer userFile.Remove()
	defer passFile.Remove()
	o, e, err = cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"VAULT_ADDR":               "http://" + s.v.addr,
		"VAULT_AUTH_USERNAME_FILE": userFile.Path(),
		"VAULT_AUTH_PASSWORD_FILE": passFile.Path(),
	})
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"VAULT_ADDR":                "http://" + s.v.addr,
		"VAULT_AUTH_USERNAME":       "dave",
		"VAULT_AUTH_PASSWORD":       "bar",
		"VAULT_AUTH_USERPASS_MOUNT": "userpass2",
	})
	assertSuccess(c, o, e, err, "bar")
}

func (s *VaultDatasourcesSuite) TestAppRoleAuth(c *C) {
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	err := s.v.vc.Sys().EnableAuth("approle", "approle", "")
	handle(c, err)
	err = s.v.vc.Sys().EnableAuth("approle2", "approle", "")
	handle(c, err)
	defer s.v.vc.Sys().DisableAuth("approle")
	defer s.v.vc.Sys().DisableAuth("approle2")
	_, err = s.v.vc.Logical().Write("auth/approle/role/testrole", map[string]interface{}{
		"secret_id_ttl": "10s", "token_ttl": "20s",
		"secret_id_num_uses": "1", "policies": "readpol",
	})
	handle(c, err)
	_, err = s.v.vc.Logical().Write("auth/approle2/role/testrole", map[string]interface{}{
		"secret_id_ttl": "10s", "token_ttl": "20s",
		"secret_id_num_uses": "1", "policies": "readpol",
	})
	handle(c, err)

	rid, _ := s.v.vc.Logical().Read("auth/approle/role/testrole/role-id")
	roleID := rid.Data["role_id"].(string)
	sid, _ := s.v.vc.Logical().Write("auth/approle/role/testrole/secret-id", nil)
	secretID := sid.Data["secret_id"].(string)
	o, e, err := cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"VAULT_ADDR":      "http://" + s.v.addr,
		"VAULT_ROLE_ID":   roleID,
		"VAULT_SECRET_ID": secretID,
	})
	assertSuccess(c, o, e, err, "bar")

	rid, _ = s.v.vc.Logical().Read("auth/approle2/role/testrole/role-id")
	roleID = rid.Data["role_id"].(string)
	sid, _ = s.v.vc.Logical().Write("auth/approle2/role/testrole/secret-id", nil)
	secretID = sid.Data["secret_id"].(string)
	o, e, err = cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"VAULT_ADDR":               "http://" + s.v.addr,
		"VAULT_ROLE_ID":            roleID,
		"VAULT_SECRET_ID":          secretID,
		"VAULT_AUTH_APPROLE_MOUNT": "approle2",
	})
	assertSuccess(c, o, e, err, "bar")
}

func (s *VaultDatasourcesSuite) TestAppIDAuth(c *C) {
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	err := s.v.vc.Sys().EnableAuth("app-id", "app-id", "")
	handle(c, err)
	err = s.v.vc.Sys().EnableAuth("app-id2", "app-id", "")
	handle(c, err)
	defer s.v.vc.Sys().DisableAuth("app-id")
	defer s.v.vc.Sys().DisableAuth("app-id2")
	_, err = s.v.vc.Logical().Write("auth/app-id/map/app-id/testappid", map[string]interface{}{
		"display_name": "test_app_id", "value": "readpol",
	})
	handle(c, err)
	_, err = s.v.vc.Logical().Write("auth/app-id/map/user-id/testuserid", map[string]interface{}{
		"value": "testappid",
	})
	handle(c, err)
	_, err = s.v.vc.Logical().Write("auth/app-id2/map/app-id/testappid", map[string]interface{}{
		"display_name": "test_app_id", "value": "readpol",
	})
	handle(c, err)
	_, err = s.v.vc.Logical().Write("auth/app-id2/map/user-id/testuserid", map[string]interface{}{
		"value": "testappid",
	})
	handle(c, err)

	o, e, err := cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"VAULT_ADDR":    "http://" + s.v.addr,
		"VAULT_APP_ID":  "testappid",
		"VAULT_USER_ID": "testuserid",
	})
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"VAULT_ADDR":              "http://" + s.v.addr,
		"VAULT_APP_ID":            "testappid",
		"VAULT_USER_ID":           "testuserid",
		"VAULT_AUTH_APP_ID_MOUNT": "app-id2",
	})
	assertSuccess(c, o, e, err, "bar")
}

func (s *VaultDatasourcesSuite) TestDynamicAuth(c *C) {
	err := s.v.vc.Sys().Mount("ssh/", &vaultapi.MountInput{Type: "ssh"})
	handle(c, err)
	defer s.v.vc.Sys().Unmount("ssh")

	_, err = s.v.vc.Logical().Write("ssh/roles/test", map[string]interface{}{
		"key_type": "otp", "default_user": "user", "cidr_list": "10.0.0.0/8",
	})
	handle(c, err)
	testCommands := []struct {
		ds, in string
	}{
		{"vault=vault:///", `{{(ds "vault" "ssh/creds/test?ip=10.1.2.3&username=user").ip}}`},
		{"vault=vault:///ssh/creds/test", `{{(ds "vault" "?ip=10.1.2.3&username=user").ip}}`},
		{"vault=vault:///ssh/creds/test?ip=10.1.2.3&username=user", `{{(ds "vault").ip}}`},
		{"vault=vault:///?ip=10.1.2.3&username=user", `{{(ds "vault" "ssh/creds/test").ip}}`},
	}
	tok, err := s.v.tokenCreate("writepol", len(testCommands)*2)
	handle(c, err)

	for _, v := range testCommands {
		args := []string{"-d", v.ds, "-i", v.in}
		o, e, err := cmdWithEnv(c, args, map[string]string{
			"VAULT_ADDR":  "http://" + s.v.addr,
			"VAULT_TOKEN": tok,
		})
		assertSuccess(c, o, e, err, "10.1.2.3")
	}
}

func (s *VaultDatasourcesSuite) TestList(c *C) {
	s.v.vc.Logical().Write("secret/dir/foo", map[string]interface{}{"value": "one"})
	s.v.vc.Logical().Write("secret/dir/bar", map[string]interface{}{"value": "two"})
	defer s.v.vc.Logical().Delete("secret/dir/foo")
	defer s.v.vc.Logical().Delete("secret/dir/bar")
	tok, err := s.v.tokenCreate("listpol", 5)
	handle(c, err)

	o, e, err := cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret/dir/",
		"-i", `{{ range (ds "vault" ) }}{{ . }}: {{ (ds "vault" .).value }} {{end}}`,
	}, map[string]string{
		"VAULT_ADDR":  "http://" + s.v.addr,
		"VAULT_TOKEN": tok,
	})
	assertSuccess(c, o, e, err, "bar: two foo: one ")

	o, e, err = cmdWithEnv(c, []string{
		"-d", "vault=vault+http://" + s.v.addr + "/secret",
		"-i", `{{ range (ds "vault" "dir/" ) }}{{ . }} {{end}}`,
	}, map[string]string{
		"VAULT_TOKEN": tok,
	})
	assertSuccess(c, o, e, err, "bar foo ")
}
