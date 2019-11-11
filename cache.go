package apollo_client

import (
	"errors"
	"sync"
)

type CacheInterface interface {
	Get(key string) (value string, err error)
	Set(key string, value string) (err error)
	Del(key string) (affected bool)
	Range(f func(key, value string) bool)
}

type CacheFactory interface {
	Create() CacheInterface
}

type DefaultCache struct {
	defaultCache sync.Map
}

func (d *DefaultCache) Get(key string) (value string, err error) {
	v, ok := d.defaultCache.Load(key)
	if !ok {
		return "", errors.New("load default cache fail")
	}
	return v.(string), nil
}

func (d *DefaultCache) Set(key, value string) (err error) {
	d.defaultCache.Store(key, value)
	return nil
}

func (d *DefaultCache) Range(f func(key, value string) bool) {
	d.defaultCache.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return true
		}
		v, ok := key.(string)
		if !ok {
			return true
		}
		return f(k, v)
	})
}

func (d *DefaultCache) Del(key string) (affected bool) {
	d.defaultCache.Delete(key)
	return true
}

type DefaultCacheFactory struct {
}

func (d *DefaultCacheFactory) Create() CacheInterface {
	return &DefaultCache{}
}
