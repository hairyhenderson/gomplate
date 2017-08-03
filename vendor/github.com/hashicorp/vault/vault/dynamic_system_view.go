package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/helper/consts"
	"github.com/hashicorp/vault/helper/pluginutil"
	"github.com/hashicorp/vault/helper/wrapping"
	"github.com/hashicorp/vault/logical"
)

type dynamicSystemView struct {
	core       *Core
	mountEntry *MountEntry
}

func (d dynamicSystemView) DefaultLeaseTTL() time.Duration {
	def, _ := d.fetchTTLs()
	return def
}

func (d dynamicSystemView) MaxLeaseTTL() time.Duration {
	_, max := d.fetchTTLs()
	return max
}

func (d dynamicSystemView) SudoPrivilege(path string, token string) bool {
	// Resolve the token policy
	te, err := d.core.tokenStore.Lookup(token)
	if err != nil {
		d.core.logger.Error("core: failed to lookup token", "error", err)
		return false
	}

	// Ensure the token is valid
	if te == nil {
		d.core.logger.Error("entry not found for given token")
		return false
	}

	// Construct the corresponding ACL object
	acl, err := d.core.policyStore.ACL(te.Policies...)
	if err != nil {
		d.core.logger.Error("failed to retrieve ACL for token's policies", "token_policies", te.Policies, "error", err)
		return false
	}

	// The operation type isn't important here as this is run from a path the
	// user has already been given access to; we only care about whether they
	// have sudo
	req := new(logical.Request)
	req.Operation = logical.ReadOperation
	req.Path = path
	_, rootPrivs := acl.AllowOperation(req)
	return rootPrivs
}

// TTLsByPath returns the default and max TTLs corresponding to a particular
// mount point, or the system default
func (d dynamicSystemView) fetchTTLs() (def, max time.Duration) {
	def = d.core.defaultLeaseTTL
	max = d.core.maxLeaseTTL

	if d.mountEntry.Config.DefaultLeaseTTL != 0 {
		def = d.mountEntry.Config.DefaultLeaseTTL
	}
	if d.mountEntry.Config.MaxLeaseTTL != 0 {
		max = d.mountEntry.Config.MaxLeaseTTL
	}

	return
}

// Tainted indicates that the mount is in the process of being removed
func (d dynamicSystemView) Tainted() bool {
	return d.mountEntry.Tainted
}

// CachingDisabled indicates whether to use caching behavior
func (d dynamicSystemView) CachingDisabled() bool {
	return d.core.cachingDisabled || (d.mountEntry != nil && d.mountEntry.Config.ForceNoCache)
}

// Checks if this is a primary Vault instance.
func (d dynamicSystemView) ReplicationState() consts.ReplicationState {
	var state consts.ReplicationState
	d.core.clusterParamsLock.RLock()
	state = d.core.replicationState
	d.core.clusterParamsLock.RUnlock()
	return state
}

// ResponseWrapData wraps the given data in a cubbyhole and returns the
// token used to unwrap.
func (d dynamicSystemView) ResponseWrapData(data map[string]interface{}, ttl time.Duration, jwt bool) (*wrapping.ResponseWrapInfo, error) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "sys/wrapping/wrap",
	}

	resp := &logical.Response{
		WrapInfo: &wrapping.ResponseWrapInfo{
			TTL: ttl,
		},
		Data: data,
	}

	if jwt {
		resp.WrapInfo.Format = "jwt"
	}

	_, err := d.core.wrapInCubbyhole(req, resp)
	if err != nil {
		return nil, err
	}

	return resp.WrapInfo, nil
}

// LookupPlugin looks for a plugin with the given name in the plugin catalog. It
// returns a PluginRunner or an error if no plugin was found.
func (d dynamicSystemView) LookupPlugin(name string) (*pluginutil.PluginRunner, error) {
	r, err := d.core.pluginCatalog.Get(name)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("no plugin found with name: %s", name)
	}

	return r, nil
}

// MlockEnabled returns the configuration setting for enabling mlock on plugins.
func (d dynamicSystemView) MlockEnabled() bool {
	return d.core.enableMlock
}
