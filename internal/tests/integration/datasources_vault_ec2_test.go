//go:build !windows
// +build !windows

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatasources_VaultEc2(t *testing.T) {
	accountID, user := "1", "Test"
	tmpDir, v, srv, cert := setupDatasourcesVaultAWSTest(t, accountID, user)

	v.vc.Logical().Write("secret/foo", map[string]any{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")

	err := v.vc.Sys().EnableAuth("aws", "aws", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("aws")

	_, err = v.vc.Logical().Write("auth/aws/config/client", map[string]any{
		"secret_key": "secret", "access_key": "access",
		"endpoint":     srv.URL + "/ec2",
		"iam_endpoint": srv.URL + "/iam",
		"sts_endpoint": srv.URL + "/sts",
		"sts_region":   "us-east-1",
	})
	require.NoError(t, err)

	_, err = v.vc.Logical().Write("auth/aws/config/certificate/testcert", map[string]any{
		"type": "pkcs7", "aws_public_cert": string(cert),
	})
	require.NoError(t, err)

	_, err = v.vc.Logical().Write("auth/aws/role/ami-00000000", map[string]any{
		"auth_type": "ec2", "bound_ami_id": "ami-00000000",
		"policies": "readpol",
	})
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("HOME", tmpDir.Join("home")).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("AWS_EC2_METADATA_SERVICE_ENDPOINT", srv.URL).
		run()
	assertSuccess(t, o, e, err, "bar")
}
