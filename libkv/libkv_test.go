package libkv

import (
	"errors"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	s := &FakeStore{data: []*store.KVPair{
		{Key: "foo", Value: []byte("bar")},
	}}
	kv := &LibKV{s}
	_, err := kv.Read("foo")

	assert.NoError(t, err)

	s = &FakeStore{err: errors.New("fail")}
	kv = &LibKV{s}
	_, err = kv.Read("foo")

	assert.Error(t, err)
}

type FakeStore struct {
	err  error
	data []*store.KVPair
}

func (s *FakeStore) Put(key string, value []byte, options *store.WriteOptions) error {
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

func (s *FakeStore) Delete(key string) error {
	return nil
}

func (s *FakeStore) Exists(key string) (bool, error) {
	return false, nil
}

func (s *FakeStore) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return nil, nil
}

func (s *FakeStore) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return nil, nil
}

func (s *FakeStore) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return nil, nil
}

func (s *FakeStore) List(directory string) ([]*store.KVPair, error) {
	return nil, nil
}

func (s *FakeStore) DeleteTree(directory string) error {
	return nil
}

func (s *FakeStore) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return false, nil, nil
}

func (s *FakeStore) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return false, nil
}

func (s *FakeStore) Close() {}
