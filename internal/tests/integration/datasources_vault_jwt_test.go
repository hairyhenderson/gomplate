//+build integration
//+build !windows

package integration

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	. "gopkg.in/check.v1"

	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
	authv1 "k8s.io/api/authentication/v1"
)

type Jwt struct {
	PublicKey  string
	PrivateKey *rsa.PrivateKey
	Token      string
}

type VaultJwtDatasourcesSuite struct {
	tmpDir      *fs.Dir
	pidDir      *fs.Dir
	vaultAddr   string
	vaultResult *icmd.Result
	v           *vaultClient
	l           *net.TCPListener
	cert        []byte
	kubeAddr    string
	jwt         Jwt
}

var _ = Suite(&VaultJwtDatasourcesSuite{})

func (s *VaultJwtDatasourcesSuite) SetUpSuite(c *C) {
	var err error

	privateKey, der, _ := certificateGenerate()
	s.jwt.PrivateKey = privateKey
	s.jwt.PublicKey = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	s.jwt.Token = s.makeJwtToken(c)

	s.pidDir, s.tmpDir, s.vaultAddr, s.vaultResult = startVault(c)

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

func (s *VaultJwtDatasourcesSuite) TearDownSuite(c *C) {
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

func tokenReview(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	review := &authv1.TokenReview{}
	err := json.Unmarshal(body, review)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	review.Status.Authenticated = true
	review.Status.User.Username = "system:serviceaccount:gomplate:gomplate"
	review.Status.User.UID = "gomplate"
	review.Status.Audiences = review.Spec.Audiences
	responseBody, _ := json.Marshal(review)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBody)
	if err != nil {
		w.WriteHeader(500)
	}
}

func (s *VaultJwtDatasourcesSuite) startKubernetes(c *C) {
	var err error
	s.l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	s.kubeAddr = s.l.Addr().(*net.TCPAddr).String()
	handle(c, err)
	priv, der, _ := certificateGenerate()
	s.cert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	http.HandleFunc("/apis/authentication.k8s.io/v1/tokenreviews", tokenReview)
	server := &http.Server{
		Addr: s.kubeAddr,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			Certificates: []tls.Certificate{
				{Certificate: [][]byte{der}, PrivateKey: priv},
			},
		},
	}
	go server.ServeTLS(s.l, "", "")
}

func (s *VaultJwtDatasourcesSuite) stopKubernetes(c *C) {
	s.l.Close()
}

func (s *VaultJwtDatasourcesSuite) makeJwtToken(c *C) string {
	signingKey := jose.SigningKey{Algorithm: jose.RS256, Key: s.jwt.PrivateKey}
	sig, err := jose.NewSigner(signingKey, (&jose.SignerOptions{}).WithType("JWT"))
	handle(c, err)

	publicClaims := jwt.Claims{
		Issuer:    "gomplate",
		Subject:   "gomplate",
		Audience:  jwt.Audience{"gomplate"},
		NotBefore: jwt.NewNumericDate(time.Now()),
		Expiry:    jwt.NewNumericDate(time.Now().AddDate(10, 0, 0)),
	}
	privateClaims := map[string]interface{}{
		"groups":                                 "test",
		"kubernetes.io/serviceaccount/namespace": "gomplate",
		"kubernetes.io/serviceaccount/service-account.name": "gomplate",
		"kubernetes.io/serviceaccount/service-account.uid":  "gomplate",
	}
	raw, err := jwt.Signed(sig).Claims(publicClaims).Claims(privateClaims).CompactSerialize()
	handle(c, err)

	return raw
}

func (s *VaultJwtDatasourcesSuite) TestJwtAuth(c *C) {
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	err := s.v.vc.Sys().EnableAuth("jwt", "jwt", "")
	handle(c, err)
	defer s.v.vc.Sys().DisableAuth("jwt")

	_, err = s.v.vc.Logical().Write("auth/jwt/config", map[string]interface{}{
		"jwt_validation_pubkeys": s.jwt.PublicKey,
	})
	handle(c, err)

	_, err = s.v.vc.Logical().Write("auth/jwt/role/test", map[string]interface{}{
		"policies":        "readpol",
		"bound_subject":   "gomplate",
		"bound_audiences": "gomplate",
		"bound_claims": map[string]string{
			"groups": "test",
		},
		"user_claim":   "sub",
		"groups_claim": "groups",
		"role_type":    "jwt",
	})
	handle(c, err)

	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"HOME=" + s.tmpDir.Join("home"),
			"VAULT_ADDR=http://" + s.v.addr,
			"VAULT_AUTH_JWT_ROLE=test",
			"VAULT_AUTH_JWT_TOKEN=" + s.jwt.Token,
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
}

func (s *VaultJwtDatasourcesSuite) TestKubernetesAuth(c *C) {
	s.startKubernetes(c)
	defer s.stopKubernetes(c)
	s.v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer s.v.vc.Logical().Delete("secret/foo")
	err := s.v.vc.Sys().EnableAuth("kubernetes", "kubernetes", "")
	handle(c, err)
	defer s.v.vc.Sys().DisableAuth("kubernetes")

	_, err = s.v.vc.Logical().Write("auth/kubernetes/config", map[string]interface{}{
		"kubernetes_host":    "https://" + s.kubeAddr,
		"kubernetes_ca_cert": string(s.cert),
		"issuer":             "gomplate",
	})
	handle(c, err)

	_, err = s.v.vc.Logical().Write("auth/kubernetes/role/test", map[string]interface{}{
		"bound_service_account_names":      "*",
		"bound_service_account_namespaces": "gomplate",
		"policies":                         "readpol",
		"ttl":                              "24h",
	})
	handle(c, err)

	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"HOME=" + s.tmpDir.Join("home"),
			"VAULT_ADDR=http://" + s.v.addr,
			"VAULT_AUTH_KUBERNETES_ROLE=test",
			"VAULT_AUTH_KUBERNETES_TOKEN=" + s.jwt.Token,
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
}
