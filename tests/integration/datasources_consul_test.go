//+build integration
//+build !windows

package integration

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
	vaultapi "github.com/hashicorp/vault/api"
)

type ConsulDatasourcesSuite struct {
	tmpDir       *fs.Dir
	pidDir       *fs.Dir
	consulAddr   string
	consulResult *icmd.Result
	vaultAddr    string
	vaultResult  *icmd.Result
}

var _ = Suite(&ConsulDatasourcesSuite{})

const consulRootToken = "00000000-1111-2222-3333-444455556666"

func (s *ConsulDatasourcesSuite) SetUpSuite(c *C) {
	s.pidDir = fs.NewDir(c, "gomplate-inttests-pid")
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile(
			"consul.json",
			`{"acl_datacenter": "dc1", "acl_master_token": "`+consulRootToken+`"}`,
		),
		fs.WithFile("vault.json", `{
			"pid_file": "`+s.pidDir.Join("vault.pid")+`"
			}`),
	)
	var port int
	port, s.consulAddr = freeport()
	consul := icmd.Command("consul", "agent",
		"-dev",
		"-config-file="+s.tmpDir.Join("consul.json"),
		"-log-level=err",
		"-http-port="+strconv.Itoa(port),
		"-pid-file="+s.pidDir.Join("consul.pid"),
	)
	s.consulResult = icmd.StartCmd(consul)

	c.Logf("Fired up Consul: %v", consul)

	err := waitForURL(c, "http://"+s.consulAddr+"/v1/status/leader")
	handle(c, err)

	s.startVault(c)
}

func (s *ConsulDatasourcesSuite) startVault(c *C) {
	// rename any existing token so it doesn't get overridden
	u, _ := user.Current()
	homeDir := u.HomeDir
	tokenFile := path.Join(homeDir, ".vault-token")
	info, err := os.Stat(tokenFile)
	if err == nil && info.Mode().IsRegular() {
		os.Rename(tokenFile, path.Join(homeDir, ".vault-token.bak"))
	}

	_, s.vaultAddr = freeport()
	vault := icmd.Command("vault", "server",
		"-dev",
		"-dev-root-token-id="+vaultRootToken,
		"-log-level=err",
		"-dev-listen-address="+s.vaultAddr,
		"-config="+s.tmpDir.Join("vault.json"),
	)
	s.vaultResult = icmd.StartCmd(vault)

	c.Logf("Fired up Vault: %v", vault)

	err = waitForURL(c, "http://"+s.vaultAddr+"/v1/sys/health")
	handle(c, err)
}

func killByPidFile(pidFile string) error {
	p, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(p))
	if err != nil {
		return err
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = process.Kill()
	return err
}

func (s *ConsulDatasourcesSuite) TearDownSuite(c *C) {
	defer s.tmpDir.Remove()
	defer s.pidDir.Remove()

	err := killByPidFile(s.pidDir.Join("vault.pid"))
	handle(c, err)

	err = killByPidFile(s.pidDir.Join("consul.pid"))
	handle(c, err)

	// restore old vault token if it was backed up
	u, _ := user.Current()
	homeDir := u.HomeDir
	tokenFile := path.Join(homeDir, ".vault-token.bak")
	info, err := os.Stat(tokenFile)
	if err == nil && info.Mode().IsRegular() {
		os.Rename(tokenFile, path.Join(homeDir, ".vault-token"))
	}
}

func (s *ConsulDatasourcesSuite) consulPut(c *C, k string, v string) {
	result := icmd.RunCmd(icmd.Command("consul", "kv", "put", k, v),
		func(c *icmd.Cmd) {
			c.Env = []string{"CONSUL_HTTP_ADDR=http://" + s.consulAddr}
		})
	result.Assert(c, icmd.Success)
}

func (s *ConsulDatasourcesSuite) consulDelete(c *C, k string) {
	result := icmd.RunCmd(icmd.Command("consul", "kv", "delete", k),
		func(c *icmd.Cmd) {
			c.Env = []string{"CONSUL_HTTP_ADDR=http://" + s.consulAddr}
		})
	result.Assert(c, icmd.Success)
}

func (s *ConsulDatasourcesSuite) TestConsulDatasource(c *C) {
	s.consulPut(c, "foo", "bar")
	defer s.consulDelete(c, "foo")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "consul=consul://",
		"-i", `{{(ds "consul" "foo")}}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{"CONSUL_HTTP_ADDR=http://" + s.consulAddr}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	s.consulPut(c, "foo", `{"bar": "baz"}`)
	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "consul=consul://?type=application/json",
		"-i", `{{(ds "consul" "foo").bar}}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{"CONSUL_HTTP_ADDR=http://" + s.consulAddr}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	s.consulPut(c, "foo", `bar`)
	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "consul=consul://"+s.consulAddr,
		"-i", `{{(ds "consul" "foo")}}`,
	))
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	s.consulPut(c, "foo", `bar`)
	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "consul=consul+http://"+s.consulAddr,
		"-i", `{{(ds "consul" "foo")}}`,
	))
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
}

func (s *ConsulDatasourcesSuite) TestConsulWithVaultAuth(c *C) {
	v, err := createVaultClient(s.vaultAddr, vaultRootToken)
	handle(c, err)

	err = v.vc.Sys().Mount("consul/", &vaultapi.MountInput{Type: "consul"})
	handle(c, err)
	defer v.vc.Sys().Unmount("consul/")

	_, err = v.vc.Logical().Write("consul/config/access", map[string]interface{}{
		"address": s.consulAddr, "token": consulRootToken,
	})
	handle(c, err)
	policy := base64.StdEncoding.EncodeToString([]byte(`key "" { policy = "read" }`))
	_, err = v.vc.Logical().Write("consul/roles/readonly", map[string]interface{}{"policy": policy})
	handle(c, err)

	s.consulPut(c, "foo", "bar")
	defer s.consulDelete(c, "foo")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "consul=consul://",
		"-i", `{{(ds "consul" "foo")}}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"VAULT_TOKEN=" + vaultRootToken,
			"VAULT_ADDR=http://" + s.vaultAddr,
			"CONSUL_VAULT_ROLE=readonly",
			"CONSUL_HTTP_ADDR=http://" + s.consulAddr,
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
}
