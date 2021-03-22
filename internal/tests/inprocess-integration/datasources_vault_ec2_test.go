//+build !windows

package integration

import (
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path"
	"strconv"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type VaultEc2DatasourcesSuite struct {
	tmpDir      *fs.Dir
	pidDir      *fs.Dir
	vaultAddr   string
	vaultResult *icmd.Result
	v           *vaultClient
	srv         *httptest.Server
	cert        []byte
}

var _ = Suite(&VaultEc2DatasourcesSuite{})

func (s *VaultEc2DatasourcesSuite) SetUpSuite(c *C) {
	priv, der, _ := certificateGenerate()
	s.cert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	mux := http.NewServeMux()
	mux.HandleFunc("/latest/dynamic/instance-identity/pkcs7", pkcsHandler(priv, der))
	mux.HandleFunc("/latest/dynamic/instance-identity/document", instanceDocumentHandler)
	mux.HandleFunc("/sts/", stsHandler)
	mux.HandleFunc("/ec2/", ec2Handler)

	s.srv = httptest.NewServer(mux)

	s.pidDir, s.tmpDir, s.vaultAddr, s.vaultResult = startVault(c)

	var err error
	s.v, err = createVaultClient(s.vaultAddr, vaultRootToken)
	handle(c, err)

	err = s.v.vc.Sys().PutPolicy("writepol", `path "*" {
  policy = "write"
}`)
	handle(c, err)
	err = s.v.vc.Sys().PutPolicy("readpol", `path "*" {
  policy = "read"
}`)
	handle(c, err)
}

func (s *VaultEc2DatasourcesSuite) TearDownSuite(c *C) {
	s.srv.Close()

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

func (s *VaultEc2DatasourcesSuite) TestEc2Auth(c *C) {
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	err := s.v.vc.Sys().EnableAuth("aws", "aws", "")
	handle(c, err)
	defer s.v.vc.Sys().DisableAuth("aws")
	_, err = s.v.vc.Logical().Write("auth/aws/config/client", map[string]interface{}{
		"secret_key": "secret", "access_key": "access",
		"endpoint":     s.srv.URL + "/ec2",
		"iam_endpoint": s.srv.URL + "/iam",
		"sts_endpoint": s.srv.URL + "/sts",
	})
	handle(c, err)

	_, err = s.v.vc.Logical().Write("auth/aws/config/certificate/testcert", map[string]interface{}{
		"type": "pkcs7", "aws_public_cert": string(s.cert),
	})
	handle(c, err)

	_, err = s.v.vc.Logical().Write("auth/aws/role/ami-00000000", map[string]interface{}{
		"auth_type": "ec2", "bound_ami_id": "ami-00000000",
		"policies": "readpol",
	})
	handle(c, err)

	o, e, err := cmdWithEnv(c, []string{
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	}, map[string]string{
		"HOME":              s.tmpDir.Join("home"),
		"VAULT_ADDR":        "http://" + s.v.addr,
		"AWS_META_ENDPOINT": s.srv.URL,
	})
	assertSuccess(c, o, e, err, "bar")
}
