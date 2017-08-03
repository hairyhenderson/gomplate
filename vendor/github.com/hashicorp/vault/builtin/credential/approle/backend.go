package approle

import (
	"sync"

	"github.com/hashicorp/vault/helper/locksutil"
	"github.com/hashicorp/vault/helper/salt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type backend struct {
	*framework.Backend

	// The salt value to be used by the information to be accessed only
	// by this backend.
	salt      *salt.Salt
	saltMutex sync.RWMutex

	// The view to use when creating the salt
	view logical.Storage

	// Guard to clean-up the expired SecretID entries
	tidySecretIDCASGuard uint32

	// Locks to make changes to role entries. These will be initialized to a
	// predefined number of locks when the backend is created, and will be
	// indexed based on salted role names.
	roleLocks []*locksutil.LockEntry

	// Locks to make changes to the storage entries of RoleIDs generated. These
	// will be initialized to a predefined number of locks when the backend is
	// created, and will be indexed based on the salted RoleIDs.
	roleIDLocks []*locksutil.LockEntry

	// Locks to make changes to the storage entries of SecretIDs generated.
	// These will be initialized to a predefined number of locks when the
	// backend is created, and will be indexed based on the HMAC-ed SecretIDs.
	secretIDLocks []*locksutil.LockEntry

	// Locks to make changes to the storage entries of SecretIDAccessors
	// generated. These will be initialized to a predefined number of locks
	// when the backend is created, and will be indexed based on the
	// SecretIDAccessors itself.
	secretIDAccessorLocks []*locksutil.LockEntry

	// secretIDListingLock is a dedicated lock for listing SecretIDAccessors
	// for all the SecretIDs issued against an approle
	secretIDListingLock sync.RWMutex
}

func Factory(conf *logical.BackendConfig) (logical.Backend, error) {
	b, err := Backend(conf)
	if err != nil {
		return nil, err
	}
	return b.Setup(conf)
}

func Backend(conf *logical.BackendConfig) (*backend, error) {
	// Create a backend object
	b := &backend{
		view: conf.StorageView,

		// Create locks to modify the registered roles
		roleLocks: locksutil.CreateLocks(),

		// Create locks to modify the generated RoleIDs
		roleIDLocks: locksutil.CreateLocks(),

		// Create locks to modify the generated SecretIDs
		secretIDLocks: locksutil.CreateLocks(),

		// Create locks to modify the generated SecretIDAccessors
		secretIDAccessorLocks: locksutil.CreateLocks(),
	}

	// Attach the paths and secrets that are to be handled by the backend
	b.Backend = &framework.Backend{
		// Register a periodic function that deletes the expired SecretID entries
		PeriodicFunc: b.periodicFunc,
		Help:         backendHelp,
		AuthRenew:    b.pathLoginRenew,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
			},
		},
		Paths: framework.PathAppend(
			rolePaths(b),
			[]*framework.Path{
				pathLogin(b),
				pathTidySecretID(b),
			},
		),
		Invalidate: b.invalidate,
	}
	return b, nil
}

func (b *backend) Salt() (*salt.Salt, error) {
	b.saltMutex.RLock()
	if b.salt != nil {
		defer b.saltMutex.RUnlock()
		return b.salt, nil
	}
	b.saltMutex.RUnlock()
	b.saltMutex.Lock()
	defer b.saltMutex.Unlock()
	if b.salt != nil {
		return b.salt, nil
	}
	salt, err := salt.NewSalt(b.view, &salt.Config{
		HashFunc: salt.SHA256Hash,
		Location: salt.DefaultLocation,
	})
	if err != nil {
		return nil, err
	}
	b.salt = salt
	return salt, nil
}

func (b *backend) invalidate(key string) {
	switch key {
	case salt.DefaultLocation:
		b.saltMutex.Lock()
		defer b.saltMutex.Unlock()
		b.salt = nil
	}
}

// periodicFunc of the backend will be invoked once a minute by the RollbackManager.
// RoleRole backend utilizes this function to delete expired SecretID entries.
// This could mean that the SecretID may live in the backend upto 1 min after its
// expiration. The deletion of SecretIDs are not security sensitive and it is okay
// to delay the removal of SecretIDs by a minute.
func (b *backend) periodicFunc(req *logical.Request) error {
	// Initiate clean-up of expired SecretID entries
	b.tidySecretID(req.Storage)
	return nil
}

const backendHelp = `
Any registered Role can authenticate itself with Vault. The credentials
depends on the constraints that are set on the Role. One common required
credential is the 'role_id' which is a unique identifier of the Role.
It can be retrieved from the 'role/<appname>/role-id' endpoint.

The default constraint configuration is 'bind_secret_id', which requires
the credential 'secret_id' to be presented during login. Refer to the
documentation for other types of constraints.`
