//+build !windows

package integration

import (
	"encoding/base64"
	"strconv"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

const consulRootToken = "00000000-1111-2222-3333-444455556666"

func setupDatasourcesConsulTest(t *testing.T) (string, *vaultClient) {
	pidDir := fs.NewDir(t, "gomplate-inttests-pid")
	t.Cleanup(pidDir.Remove)

	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile(
			"consul.json",
			`{"acl_datacenter": "dc1", "acl_master_token": "`+consulRootToken+`"}`,
		),
		fs.WithFile("vault.json", `{
		"pid_file": "`+pidDir.Join("vault.pid")+`"
		}`),
	)
	t.Cleanup(tmpDir.Remove)

	port, consulAddr := freeport()
	consul := icmd.Command("consul", "agent",
		"-dev",
		"-config-file="+tmpDir.Join("consul.json"),
		"-log-level=err",
		"-http-port="+strconv.Itoa(port),
		"-pid-file="+pidDir.Join("consul.pid"),
	)
	result := icmd.StartCmd(consul)
	t.Cleanup(func() {
		err := result.Cmd.Process.Kill()
		require.NoError(t, err)

		result.Cmd.Wait()

		result.Assert(t, icmd.Expected{ExitCode: 0})
	})

	t.Logf("Fired up Consul: %v", consul)

	err := waitForURL(t, "http://"+consulAddr+"/v1/status/leader")
	require.NoError(t, err)

	_, vaultClient := startVault(t)

	return consulAddr, vaultClient
}

func consulPut(t *testing.T, consulAddr, k, v string) {
	result := icmd.RunCmd(icmd.Command("consul", "kv", "put", k, v),
		func(c *icmd.Cmd) {
			c.Env = []string{"CONSUL_HTTP_ADDR=http://" + consulAddr}
		})
	result.Assert(t, icmd.Success)
	t.Cleanup(func() {
		result := icmd.RunCmd(icmd.Command("consul", "kv", "delete", k),
			func(c *icmd.Cmd) {
				c.Env = []string{"CONSUL_HTTP_ADDR=http://" + consulAddr}
			})
		result.Assert(t, icmd.Success)
	})
}

func TestDatasources_Consul(t *testing.T) {
	consulAddr, v := setupDatasourcesConsulTest(t)
	// run as subtests since we want to maintain the same consul process for the
	// whole test
	t.Run("basic", consulBasicTest(consulAddr))
	t.Run("listKeys", consulListKeysTest(consulAddr))
	t.Run("vaultAuth", consulVaultAuthTest(consulAddr, v))
}

func consulBasicTest(consulAddr string) func(t *testing.T) {
	return func(t *testing.T) {
		consulPut(t, consulAddr, "foo1", "bar")

		o, e, err := cmd(t, "-d", "consul=consul://",
			"-i", `{{(ds "consul" "foo1")}}`).
			withEnv("CONSUL_HTTP_ADDR", "http://"+consulAddr).run()
		assertSuccess(t, o, e, err, "bar")

		consulPut(t, consulAddr, "foo2", `{"bar": "baz"}`)

		o, e, err = cmd(t, "-d", "consul=consul://?type=application/json",
			"-i", `{{(ds "consul" "foo2").bar}}`).
			withEnv("CONSUL_HTTP_ADDR", "http://"+consulAddr).run()
		assertSuccess(t, o, e, err, "baz")

		consulPut(t, consulAddr, "foo2", `bar`)

		o, e, err = cmd(t, "-d", "consul=consul://"+consulAddr,
			"-i", `{{(ds "consul" "foo2")}}`).run()
		assertSuccess(t, o, e, err, "bar")

		consulPut(t, consulAddr, "foo3", `bar`)

		o, e, err = cmd(t, "-d", "consul=consul+http://"+consulAddr,
			"-i", `{{(ds "consul" "foo3")}}`).run()
		assertSuccess(t, o, e, err, "bar")
	}
}

func consulListKeysTest(consulAddr string) func(t *testing.T) {
	return func(t *testing.T) {
		consulPut(t, consulAddr, "list-of-keys/foo1", `{"bar1": "bar1"}`)
		consulPut(t, consulAddr, "list-of-keys/foo2", "bar2")

		// Get a list of keys using the ds args
		expectedResult := `[{"key":"foo1","value":"{\"bar1\": \"bar1\"}"},{"key":"foo2","value":"bar2"}]`
		o, e, err := cmd(t, "-d", "consul=consul://",
			"-i", `{{(ds "consul" "list-of-keys/") | data.ToJSON }}`).
			withEnv("CONSUL_HTTP_ADDR", "http://"+consulAddr).run()
		assertSuccess(t, o, e, err, expectedResult)

		// Get a list of keys using the ds uri
		expectedResult = `[{"key":"foo1","value":"{\"bar1\": \"bar1\"}"},{"key":"foo2","value":"bar2"}]`
		o, e, err = cmd(t, "-d", "consul=consul+http://"+consulAddr+"/list-of-keys/",
			"-i", `{{(ds "consul" ) | data.ToJSON }}`).run()
		assertSuccess(t, o, e, err, expectedResult)

		// Get a specific value from the list of Consul keys
		expectedResult = `{"bar1": "bar1"}`
		o, e, err = cmd(t, "-d", "consul=consul+http://"+consulAddr+"/list-of-keys/",
			"-i", `{{ $data := ds "consul" }}{{ (index $data 0).value }}`).run()
		assertSuccess(t, o, e, err, expectedResult)
	}
}

func consulVaultAuthTest(consulAddr string, v *vaultClient) func(t *testing.T) {
	return func(t *testing.T) {
		err := v.vc.Sys().Mount("consul/", &vaultapi.MountInput{Type: "consul"})
		require.NoError(t, err)
		defer v.vc.Sys().Unmount("consul/")

		_, err = v.vc.Logical().Write("consul/config/access", map[string]interface{}{
			"address": consulAddr, "token": consulRootToken,
		})
		require.NoError(t, err)
		policy := base64.StdEncoding.EncodeToString([]byte(`key "" { policy = "read" }`))
		_, err = v.vc.Logical().Write("consul/roles/readonly", map[string]interface{}{"policy": policy})
		require.NoError(t, err)

		consulPut(t, consulAddr, "foo", "bar")

		o, e, err := cmd(t,
			"-d", "consul=consul://",
			"-i", `{{(ds "consul" "foo")}}`).
			withEnv("VAULT_TOKEN", vaultRootToken).
			withEnv("VAULT_ADDR", "http://"+v.addr).
			withEnv("CONSUL_VAULT_ROLE", "readonly").
			withEnv("CONSUL_HTTP_ADDR", "http://"+consulAddr).
			run()
		assertSuccess(t, o, e, err, "bar")
	}
}
