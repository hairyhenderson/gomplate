package okta

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/vault/helper/logformat"
	"github.com/hashicorp/vault/helper/policyutil"
	log "github.com/mgutz/logxi/v1"

	"github.com/hashicorp/vault/logical"
	logicaltest "github.com/hashicorp/vault/logical/testing"
)

func TestBackend_Config(t *testing.T) {
	b, err := Factory(&logical.BackendConfig{
		Logger: logformat.NewVaultLogger(log.LevelTrace),
		System: &logical.StaticSystemView{},
	})
	if err != nil {
		t.Fatalf("Unable to create backend: %s", err)
	}

	username := os.Getenv("OKTA_USERNAME")
	password := os.Getenv("OKTA_PASSWORD")

	configData := map[string]interface{}{
		"organization": os.Getenv("OKTA_ORG"),
		"base_url":     "oktapreview.com",
	}

	configDataToken := map[string]interface{}{
		"token": os.Getenv("OKTA_API_TOKEN"),
	}

	logicaltest.Test(t, logicaltest.TestCase{
		AcceptanceTest: true,
		PreCheck:       func() { testAccPreCheck(t) },
		Backend:        b,
		Steps: []logicaltest.TestStep{
			testConfigCreate(t, configData),
			testLoginWrite(t, username, "wrong", "E0000004", nil),
			testLoginWrite(t, username, password, "user is not a member of any authorized policy", nil),
			testAccUserGroups(t, username, "local_group,local_group2"),
			testAccGroups(t, "local_group", "local_group_policy"),
			testLoginWrite(t, username, password, "", []string{"local_group_policy"}),
			testAccGroups(t, "Everyone", "everyone_group_policy,every_group_policy2"),
			testLoginWrite(t, username, password, "", []string{"local_group_policy"}),
			testConfigUpdate(t, configDataToken),
			testConfigRead(t, configData),
			testLoginWrite(t, username, password, "", []string{"everyone_group_policy", "every_group_policy2", "local_group_policy"}),
			testAccGroups(t, "local_group2", "testgroup_group_policy"),
			testLoginWrite(t, username, password, "", []string{"everyone_group_policy", "every_group_policy2", "local_group_policy", "testgroup_group_policy"}),
		},
	})
}

func testLoginWrite(t *testing.T, username, password, reason string, policies []string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "login/" + username,
		ErrorOk:   true,
		Data: map[string]interface{}{
			"password": password,
		},
		Check: func(resp *logical.Response) error {
			if resp.IsError() {
				if reason == "" || !strings.Contains(resp.Error().Error(), reason) {
					return resp.Error()
				}
			}

			if resp.Auth != nil {
				if !policyutil.EquivalentPolicies(resp.Auth.Policies, policies) {
					return fmt.Errorf("policy mismatch expected %v but got %v", policies, resp.Auth.Policies)
				}
			}

			return nil
		},
	}
}

func testConfigCreate(t *testing.T, d map[string]interface{}) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.CreateOperation,
		Path:      "config",
		Data:      d,
	}
}

func testConfigUpdate(t *testing.T, d map[string]interface{}) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      d,
	}
}

func testConfigRead(t *testing.T, d map[string]interface{}) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.ReadOperation,
		Path:      "config",
		Check: func(resp *logical.Response) error {
			if resp.IsError() {
				return resp.Error()
			}

			if resp.Data["Org"] != d["organization"] {
				return fmt.Errorf("Org mismatch expected %s but got %s", d["organization"], resp.Data["Org"])
			}

			if resp.Data["BaseURL"] != d["base_url"] {
				return fmt.Errorf("BaseURL mismatch expected %s but got %s", d["base_url"], resp.Data["BaseURL"])
			}

			if _, exists := resp.Data["Token"]; exists {
				return fmt.Errorf("token should not be returned on a read request")
			}

			return nil
		},
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("OKTA_USERNAME"); v == "" {
		t.Fatal("OKTA_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("OKTA_PASSWORD"); v == "" {
		t.Fatal("OKTA_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("OKTA_ORG"); v == "" {
		t.Fatal("OKTA_ORG must be set for acceptance tests")
	}
}

func testAccUserGroups(t *testing.T, user string, groups string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "users/" + user,
		Data: map[string]interface{}{
			"groups": groups,
		},
	}
}

func testAccGroups(t *testing.T, group string, policies string) logicaltest.TestStep {
	t.Logf("[testAccGroups] - Registering group %s, policy %s", group, policies)
	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "groups/" + group,
		Data: map[string]interface{}{
			"policies": policies,
		},
	}
}

func testAccLogin(t *testing.T, user, password string, keys []string) logicaltest.TestStep {
	return logicaltest.TestStep{
		Operation: logical.UpdateOperation,
		Path:      "login/" + user,
		Data: map[string]interface{}{
			"password": password,
		},
		Unauthenticated: true,

		Check: logicaltest.TestCheckAuth(keys),
	}
}
