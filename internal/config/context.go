package config

import (
	"context"
	"net/http"

	"github.com/spf13/afero"
)

// context keys
type (
	httpClientKey struct{}
	fsKey         struct{}
	dsKey         struct{}
	vaultKey      struct{}
)

// WithHTTPClient injects the given *http.Client into the context for later
// retrieval with HTTPClientFromContext.
func WithHTTPClient(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, httpClientKey{}, client)
}

// HTTPClientFromContext returns the *http.Client previously injected into the
// given context with WithHTTPClient. If none is found, http.DefaultClient is
// returned.
func HTTPClientFromContext(ctx context.Context) *http.Client {
	if ctx == nil {
		return http.DefaultClient
	}
	hc, ok := ctx.Value(httpClientKey{}).(*http.Client)
	if !ok {
		return http.DefaultClient
	}
	return hc
}

// WithFileSystem injects the given afero.Fs into the context for later
// retrieval with FileSystemFromContext.
func WithFileSystem(ctx context.Context, fs afero.Fs) context.Context {
	return context.WithValue(ctx, fsKey{}, fs)
}

// FileSystemFromContext returns the afero.Fs previously injected into the
// given context with WithFileSystem. If none is found, nil is
// returned.
func FileSystemFromContext(ctx context.Context) afero.Fs {
	if ctx == nil {
		return nil
	}
	fs, ok := ctx.Value(fsKey{}).(afero.Fs)
	if !ok {
		return nil
	}
	return fs
}

// WithDataSources injects the given afero.Fs into the context for later
// retrieval with DataSourcesFromContext.
func WithDataSources(ctx context.Context, ds map[string]DataSource) context.Context {
	return context.WithValue(ctx, dsKey{}, ds)
}

// DataSourcesFromContext returns the afero.Fs previously injected into the
// given context with WithDataSources. If none is found, nil is
// returned.
func DataSourcesFromContext(ctx context.Context) map[string]DataSource {
	if ctx == nil {
		return nil
	}
	fs, ok := ctx.Value(dsKey{}).(map[string]DataSource)
	if !ok {
		return nil
	}
	return fs
}

// // WithVaultClient injects the given vault client into the context for later
// // retrieval with VaultClientFromContext.
// func WithVaultClient(ctx context.Context, ds *vault.Vault) context.Context {
// 	return context.WithValue(ctx, vaultKey{}, ds)
// }

// // VaultClientFromContext returns the vault client previously injected into the
// // given context with WithVaultClient. If none is found, nil is
// // returned.
// func VaultClientFromContext(ctx context.Context) *vault.Vault {
// 	if ctx == nil {
// 		return nil
// 	}
// 	fs, ok := ctx.Value(vaultKey{}).(*vault.Vault)
// 	if !ok {
// 		return nil
// 	}
// 	return fs
// }
