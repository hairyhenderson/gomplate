package vault

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	log "github.com/mgutz/logxi/v1"

	"github.com/hashicorp/errwrap"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/helper/consts"
	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/hashicorp/vault/helper/locksutil"
	"github.com/hashicorp/vault/logical"
)

const (
	// expirationSubPath is the sub-path used for the expiration manager
	// view. This is nested under the system view.
	expirationSubPath = "expire/"

	// leaseViewPrefix is the prefix used for the ID based lookup of leases.
	leaseViewPrefix = "id/"

	// tokenViewPrefix is the prefix used for the token based lookup of leases.
	tokenViewPrefix = "token/"

	// maxRevokeAttempts limits how many revoke attempts are made
	maxRevokeAttempts = 6

	// revokeRetryBase is a baseline retry time
	revokeRetryBase = 10 * time.Second

	// minRevokeDelay is used to prevent an instant revoke on restore
	minRevokeDelay = 5 * time.Second

	// maxLeaseDuration is the default maximum lease duration
	maxLeaseTTL = 32 * 24 * time.Hour

	// defaultLeaseDuration is the default lease duration used when no lease is specified
	defaultLeaseTTL = maxLeaseTTL
)

// ExpirationManager is used by the Core to manage leases. Secrets
// can provide a lease, meaning that they can be renewed or revoked.
// If a secret is not renewed in timely manner, it may be expired, and
// the ExpirationManager will handle doing automatic revocation.
type ExpirationManager struct {
	router     *Router
	idView     *BarrierView
	tokenView  *BarrierView
	tokenStore *TokenStore
	logger     log.Logger

	pending     map[string]*time.Timer
	pendingLock sync.Mutex

	tidyLock int64
}

// NewExpirationManager creates a new ExpirationManager that is backed
// using a given view, and uses the provided router for revocation.
func NewExpirationManager(router *Router, view *BarrierView, ts *TokenStore, logger log.Logger) *ExpirationManager {
	if logger == nil {
		logger = log.New("expiration_manager")

	}
	exp := &ExpirationManager{
		router:     router,
		idView:     view.SubView(leaseViewPrefix),
		tokenView:  view.SubView(tokenViewPrefix),
		tokenStore: ts,
		logger:     logger,
		pending:    make(map[string]*time.Timer),
	}
	return exp
}

// setupExpiration is invoked after we've loaded the mount table to
// initialize the expiration manager
func (c *Core) setupExpiration() error {
	c.metricsMutex.Lock()
	defer c.metricsMutex.Unlock()
	// Create a sub-view
	view := c.systemBarrierView.SubView(expirationSubPath)

	// Create the manager
	mgr := NewExpirationManager(c.router, view, c.tokenStore, c.logger)
	c.expiration = mgr

	// Link the token store to this
	c.tokenStore.SetExpirationManager(mgr)

	// Restore the existing state
	c.logger.Info("expiration: restoring leases")
	if err := c.expiration.Restore(); err != nil {
		return fmt.Errorf("expiration state restore failed: %v", err)
	}
	return nil
}

// stopExpiration is used to stop the expiration manager before
// sealing the Vault.
func (c *Core) stopExpiration() error {
	if c.expiration != nil {
		if err := c.expiration.Stop(); err != nil {
			return err
		}
		c.metricsMutex.Lock()
		defer c.metricsMutex.Unlock()
		c.expiration = nil
	}
	return nil
}

