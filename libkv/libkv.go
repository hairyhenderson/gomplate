package libkv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/libkv/store"
)

// LibKV -
type LibKV struct {
	store store.Store
}

// Login -
func (kv *LibKV) Login() error {
	return nil
}

// Logout -
func (kv *LibKV) Logout() {
}

// Read -
func (kv *LibKV) Read(path string) ([]byte, error) {
	data, err := kv.store.Get(path)
	if err != nil {
		return nil, fmt.Errorf("libkv Read failed for key %q: %w", path, err)
	}

	return data.Value, nil
}

// List -
func (kv *LibKV) List(path string) ([]byte, error) {
	data, err := kv.store.List(path)
	if err != nil {
		return nil, err
	}

	result := []map[string]string{}
	for _, pair := range data {
		// Remove the path from the key
		key := strings.TrimPrefix(
			pair.Key,
			strings.TrimLeft(path, "/"),
		)
		result = append(
			result,
			map[string]string{
				"key":   key,
				"value": string(pair.Value),
			},
		)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(result); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
