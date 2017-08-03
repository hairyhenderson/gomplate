package vault

import (
	"sort"

	"github.com/hashicorp/vault/logical"
)

// Capabilities is used to fetch the capabilities of the given token on the given path
func (c *Core) Capabilities(token, path string) ([]string, error) {
	if path == "" {
		return nil, &logical.StatusBadRequest{Err: "missing path"}
	}

	if token == "" {
		return nil, &logical.StatusBadRequest{Err: "missing token"}
	}

	te, err := c.tokenStore.Lookup(token)
	if err != nil {
		return nil, err
	}
	if te == nil {
		return nil, &logical.StatusBadRequest{Err: "invalid token"}
	}

	if te.Policies == nil {
		return []string{DenyCapability}, nil
	}

	var policies []*Policy
	for _, tePolicy := range te.Policies {
		policy, err := c.policyStore.GetPolicy(tePolicy)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	if len(policies) == 0 {
		return []string{DenyCapability}, nil
	}

	acl, err := NewACL(policies)
	if err != nil {
		return nil, err
	}

	capabilities := acl.Capabilities(path)
	sort.Strings(capabilities)
	return capabilities, nil
}