// Tidy cleans up the dangling storage entries for leases. It scans the storage
// view to find all the available leases, checks if the token embedded in it is
// either empty or invalid and in both the cases, it revokes them. It also uses
// a token cache to avoid multiple lookups of the same token ID. It is normally
// not required to use the API that invokes this. This is only intended to
// clean up the corrupt storage due to bugs.
func (m *ExpirationManager) Tidy() error {
	var tidyErrors *multierror.Error

	if !atomic.CompareAndSwapInt64(&m.tidyLock, 0, 1) {
		m.logger.Warn("expiration: tidy operation on leases is already in progress")
		return fmt.Errorf("tidy operation on leases is already in progress")
	}

	defer atomic.CompareAndSwapInt64(&m.tidyLock, 1, 0)

	m.logger.Info("expiration: beginning tidy operation on leases")
	defer m.logger.Info("expiration: finished tidy operation on leases")

	// Create a cache to keep track of looked up tokens
	tokenCache := make(map[string]bool)
	var countLease, revokedCount, deletedCountInvalidToken, deletedCountEmptyToken int64

	tidyFunc := func(leaseID string) {
		countLease++
		if countLease%500 == 0 {
			m.logger.Info("expiration: tidying leases", "progress", countLease)
		}

		le, err := m.loadEntry(leaseID)
		if err != nil {
			tidyErrors = multierror.Append(tidyErrors, fmt.Errorf("failed to load the lease ID %q: %v", leaseID, err))
			return
		}

		if le == nil {
			tidyErrors = multierror.Append(tidyErrors, fmt.Errorf("nil entry for lease ID %q: %v", leaseID, err))
			return
		}

		var isValid, ok bool
		revokeLease := false
		if le.ClientToken == "" {
			m.logger.Trace("expiration: revoking lease which has an empty token", "lease_id", leaseID)
			revokeLease = true
			deletedCountEmptyToken++
			goto REVOKE_CHECK
		}

		isValid, ok = tokenCache[le.ClientToken]
		if !ok {
			saltedID := m.tokenStore.SaltID(le.ClientToken)
			lock := locksutil.LockForKey(m.tokenStore.tokenLocks, le.ClientToken)
			lock.RLock()
			te, err := m.tokenStore.lookupSalted(saltedID, true)
			lock.RUnlock()

			if err != nil {
				tidyErrors = multierror.Append(tidyErrors, fmt.Errorf("failed to lookup token: %v", err))
				return
			}

			if te == nil {
				m.logger.Trace("expiration: revoking lease which holds an invalid token", "lease_id", leaseID)
				revokeLease = true
				deletedCountInvalidToken++
				tokenCache[le.ClientToken] = false
			} else {
				tokenCache[le.ClientToken] = true
			}
			goto REVOKE_CHECK
		} else {
			if isValid {
				return
			} else {
				m.logger.Trace("expiration: revoking lease which contains an invalid token", "lease_id", leaseID)
				revokeLease = true
				deletedCountInvalidToken++
			}
			goto REVOKE_CHECK
		}

	REVOKE_CHECK:
		if revokeLease {
			// Force the revocation and skip going through the token store
			// again
			err = m.revokeCommon(leaseID, true, true)
			if err != nil {
				tidyErrors = multierror.Append(tidyErrors, fmt.Errorf("failed to revoke an invalid lease with ID %q: %v", leaseID, err))
				return
			}
			revokedCount++
		}
	}

	if err := logical.ScanView(m.idView, tidyFunc); err != nil {
		return err
	}

	m.logger.Debug("expiration: number of leases scanned", "count", countLease)
	m.logger.Debug("expiration: number of leases which had empty tokens", "count", deletedCountEmptyToken)
	m.logger.Debug("expiration: number of leases which had invalid tokens", "count", deletedCountInvalidToken)
	m.logger.Debug("expiration: number of leases successfully revoked", "count", revokedCount)

	return tidyErrors.ErrorOrNil()
}

