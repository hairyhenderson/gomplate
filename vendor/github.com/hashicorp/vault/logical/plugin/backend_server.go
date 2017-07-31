package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/vault/logical"
)

// backendPluginServer is the RPC server that backendPluginClient talks to,
// it methods conforming to requirements by net/rpc
type backendPluginServer struct {
	broker  *plugin.MuxBroker
	backend logical.Backend
	factory func(*logical.BackendConfig) (logical.Backend, error)

	loggerClient  *rpc.Client
	sysViewClient *rpc.Client
	storageClient *rpc.Client
}

func (b *backendPluginServer) HandleRequest(args *HandleRequestArgs, reply *HandleRequestReply) error {
	storage := &StorageClient{client: b.storageClient}
	args.Request.Storage = storage

	resp, err := b.backend.HandleRequest(args.Request)
	*reply = HandleRequestReply{
		Response: resp,
		Error:    plugin.NewBasicError(err),
	}

	return nil
}

func (b *backendPluginServer) SpecialPaths(_ interface{}, reply *SpecialPathsReply) error {
	*reply = SpecialPathsReply{
		Paths: b.backend.SpecialPaths(),
	}
	return nil
}

func (b *backendPluginServer) HandleExistenceCheck(args *HandleExistenceCheckArgs, reply *HandleExistenceCheckReply) error {
	storage := &StorageClient{client: b.storageClient}
	args.Request.Storage = storage

	checkFound, exists, err := b.backend.HandleExistenceCheck(args.Request)
	*reply = HandleExistenceCheckReply{
		CheckFound: checkFound,
		Exists:     exists,
		Error:      plugin.NewBasicError(err),
	}

	return nil
}

func (b *backendPluginServer) Cleanup(_ interface{}, _ *struct{}) error {
	b.backend.Cleanup()

	// Close rpc clients
	b.loggerClient.Close()
	b.sysViewClient.Close()
	b.storageClient.Close()
	return nil
}

func (b *backendPluginServer) Initialize(_ interface{}, _ *struct{}) error {
	err := b.backend.Initialize()
	return err
}

func (b *backendPluginServer) InvalidateKey(args string, _ *struct{}) error {
	b.backend.InvalidateKey(args)
	return nil
}

// Setup dials into the plugin's broker to get a shimmed storage, logger, and
// system view of the backend. This method also instantiates the underlying
// backend through its factory func for the server side of the plugin.
func (b *backendPluginServer) Setup(args *SetupArgs, reply *SetupReply) error {
	// Dial for storage
	storageConn, err := b.broker.Dial(args.StorageID)
	if err != nil {
		*reply = SetupReply{
			Error: plugin.NewBasicError(err),
		}
		return nil
	}
	rawStorageClient := rpc.NewClient(storageConn)
	b.storageClient = rawStorageClient

	storage := &StorageClient{client: rawStorageClient}

	// Dial for logger
	loggerConn, err := b.broker.Dial(args.LoggerID)
	if err != nil {
		*reply = SetupReply{
			Error: plugin.NewBasicError(err),
		}
		return nil
	}
	rawLoggerClient := rpc.NewClient(loggerConn)
	b.loggerClient = rawLoggerClient

	logger := &LoggerClient{client: rawLoggerClient}

	// Dial for sys view
	sysViewConn, err := b.broker.Dial(args.SysViewID)
	if err != nil {
		*reply = SetupReply{
			Error: plugin.NewBasicError(err),
		}
		return nil
	}
	rawSysViewClient := rpc.NewClient(sysViewConn)
	b.sysViewClient = rawSysViewClient

	sysView := &SystemViewClient{client: rawSysViewClient}

	config := &logical.BackendConfig{
		StorageView: storage,
		Logger:      logger,
		System:      sysView,
		Config:      args.Config,
	}

	// Call the underlying backend factory after shims have been created
	// to set b.backend
	backend, err := b.factory(config)
	if err != nil {
		*reply = SetupReply{
			Error: plugin.NewBasicError(err),
		}
	}
	b.backend = backend

	return nil
}

func (b *backendPluginServer) Type(_ interface{}, reply *TypeReply) error {
	*reply = TypeReply{
		Type: b.backend.Type(),
	}

	return nil
}

func (b *backendPluginServer) RegisterLicense(args *RegisterLicenseArgs, reply *RegisterLicenseReply) error {
	err := b.backend.RegisterLicense(args.License)
	if err != nil {
		*reply = RegisterLicenseReply{
			Error: plugin.NewBasicError(err),
		}
	}

	return nil
}
