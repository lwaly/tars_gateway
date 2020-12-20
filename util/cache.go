package util

import (
	"errors"
	"strings"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util/cache"
)

var mapCache map[string]*cache.Cache

func init() {
	mapCache = make(map[string]*cache.Cache)
}

func InitCache(obj string, defaultExpiration, cleanupInterval time.Duration, maxCacheSize int64) {
	_, ok := mapCache[obj]
	if ok {
		common.Warnf("repeat cache obj.%v", obj)
	} else {
		mapCache[obj] = cache.New(defaultExpiration, cleanupInterval, maxCacheSize)
	}

	return
}

func CacheTcpAdd(obj string, key string, value []byte) (err error) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}
	return cacheAdd(ss, key, value)
}

func CacheHttpAdd(obj string, key string, value []byte) (err error) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}
	return cacheAdd(ss, key, value)
}

func cacheAdd(obj []string, key string, value []byte) (err error) {
	for _, v := range obj {
		v, ok := mapCache[v]
		if ok {
			v.SetDefault(key, value)
		}
	}
	return
}

func CacheTcpGet(obj string, key string) (err error, value interface{}) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj"), nil
	}
	return cacheGet(ss, key)
}

func CacheHttpGet(obj string, key string) (err error, value interface{}) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj"), nil
	}
	return cacheGet(ss, key)
}

func cacheGet(obj []string, key string) (err error, value interface{}) {
	for _, v := range obj {
		v, ok := mapCache[v]
		if ok {
			v1, ok1 := v.Get(key)
			if ok1 {
				return nil, v1
			}
		}
	}
	return errors.New("not find"), nil
}