// Restore is used to recover the lease states when starting.
// This is used after starting the vault.
func (m *ExpirationManager) Restore() error {
	m.pendingLock.Lock()
	defer m.pendingLock.Unlock()

	// Accumulate existing leases
	m.logger.Debug("expiration: collecting leases")
	existing, err := logical.CollectKeys(m.idView)
	if err != nil {
		return fmt.Errorf("failed to scan for leases: %v", err)
	}
	m.logger.Debug("expiration: leases collected", "num_existing", len(existing))

	// Make the channels used for the worker pool
	broker := make(chan string)
	quit := make(chan bool)
	// Buffer these channels to prevent deadlocks
	errs := make(chan error, len(existing))
	result := make(chan *leaseEntry, len(existing))

	// Use a wait group
	wg := &sync.WaitGroup{}

	// Create 64 workers to distribute work to
	for i := 0; i < consts.ExpirationRestoreWorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case leaseID, ok := <-broker:
					// broker has been closed, we are done
					if !ok {
						return
					}

					le, err := m.loadEntry(leaseID)
					if err != nil {
						errs <- err
						continue
					}

					// Write results out to the result channel
					result <- le

				// quit early
				case <-quit:
					return
				}
			}
		}()
	}

	// Distribute the collected keys to the workers in a go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i, leaseID := range existing {
			if i%500 == 0 {
				m.logger.Trace("expiration: leases loading", "progress", i)
			}

			select {
			case <-quit:
				return

			default:
				broker <- leaseID
			}
		}

		// Close the broker, causing worker routines to exit
		close(broker)
	}()

	// Restore each key by pulling from the result chan
	for i := 0; i < len(existing); i++ {
		select {
		case err := <-errs:
			// Close all go routines
			close(quit)

			return err

		case le := <-result:

			// If there is no entry, nothing to restore
			if le == nil {
				continue
			}

			// If there is no expiry time, don't do anything
			if le.ExpireTime.IsZero() {
				continue
			}

			// Determine the remaining time to expiration
			expires := le.ExpireTime.Sub(time.Now())
			if expires <= 0 {
				expires = minRevokeDelay
			}

			// Setup revocation timer
			m.pending[le.LeaseID] = time.AfterFunc(expires, func() {
				m.expireID(le.LeaseID)
			})
		}
	}

	// Let all go routines finish
	wg.Wait()

	if len(m.pending) > 0 {
		if m.logger.IsInfo() {
			m.logger.Info("expire: leases restored", "restored_lease_count", len(m.pending))
		}
	}

	return nil
}

// Stop is used to prevent further automatic revocations.
// This must be called before sealing the view.
func (m *ExpirationManager) Stop() error {
	// Stop all the pending expiration timers
	m.pendingLock.Lock()
	for _, timer := range m.pending {
		timer.Stop()
	}
	m.pending = make(map[string]*time.Timer)
	m.pendingLock.Unlock()
	return nil
}

// Revoke is used to revoke a secret named by the given LeaseID
func (m *ExpirationManager) Revoke(leaseID string) error {
	defer metrics.MeasureSince([]string{"expire", "revoke"}, time.Now())

	return m.revokeCommon(leaseID, false, false)
}

// revokeCommon does the heavy lifting. If force is true, we ignore a problem
// during revocation and still remove entries/index/lease timers
func (m *ExpirationManager) revokeCommon(leaseID string, force, skipToken bool) error {
	defer metrics.MeasureSince([]string{"expire", "revoke-common"}, time.Now())
	// Load the entry
	le, err := m.loadEntry(leaseID)
	if err != nil {
		return err
	}

	// If there is no entry, nothing to revoke
	if le == nil {
		return nil
	}

	// Revoke the entry
	if !skipToken || le.Auth == nil {
		if err := m.revokeEntry(le); err != nil {
			if !force {
				return err
			} else {
				if m.logger.IsWarn() {
					m.logger.Warn("revocation from the backend failed, but in force mode so ignoring", "error", err)
				}
			}
		}
	}

	// Delete the entry
	if err := m.deleteEntry(leaseID); err != nil {
		return err
	}

	// Delete the secondary index, but only if it's a leased secret (not auth)
	if le.Secret != nil {
		if err := m.removeIndexByToken(le.ClientToken, le.LeaseID); err != nil {
			return err
		}
	}

	// Clear the expiration handler
	m.pendingLock.Lock()
	if timer, ok := m.pending[leaseID]; ok {
		timer.Stop()
		delete(m.pending, leaseID)
	}
	m.pendingLock.Unlock()
	return nil
}

// RevokeForce works similarly to RevokePrefix but continues in the case of a
// revocation error; this is mostly meant for recovery operations
func (m *ExpirationManager) RevokeForce(prefix string) error {
	defer metrics.MeasureSince([]string{"expire", "revoke-force"}, time.Now())

	return m.revokePrefixCommon(prefix, true)
}

