package libkv

import (
	"log"
	"strconv"

	"github.com/docker/libkv/store"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

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
		return nil, err
	}

	return data.Value, nil
}

func mustParseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func mustParseInt(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 16)
	return i
}
