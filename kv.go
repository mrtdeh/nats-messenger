package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
)

type KeyValue struct {
	ctx context.Context
	l   *sync.RWMutex
	js  jetstream.JetStream
	kvs map[string]jetstream.KeyValue
}

func NewKeyValueStore(js jetstream.JetStream) *KeyValue {
	return &KeyValue{
		l:   &sync.RWMutex{},
		js:  js,
		kvs: make(map[string]jetstream.KeyValue),
	}
}

func (kv *KeyValue) CreateBuckets(ctx context.Context, buckets ...jetstream.KeyValueConfig) error {
	kv.l.Lock()
	defer kv.l.Unlock()

	for _, b := range buckets {
		if _, ok := kv.kvs[b.Bucket]; ok {
			return fmt.Errorf("bucket existed : %s", b.Bucket)
		}

		var k jetstream.KeyValue
		var err error
		// k, err = kv.js.KeyValue(b.Bucket)
		k, err = kv.js.KeyValue(ctx, b.Bucket)
		if err != nil {
			k, err = kv.js.CreateKeyValue(ctx, b)
			if err != nil {
				return err
			}
		}

		kv.kvs[b.Bucket] = k
		kv.ctx = ctx
	}
	return nil
}

func (kv *KeyValue) Put(bucket, key string, val any) (uint64, error) {
	kv.l.Lock()
	defer kv.l.Unlock()

	b, ok := kv.kvs[bucket]
	if !ok {
		return 0, fmt.Errorf("bucket not existed : %s", b.Bucket())
	}
	switch v := val.(type) {
	case string:
		return b.PutString(kv.ctx, key, v)
	case []byte:
		return b.Put(kv.ctx, key, v)
	default:
		d, err := json.Marshal(v)
		if err != nil {
			return 0, err
		}
		return b.Put(kv.ctx, key, d)
	}
}

func (kv *KeyValue) Get(bucket, key string) (jetstream.KeyValueEntry, error) {
	kv.l.RLock()
	defer kv.l.RUnlock()

	b, ok := kv.kvs[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket not existed : %s", b.Bucket())
	}
	return b.Get(kv.ctx, key)
}

func (kv *KeyValue) Keys(bucket string, prefix string) ([]string, error) {
	kv.l.RLock()
	defer kv.l.RUnlock()

	b, ok := kv.kvs[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket not existed : %s", b.Bucket())
	}

	keys, err := b.Keys(kv.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	var result []string

	if prefix == "" {
		return keys, nil
	}

	for _, key := range keys {
		if strings.HasPrefix(key, prefix) {
			result = append(result, key)
		}
	}
	return result, nil

}

func (kv *KeyValue) MapKeys(bucket string, prefix string) (map[string]bool, error) {
	keys, err := kv.Keys(bucket, prefix)
	if err != nil {
		return nil, err
	}

	var km = map[string]bool{}
	for _, k := range keys {
		km[k] = true
	}
	return km, nil
}