// RevokePrefix is used to revoke all secrets with a given prefix.
// The prefix maps to that of the mount table to make this simpler
// to reason about.
func (m *ExpirationManager) RevokePrefix(prefix string) error {
	defer metrics.MeasureSince([]string{"expire", "revoke-prefix"}, time.Now())

	return m.revokePrefixCommon(prefix, false)
}

// RevokeByToken is used to revoke all the secrets issued with a given token.
// This is done by using the secondary index. It also removes the lease entry
// for the token itself. As a result it should *ONLY* ever be called from the
// token store's revokeSalted function.
func (m *ExpirationManager) RevokeByToken(te *TokenEntry) error {
	defer metrics.MeasureSince([]string{"expire", "revoke-by-token"}, time.Now())
	// Lookup the leases
	existing, err := m.lookupByToken(te.ID)
	if err != nil {
		return fmt.Errorf("failed to scan for leases: %v", err)
	}

	// Revoke all the keys
	for idx, leaseID := range existing {
		if err := m.Revoke(leaseID); err != nil {
			return fmt.Errorf("failed to revoke '%s' (%d / %d): %v",
				leaseID, idx+1, len(existing), err)
		}
	}

	if te.Path != "" {
		tokenLeaseID := path.Join(te.Path, m.tokenStore.SaltID(te.ID))

		// We want to skip the revokeEntry call as that will call back into
		// revocation logic in the token store, which is what is running this
		// function in the first place -- it'd be a deadlock loop. Since the only
		// place that this function is called is revokeSalted in the token store,
		// we're already revoking the token, so we just want to clean up the lease.
		// This avoids spurious revocations later in the log when the timer runs
		// out, and eases up resource usage.
		return m.revokeCommon(tokenLeaseID, false, true)
	}

	return nil
}

func (m *ExpirationManager) revokePrefixCommon(prefix string, force bool) error {
	// Ensure there is a trailing slash
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Accumulate existing leases
	sub := m.idView.SubView(prefix)
	existing, err := logical.CollectKeys(sub)
	if err != nil {
		return fmt.Errorf("failed to scan for leases: %v", err)
	}

	// Revoke all the keys
	for idx, suffix := range existing {
		leaseID := prefix + suffix
		if err := m.revokeCommon(leaseID, force, false); err != nil {
			return fmt.Errorf("failed to revoke '%s' (%d / %d): %v",
				leaseID, idx+1, len(existing), err)
		}
	}
	return nil
}

// Renew is used to renew a secret using the given leaseID
// and a renew interval. The increment may be ignored.
func (m *ExpirationManager) Renew(leaseID string, increment time.Duration) (*logical.Response, error) {
	defer metrics.MeasureSince([]string{"expire", "renew"}, time.Now())
	// Load the entry
	le, err := m.loadEntry(leaseID)
	if err != nil {
		return nil, err
	}

	// Check if the lease is renewable
	if _, err := le.renewable(); err != nil {
		return nil, err
	}

	// Attempt to renew the entry
	resp, err := m.renewEntry(le, increment)
	if err != nil {
		return nil, err
	}

	// Fast-path if there is no lease
	if resp == nil || resp.Secret == nil || !resp.Secret.LeaseEnabled() {
		return resp, nil
	}

	// Validate the lease
	if err := resp.Secret.Validate(); err != nil {
		return nil, err
	}

	// Attach the LeaseID
	resp.Secret.LeaseID = leaseID

	// Update the lease entry
	le.Data = resp.Data
	le.Secret = resp.Secret
	le.ExpireTime = resp.Secret.ExpirationTime()
	le.LastRenewalTime = time.Now()
	if err := m.persistEntry(le); err != nil {
		return nil, err
	}

	// Update the expiration time
	m.updatePending(le, resp.Secret.LeaseTotal())

	// Return the response
	return resp, nil
}

