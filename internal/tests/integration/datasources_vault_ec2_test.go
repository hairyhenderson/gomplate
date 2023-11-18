//go:build !windows
// +build !windows

package integration

import (
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
)

func setupDatasourcesVaultEc2Test(t *testing.T) (*fs.Dir, *vaultClient, *httptest.Server, []byte) {
	t.Helper()

	priv, der, _ := certificateGenerate()
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	mux := http.NewServeMux()
	mux.HandleFunc("/latest/dynamic/instance-identity/pkcs7", pkcsHandler(priv, der))
	mux.HandleFunc("/latest/dynamic/instance-identity/document", instanceDocumentHandler)
	mux.HandleFunc("/latest/api/token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b []byte
		if r.Body != nil {
			var err error
			b, err = io.ReadAll(r.Body)
			require.NoError(t, err)
			defer r.Body.Close()
		}
		t.Logf("IMDS Token request: %s %s: %s", r.Method, r.URL, b)

		w.Write([]byte("testtoken"))
	}))
	mux.HandleFunc("/sts/", stsHandler)
	mux.HandleFunc("/ec2/", ec2Handler)
	mux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("unhandled request: %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	tmpDir, v := startVault(t)

	err := v.vc.Sys().PutPolicy("writepol", `path "*" {
  policy = "write"
}`)
	require.NoError(t, err)
	err = v.vc.Sys().PutPolicy("readpol", `path "*" {
  policy = "read"
}`)
	require.NoError(t, err)

	return tmpDir, v, srv, cert
}

func TestDatasources_VaultEc2(t *testing.T) {
	tmpDir, v, srv, cert := setupDatasourcesVaultEc2Test(t)

	v.vc.Logical().Write("secret/foo", map[string]interface{}{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")

	err := v.vc.Sys().EnableAuth("aws", "aws", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("aws")

	_, err = v.vc.Logical().Write("auth/aws/config/client", map[string]interface{}{
		"secret_key": "secret", "access_key": "access",
		"endpoint":     srv.URL + "/ec2",
		"iam_endpoint": srv.URL + "/iam",
		"sts_endpoint": srv.URL + "/sts",
	})
	require.NoError(t, err)

	_, err = v.vc.Logical().Write("auth/aws/config/certificate/testcert", map[string]interface{}{
		"type": "pkcs7", "aws_public_cert": string(cert),
	})
	require.NoError(t, err)

	_, err = v.vc.Logical().Write("auth/aws/role/ami-00000000", map[string]interface{}{
		"auth_type": "ec2", "bound_ami_id": "ami-00000000",
		"policies": "readpol",
	})
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("HOME", tmpDir.Join("home")).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("AWS_META_ENDPOINT", srv.URL).
		run()
	assertSuccess(t, o, e, err, "bar")
}
