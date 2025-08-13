package datafs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/azure"
)

// azureKVFSProvider provides Azure Key Vault filesystem support
var azureKVFSProvider = fsimpl.FSProviderFunc(newAzureKVFS, "azure+kv")

// azureKVFS implements fs.FS for Azure Key Vault
type azureKVFS struct {
	client   *azure.KeyVaultClient
	vaultURL string
	ctx      context.Context
}

// newAzureKVFS creates a new Azure Key Vault filesystem
func newAzureKVFS(u *url.URL) (fs.FS, error) {
	ctx := context.Background()
	var vaultURL string

	// Handle different URL formats
	if u.Host != "" {
		// Full URL: azure+kv://myvault.vault.azure.net
		vaultURL = fmt.Sprintf("https://%s", u.Host)
	} else {
		// For opaque URLs (azure+kv:secretname) or just scheme (azure+kv:)
		// Get vault URL from environment
		envVaultURL := os.Getenv("AZURE_KEYVAULT_URL")
		if envVaultURL == "" {
			return nil, fmt.Errorf("AZURE_KEYVAULT_URL environment variable must be set when using opaque azure+kv URLs")
		}
		vaultURL = envVaultURL
	}

	// Create Azure Key Vault client
	client, err := azure.NewKeyVaultClient(ctx, vaultURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Key Vault client: %w", err)
	}

	return &azureKVFS{
		client:   client,
		vaultURL: vaultURL,
		ctx:      ctx,
	}, nil
}

// Open opens a file from Azure Key Vault
func (fsys *azureKVFS) Open(name string) (fs.File, error) {
	// Handle empty name (root directory)
	if name == "" || name == "." {
		return fsys.openDir()
	}

	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	// Remove leading slash if present
	secretPath := strings.TrimPrefix(name, "/")

	// Parse version from secret path (format: secretname/version)
	version := ""
	secretName := secretPath

	// Support Azure standard format: secret/version
	if parts := strings.Split(secretPath, "/"); len(parts) == 2 {
		secretName = parts[0]
		version = parts[1]
	}

	// Get the secret value
	secretValue, err := fsys.client.GetSecret(fsys.ctx, secretName, version)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	return &azureKVFile{
		name:    secretName,
		content: []byte(secretValue),
		modTime: time.Now(),
	}, nil
}

// openDir returns a directory listing of all secrets
func (fsys *azureKVFS) openDir() (fs.File, error) {
	secrets, err := fsys.client.ListSecrets(fsys.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var entries []fs.DirEntry
	for _, secret := range secrets {
		entries = append(entries, &azureKVDirEntry{
			name: secret,
		})
	}

	return &azureKVDirFile{
		name:    ".",
		entries: entries,
		modTime: time.Now(),
	}, nil
}

// azureKVFile implements fs.File for individual secrets
type azureKVFile struct {
	name    string
	content []byte
	offset  int64
	modTime time.Time
}

func (f *azureKVFile) Close() error               { return nil }
func (f *azureKVFile) Stat() (fs.FileInfo, error) { return f, nil }

func (f *azureKVFile) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.content)) {
		return 0, io.EOF
	}
	n := copy(b, f.content[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// fs.FileInfo implementation
func (f *azureKVFile) Name() string       { return path.Base(f.name) }
func (f *azureKVFile) Size() int64        { return int64(len(f.content)) }
func (f *azureKVFile) Mode() fs.FileMode  { return 0444 }
func (f *azureKVFile) ModTime() time.Time { return f.modTime }
func (f *azureKVFile) IsDir() bool        { return false }
func (f *azureKVFile) Sys() any           { return nil }

// azureKVDirFile implements fs.File for directory listings
type azureKVDirFile struct {
	name    string
	entries []fs.DirEntry
	offset  int
	modTime time.Time
}

func (f *azureKVDirFile) Close() error               { return nil }
func (f *azureKVDirFile) Stat() (fs.FileInfo, error) { return f, nil }

func (f *azureKVDirFile) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("cannot read from directory")
}

func (f *azureKVDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 || f.offset >= len(f.entries) {
		return f.entries[f.offset:], nil
	}

	end := f.offset + n
	if end > len(f.entries) {
		end = len(f.entries)
	}

	entries := f.entries[f.offset:end]
	f.offset = end
	return entries, nil
}

// fs.FileInfo implementation for directory
func (f *azureKVDirFile) Name() string       { return f.name }
func (f *azureKVDirFile) Size() int64        { return 0 }
func (f *azureKVDirFile) Mode() fs.FileMode  { return fs.ModeDir | 0555 }
func (f *azureKVDirFile) ModTime() time.Time { return f.modTime }
func (f *azureKVDirFile) IsDir() bool        { return true }
func (f *azureKVDirFile) Sys() any           { return nil }

// azureKVDirEntry implements fs.DirEntry
type azureKVDirEntry struct {
	name string
}

func (e *azureKVDirEntry) Name() string      { return e.name }
func (e *azureKVDirEntry) IsDir() bool       { return false }
func (e *azureKVDirEntry) Type() fs.FileMode { return 0444 }

func (e *azureKVDirEntry) Info() (fs.FileInfo, error) {
	return &azureKVFile{
		name:    e.name,
		content: []byte{},
		modTime: time.Now(),
	}, nil
}