// RenewToken is used to renew a token which does not need to
// invoke a logical backend.
func (m *ExpirationManager) RenewToken(req *logical.Request, source string, token string,
	increment time.Duration) (*logical.Response, error) {
	defer metrics.MeasureSince([]string{"expire", "renew-token"}, time.Now())
	// Compute the Lease ID
	leaseID := path.Join(source, m.tokenStore.SaltID(token))

	// Load the entry
	le, err := m.loadEntry(leaseID)
	if err != nil {
		return nil, err
	}

	// Check if the lease is renewable. Note that this also checks for a nil
	// lease and errors in that case as well.
	if _, err := le.renewable(); err != nil {
		return logical.ErrorResponse(err.Error()), logical.ErrInvalidRequest
	}

	// Attempt to renew the auth entry
	resp, err := m.renewAuthEntry(req, le, increment)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil
	}

	if resp.IsError() {
		return &logical.Response{
			Data: resp.Data,
		}, nil
	}

	if resp.Auth == nil || !resp.Auth.LeaseEnabled() {
		return &logical.Response{
			Auth: resp.Auth,
		}, nil
	}

	// Attach the ClientToken
	resp.Auth.ClientToken = token
	resp.Auth.Increment = 0

	// Update the lease entry
	le.Auth = resp.Auth
	le.ExpireTime = resp.Auth.ExpirationTime()
	le.LastRenewalTime = time.Now()
	if err := m.persistEntry(le); err != nil {
		return nil, err
	}

	// Update the expiration time
	m.updatePending(le, resp.Auth.LeaseTotal())
	return &logical.Response{
		Auth: resp.Auth,
	}, nil
}

// Register is used to take a request and response with an associated
// lease. The secret gets assigned a LeaseID and the management of
// of lease is assumed by the expiration manager.
func (m *ExpirationManager) Register(req *logical.Request, resp *logical.Response) (id string, retErr error) {
	defer metrics.MeasureSince([]string{"expire", "register"}, time.Now())

	if req.ClientToken == "" {
		return "", fmt.Errorf("expiration: cannot register a lease with an empty client token")
	}

	// Ignore if there is no leased secret
	if resp == nil || resp.Secret == nil {
		return "", nil
	}

	// Validate the secret
	if err := resp.Secret.Validate(); err != nil {
		return "", err
	}

	// Create a lease entry
	leaseUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}

	leaseID := path.Join(req.Path, leaseUUID)

	defer func() {
		// If there is an error we want to rollback as much as possible (note
		// that errors here are ignored to do as much cleanup as we can). We
		// want to revoke a generated secret (since an error means we may not
		// be successfully tracking it), remove indexes, and delete the entry.
		if retErr != nil {
			revResp, err := m.router.Route(logical.RevokeRequest(req.Path, resp.Secret, resp.Data))
			if err != nil {
				retErr = multierror.Append(retErr, errwrap.Wrapf("an additional internal error was encountered revoking the newly-generated secret: {{err}}", err))
			} else if revResp != nil && revResp.IsError() {
				retErr = multierror.Append(retErr, errwrap.Wrapf("an additional error was encountered revoking the newly-generated secret: {{err}}", revResp.Error()))
			}

			if err := m.deleteEntry(leaseID); err != nil {
				retErr = multierror.Append(retErr, errwrap.Wrapf("an additional error was encountered deleting any lease associated with the newly-generated secret: {{err}}", err))
			}

			if err := m.removeIndexByToken(req.ClientToken, leaseID); err != nil {
				retErr = multierror.Append(retErr, errwrap.Wrapf("an additional error was encountered removing lease indexes associated with the newly-generated secret: {{err}}", err))
			}
		}
	}()

	le := leaseEntry{
		LeaseID:     leaseID,
		ClientToken: req.ClientToken,
		Path:        req.Path,
		Data:        resp.Data,
		Secret:      resp.Secret,
		IssueTime:   time.Now(),
		ExpireTime:  resp.Secret.ExpirationTime(),
	}

	// Encode the entry
	if err := m.persistEntry(&le); err != nil {
		return "", err
	}

	// Maintain secondary index by token
	if err := m.createIndexByToken(le.ClientToken, le.LeaseID); err != nil {
		return "", err
	}

	// Setup revocation timer if there is a lease
	m.updatePending(&le, resp.Secret.LeaseTotal())

	// Done
	return le.LeaseID, nil
}

