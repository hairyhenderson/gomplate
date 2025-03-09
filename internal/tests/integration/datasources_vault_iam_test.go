//go:build !windows
// +build !windows

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatasources_VaultIAM(t *testing.T) {
	accountID := "000000000000"
	user := "foo"

	tmpDir, v, srv, _ := setupDatasourcesVaultAWSTest(t, accountID, user)

	v.vc.Logical().Write("secret/foo", map[string]any{"value": "bar"})
	defer v.vc.Logical().Delete("secret/foo")

	err := v.vc.Sys().EnableAuth("aws", "aws", "")
	require.NoError(t, err)
	defer v.vc.Sys().DisableAuth("aws")

	endpoint := srv.URL

	accessKeyID := "secret"
	secretAccessKey := "access"

	_, err = v.vc.Logical().Write("auth/aws/config/client", map[string]any{
		"access_key":   accessKeyID,
		"secret_key":   secretAccessKey,
		"endpoint":     endpoint,
		"iam_endpoint": endpoint + "/iam",
		"sts_endpoint": endpoint + "/sts",
		"sts_region":   "us-east-1",
	})
	require.NoError(t, err)

	_, err = v.vc.Logical().Write("auth/aws/role/foo", map[string]any{
		"auth_type":               "iam",
		"bound_iam_principal_arn": "arn:aws:iam::" + accountID + ":*",
		"policies":                "readpol",
		"max_ttl":                 "5m",
	})
	require.NoError(t, err)

	o, e, err := cmd(t, "-d", "vault=vault:///secret/",
		"-i", `{{(ds "vault" "foo").value}}`).
		withEnv("HOME", tmpDir.Join("home")).
		withEnv("VAULT_ADDR", "http://"+v.addr).
		withEnv("AWS_ACCESS_KEY_ID", accessKeyID).
		withEnv("AWS_SECRET_ACCESS_KEY", secretAccessKey).
		run()
	assertSuccess(t, o, e, err, "bar")
}
