package libkv

import (
	"errors"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	s := &FakeStore{data: []*store.KVPair{
		{Key: "foo", Value: []byte("bar")},
	}}
	kv := &LibKV{s}
	_, err := kv.Read("foo")

	require.NoError(t, err)

	s = &FakeStore{err: errors.New("fail")}
	kv = &LibKV{s}
	_, err = kv.Read("foo")

	assert.Error(t, err)
}

type FakeStore struct {
	err  error
	data []*store.KVPair
}

func (s *FakeStore) Put(_ string, _ []byte, _ *store.WriteOptions) error {
	return nil
}

func (s *FakeStore) Get(key string) (*store.KVPair, error) {
	if s.err != nil {
		return nil, s.err
	}

	for _, v := range s.data {
		if v.Key == key {
			return v, nil
		}
	}
	return nil, nil
}

func (s *FakeStore) Delete(_ string) error {
	return nil
}

func (s *FakeStore) Exists(_ string) (bool, error) {
	return false, nil
}

func (s *FakeStore) Watch(_ string, _ <-chan struct{}) (<-chan *store.KVPair, error) {
	return nil, nil
}

func (s *FakeStore) WatchTree(_ string, _ <-chan struct{}) (<-chan []*store.KVPair, error) {
	return nil, nil
}

func (s *FakeStore) NewLock(_ string, _ *store.LockOptions) (store.Locker, error) {
	return nil, nil
}

func (s *FakeStore) List(_ string) ([]*store.KVPair, error) {
	return nil, nil
}

func (s *FakeStore) DeleteTree(_ string) error {
	return nil
}

func (s *FakeStore) AtomicPut(_ string, _ []byte, _ *store.KVPair, _ *store.WriteOptions) (bool, *store.KVPair, error) {
	return false, nil, nil
}

func (s *FakeStore) AtomicDelete(_ string, _ *store.KVPair) (bool, error) {
	return false, nil
}

func (s *FakeStore) Close() {}