// RegisterAuth is used to take an Auth response with an associated lease.
// The token does not get a LeaseID, but the lease management is handled by
// the expiration manager.
func (m *ExpirationManager) RegisterAuth(source string, auth *logical.Auth) error {
	defer metrics.MeasureSince([]string{"expire", "register-auth"}, time.Now())

	if auth.ClientToken == "" {
		return fmt.Errorf("expiration: cannot register an auth lease with an empty token")
	}

	if strings.Contains(source, "..") {
		return fmt.Errorf("expiration: %s", consts.ErrPathContainsParentReferences)
	}

	// Create a lease entry
	le := leaseEntry{
		LeaseID:     path.Join(source, m.tokenStore.SaltID(auth.ClientToken)),
		ClientToken: auth.ClientToken,
		Auth:        auth,
		Path:        source,
		IssueTime:   time.Now(),
		ExpireTime:  auth.ExpirationTime(),
	}

	// Encode the entry
	if err := m.persistEntry(&le); err != nil {
		return err
	}

	// Setup revocation timer
	m.updatePending(&le, auth.LeaseTotal())
	return nil
}

// FetchLeaseTimesByToken is a helper function to use token values to compute
// the leaseID, rather than pushing that logic back into the token store.
func (m *ExpirationManager) FetchLeaseTimesByToken(source, token string) (*leaseEntry, error) {
	defer metrics.MeasureSince([]string{"expire", "fetch-lease-times-by-token"}, time.Now())

	// Compute the Lease ID
	leaseID := path.Join(source, m.tokenStore.SaltID(token))
	return m.FetchLeaseTimes(leaseID)
}

// FetchLeaseTimes is used to fetch the issue time, expiration time, and last
// renewed time of a lease entry. It returns a leaseEntry itself, but with only
// those values copied over.
func (m *ExpirationManager) FetchLeaseTimes(leaseID string) (*leaseEntry, error) {
	defer metrics.MeasureSince([]string{"expire", "fetch-lease-times"}, time.Now())

	// Load the entry
	le, err := m.loadEntry(leaseID)
	if err != nil {
		return nil, err
	}
	if le == nil {
		return nil, nil
	}

	ret := &leaseEntry{
		IssueTime:       le.IssueTime,
		ExpireTime:      le.ExpireTime,
		LastRenewalTime: le.LastRenewalTime,
	}
	if le.Secret != nil {
		ret.Secret = &logical.Secret{}
		ret.Secret.Renewable = le.Secret.Renewable
		ret.Secret.TTL = le.Secret.TTL
	}
	if le.Auth != nil {
		ret.Auth = &logical.Auth{}
		ret.Auth.Renewable = le.Auth.Renewable
		ret.Auth.TTL = le.Auth.TTL
	}

	return ret, nil
}

// updatePending is used to update a pending invocation for a lease
func (m *ExpirationManager) updatePending(le *leaseEntry, leaseTotal time.Duration) {
	m.pendingLock.Lock()
	defer m.pendingLock.Unlock()

	// Check for an existing timer
	timer, ok := m.pending[le.LeaseID]

	// Create entry if it does not exist
	if !ok && leaseTotal > 0 {
		timer := time.AfterFunc(leaseTotal, func() {
			m.expireID(le.LeaseID)
		})
		m.pending[le.LeaseID] = timer
		return
	}

	// Delete the timer if the expiration time is zero
	if ok && leaseTotal == 0 {
		timer.Stop()
		delete(m.pending, le.LeaseID)
		return
	}

	// Extend the timer by the lease total
	if ok && leaseTotal > 0 {
		timer.Reset(leaseTotal)
	}
}

