package etcd

import (
	"context"
	"strings"
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"github.com/rpcxio/libkv/store"
	"go.etcd.io/etcd/client/v2"
)

var (
	once       sync.Once
	etcdConf   client.Config
	etcdClient client.Client
)

// Init etcd初始化
func Init(conf *config.Config) error {
	var rErr error
	once.Do(func() {
		var err error
		host := conf.Get("config_center.host")
		if host == nil {
			rErr = errors.New("etcd host not found")
			return
		}
		etcdConf.Endpoints = []string{host.(string)}

		etcdClient, err = client.New(etcdConf)
		if err != nil {
			rErr = errors.Wrapf(err, "init etcd")
			return
		}
	})

	return rErr
}

func Client() client.Client {
	return etcdClient
}

func Api() client.KeysAPI {
	return client.NewKeysAPI(etcdClient)
}

// Watch for changes on a "key"
// It returns a channel that will receive changes or pass
// on errors. Upon creation, the current value will first
// be sent to the channel. Providing a non-nil stopCh can
// be used to stop watching.
func Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	opts := &client.WatcherOptions{Recursive: false}

	watcher := client.NewKeysAPI(etcdClient).Watcher(normalize(key), opts)

	// watchCh is sending back events to the caller
	watchCh := make(chan *store.KVPair)

	go func() {
		defer close(watchCh)

		// Get the current value
		pair, err := Get(key)
		if err != nil {
			return
		}

		// Push the current value through the channel.
		watchCh <- pair

		for {
			// Check if the watch was stopped by the caller
			select {
			case <-stopCh:
				return
			default:
			}

			result, err := watcher.Next(context.Background())

			if err != nil {
				return
			}

			watchCh <- &store.KVPair{
				Key:       key,
				Value:     []byte(result.Node.Value),
				LastIndex: result.Node.ModifiedIndex,
			}
		}
	}()

	return watchCh, nil
}

// Get the value at "key", returns the last modified
// index to use in conjunction to Atomic calls
func Get(key string) (pair *store.KVPair, err error) {
	getOpts := &client.GetOptions{
		Quorum: true,
	}

	result, err := client.NewKeysAPI(etcdClient).Get(context.Background(), normalize(key), getOpts)
	if err != nil {
		if keyNotFound(err) {
			return nil, store.ErrKeyNotFound
		}
		return nil, err
	}

	pair = &store.KVPair{
		Key:       key,
		Value:     []byte(result.Node.Value),
		LastIndex: result.Node.ModifiedIndex,
	}

	return pair, nil
}

// Normalize the key for usage in Etcd
func normalize(key string) string {
	key = store.Normalize(key)
	return strings.TrimPrefix(key, "/")
}

func keyNotFound(err error) bool {
	if err != nil {
		if etcdError, ok := err.(client.Error); ok {
			if etcdError.Code == client.ErrorCodeKeyNotFound ||
				etcdError.Code == client.ErrorCodeNotFile ||
				etcdError.Code == client.ErrorCodeNotDir {
				return true
			}
		}
	}
	return false
}
