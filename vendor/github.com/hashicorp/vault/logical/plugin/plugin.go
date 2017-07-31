package plugin

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/gob"
	"fmt"

	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/vault/helper/pluginutil"
	"github.com/hashicorp/vault/logical"
)

// Register these types since we have to serialize and de-serialize tls.ConnectionState
// over the wire as part of logical.Request.Connection.
func init() {
	gob.Register(rsa.PublicKey{})
	gob.Register(ecdsa.PublicKey{})
}

// BackendPluginClient is a wrapper around backendPluginClient
// that also contains its plugin.Client instance. It's primarily
// used to cleanly kill the client on Cleanup()
type BackendPluginClient struct {
	client *plugin.Client
	sync.Mutex

	*backendPluginClient
}

// Cleanup calls the RPC client's Cleanup() func and also calls
// the go-plugin's client Kill() func
func (b *BackendPluginClient) Cleanup() {
	b.backendPluginClient.Cleanup()
	b.client.Kill()
}

// NewBackend will return an instance of an RPC-based client implementation of the backend for
// external plugins, or a concrete implementation of the backend if it is a builtin backend.
// The backend is returned as a logical.Backend interface.
func NewBackend(pluginName string, sys pluginutil.LookRunnerUtil) (logical.Backend, error) {
	// Look for plugin in the plugin catalog
	pluginRunner, err := sys.LookupPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	var backend logical.Backend
	if pluginRunner.Builtin {
		// Plugin is builtin so we can retrieve an instance of the interface
		// from the pluginRunner. Then cast it to logical.Backend.
		backendRaw, err := pluginRunner.BuiltinFactory()
		if err != nil {
			return nil, fmt.Errorf("error getting plugin type: %s", err)
		}

		var ok bool
		backend, ok = backendRaw.(logical.Backend)
		if !ok {
			return nil, fmt.Errorf("unsuported backend type: %s", pluginName)
		}

	} else {
		// create a backendPluginClient instance
		backend, err = newPluginClient(sys, pluginRunner)
		if err != nil {
			return nil, err
		}
	}

	return backend, nil
}

func newPluginClient(sys pluginutil.RunnerUtil, pluginRunner *pluginutil.PluginRunner) (logical.Backend, error) {
	// pluginMap is the map of plugins we can dispense.
	pluginMap := map[string]plugin.Plugin{
		"backend": &BackendPlugin{},
	}
	client, err := pluginRunner.Run(sys, pluginMap, handshakeConfig, []string{})
	if err != nil {
		return nil, err
	}

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("backend")
	if err != nil {
		return nil, err
	}

	// We should have a logical backend type now. This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	backendRPC := raw.(*backendPluginClient)

	return &BackendPluginClient{
		client:              client,
		backendPluginClient: backendRPC,
	}, nil
}