// expireID is invoked when a given ID is expired
func (m *ExpirationManager) expireID(leaseID string) {
	// Clear from the pending expiration
	m.pendingLock.Lock()
	delete(m.pending, leaseID)
	m.pendingLock.Unlock()

	for attempt := uint(0); attempt < maxRevokeAttempts; attempt++ {
		err := m.Revoke(leaseID)
		if err == nil {
			if m.logger.IsInfo() {
				m.logger.Info("expire: revoked lease", "lease_id", leaseID)
			}
			return
		}
		m.logger.Error("expire: failed to revoke lease", "lease_id", leaseID, "error", err)
		time.Sleep((1 << attempt) * revokeRetryBase)
	}
	m.logger.Error("expire: maximum revoke attempts reached", "lease_id", leaseID)
}

// revokeEntry is used to attempt revocation of an internal entry
func (m *ExpirationManager) revokeEntry(le *leaseEntry) error {
	// Revocation of login tokens is special since we can by-pass the
	// backend and directly interact with the token store
	if le.Auth != nil {
		if err := m.tokenStore.RevokeTree(le.ClientToken); err != nil {
			return fmt.Errorf("failed to revoke token: %v", err)
		}

		return nil
	}

	// Handle standard revocation via backends
	resp, err := m.router.Route(logical.RevokeRequest(
		le.Path, le.Secret, le.Data))
	if err != nil || (resp != nil && resp.IsError()) {
		return fmt.Errorf("failed to revoke entry: resp:%#v err:%s", resp, err)
	}
	return nil
}

// renewEntry is used to attempt renew of an internal entry
func (m *ExpirationManager) renewEntry(le *leaseEntry, increment time.Duration) (*logical.Response, error) {
	secret := *le.Secret
	secret.IssueTime = le.IssueTime
	secret.Increment = increment
	secret.LeaseID = ""

	req := logical.RenewRequest(le.Path, &secret, le.Data)
	resp, err := m.router.Route(req)
	if err != nil || (resp != nil && resp.IsError()) {
		return nil, fmt.Errorf("failed to renew entry: resp:%#v err:%s", resp, err)
	}
	return resp, nil
}

// renewAuthEntry is used to attempt renew of an auth entry. Only the token
// store should get the actual token ID intact.
func (m *ExpirationManager) renewAuthEntry(req *logical.Request, le *leaseEntry, increment time.Duration) (*logical.Response, error) {
	auth := *le.Auth
	auth.IssueTime = le.IssueTime
	auth.Increment = increment
	if strings.HasPrefix(le.Path, "auth/token/") {
		auth.ClientToken = le.ClientToken
	} else {
		auth.ClientToken = ""
	}

	authReq := logical.RenewAuthRequest(le.Path, &auth, nil)
	authReq.Connection = req.Connection
	resp, err := m.router.Route(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to renew entry: %v", err)
	}
	return resp, nil
}

// loadEntry is used to read a lease entry
func (m *ExpirationManager) loadEntry(leaseID string) (*leaseEntry, error) {
	out, err := m.idView.Get(leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to read lease entry: %v", err)
	}
	if out == nil {
		return nil, nil
	}
	le, err := decodeLeaseEntry(out.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode lease entry: %v", err)
	}
	return le, nil
}

// persistEntry is used to persist a lease entry
func (m *ExpirationManager) persistEntry(le *leaseEntry) error {
	// Encode the entry
	buf, err := le.encode()
	if err != nil {
		return fmt.Errorf("failed to encode lease entry: %v", err)
	}

	// Write out to the view
	ent := logical.StorageEntry{
		Key:   le.LeaseID,
		Value: buf,
	}
	if err := m.idView.Put(&ent); err != nil {
		return fmt.Errorf("failed to persist lease entry: %v", err)
	}
	return nil
}

// deleteEntry is used to delete a lease entry
func (m *ExpirationManager) deleteEntry(leaseID string) error {
	if err := m.idView.Delete(leaseID); err != nil {
		return fmt.Errorf("failed to delete lease entry: %v", err)
	}
	return nil
}

// createIndexByToken creates a secondary index from the token to a lease entry
func (m *ExpirationManager) createIndexByToken(token, leaseID string) error {
	ent := logical.StorageEntry{
		Key:   m.tokenStore.SaltID(token) + "/" + m.tokenStore.SaltID(leaseID),
		Value: []byte(leaseID),
	}
	if err := m.tokenView.Put(&ent); err != nil {
		return fmt.Errorf("failed to persist lease index entry: %v", err)
	}
	return nil
}

// indexByToken looks up the secondary index from the token to a lease entry
func (m *ExpirationManager) indexByToken(token, leaseID string) (*logical.StorageEntry, error) {
	key := m.tokenStore.SaltID(token) + "/" + m.tokenStore.SaltID(leaseID)
	entry, err := m.tokenView.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to look up secondary index entry")
	}
	return entry, nil
}

// removeIndexByToken removes the secondary index from the token to a lease entry
func (m *ExpirationManager) removeIndexByToken(token, leaseID string) error {
	key := m.tokenStore.SaltID(token) + "/" + m.tokenStore.SaltID(leaseID)
	if err := m.tokenView.Delete(key); err != nil {
		return fmt.Errorf("failed to delete lease index entry: %v", err)
	}
	return nil
}

// lookupByToken is used to lookup all the leaseID's via the
func (m *ExpirationManager) lookupByToken(token string) ([]string, error) {
	// Scan via the index for sub-leases
	prefix := m.tokenStore.SaltID(token) + "/"
	subKeys, err := m.tokenView.List(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list leases: %v", err)
	}

	// Read each index entry
	leaseIDs := make([]string, 0, len(subKeys))
	for _, sub := range subKeys {
		out, err := m.tokenView.Get(prefix + sub)
		if err != nil {
			return nil, fmt.Errorf("failed to read lease index: %v", err)
		}
		if out == nil {
			continue
		}
		leaseIDs = append(leaseIDs, string(out.Value))
	}
	return leaseIDs, nil
}

// emitMetrics is invoked periodically to emit statistics
func (m *ExpirationManager) emitMetrics() {
	m.pendingLock.Lock()
	num := len(m.pending)
	m.pendingLock.Unlock()
	metrics.SetGauge([]string{"expire", "num_leases"}, float32(num))
}

// leaseEntry is used to structure the values the expiration
// manager stores. This is used to handle renew and revocation.
type leaseEntry struct {
	LeaseID         string                 `json:"lease_id"`
	ClientToken     string                 `json:"client_token"`
	Path            string                 `json:"path"`
	Data            map[string]interface{} `json:"data"`
	Secret          *logical.Secret        `json:"secret"`
	Auth            *logical.Auth          `json:"auth"`
	IssueTime       time.Time              `json:"issue_time"`
	ExpireTime      time.Time              `json:"expire_time"`
	LastRenewalTime time.Time              `json:"last_renewal_time"`
}

// encode is used to JSON encode the lease entry
func (le *leaseEntry) encode() ([]byte, error) {
	return json.Marshal(le)
}

func (le *leaseEntry) renewable() (bool, error) {
	var err error
	switch {
	// If there is no entry, cannot review
	case le == nil || le.ExpireTime.IsZero():
		err = fmt.Errorf("lease not found or lease is not renewable")
	// Determine if the lease is expired
	case le.ExpireTime.Before(time.Now()):
		err = fmt.Errorf("lease expired")
	// Determine if the lease is renewable
	case le.Secret != nil && !le.Secret.Renewable:
		err = fmt.Errorf("lease is not renewable")
	case le.Auth != nil && !le.Auth.Renewable:
		err = fmt.Errorf("lease is not renewable")
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func (le *leaseEntry) ttl() int64 {
	return int64(le.ExpireTime.Sub(time.Now().Round(time.Second)).Seconds())
}

// decodeLeaseEntry is used to reverse encode and return a new entry
func decodeLeaseEntry(buf []byte) (*leaseEntry, error) {
	out := new(leaseEntry)
	return out, jsonutil.DecodeJSON(buf, out)
}
