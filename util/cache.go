package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util/cache"
)

var mapCache map[string]*cache.Cache

func init() {
	mapCache = make(map[string]*cache.Cache)
}

func InitCache(obj, cacheExpirationCleanTime string, defaultExpiration, cleanupInterval time.Duration, maxCacheSize int64) {
	common.Infof("obj=%v", obj)
	_, ok := mapCache[obj]
	if ok {
		common.Warnf("repeat cache obj.%v", obj)
	} else {
		mapCache[obj] = cache.New(cacheExpirationCleanTime, defaultExpiration, cleanupInterval, maxCacheSize)
	}

	return
}

func CacheTcpAdd(obj string, key string, value []byte) (err error) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj")
	}
	return cacheAdd(ss, key, value, int64(len(value)))
}

func CacheHttpBodyAdd(obj string, key string, value []byte) (err error) {
	common.Infof("obj=%v.key=%s", obj, key)
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj")
	}

	return cacheAdd(ss, key, value, int64(len(value)))
}

func CacheHttpHeadAdd(obj string, key string, value interface{}) (err error) {
	common.Infof("obj=%v.key=%s", obj, key)
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj")
	}
	buf := &bytes.Buffer{}
	if err = binary.Write(buf, binary.BigEndian, value); nil != err {
		common.Errorf("binary Write error.=%s,%v", obj, err)
		return
	}

	return cacheAdd(ss, key, buf.Bytes(), int64(buf.Len()))
}

func cacheAdd(obj []string, key string, value interface{}, len int64) (err error) {
	for _, v := range obj {
		v, ok := mapCache[v]
		if ok {
			v.SetDefault(key, value, len)
		}
	}
	return
}

func CacheTcpGet(obj string, key string) (err error, value interface{}) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj"), nil
	}
	return cacheGet(ss, key)
}

func CacheHttpBodyGet(obj string, key string) (err error, value interface{}) {
	common.Infof("obj=%v.key=%s", obj, key)
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj"), nil
	}
	return cacheGet(ss, key)
}

func CacheHttpHeadGet(obj string, key string) (err error, value interface{}) {
	common.Infof("obj=%v.key=%s", obj, key)
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj"), nil
	}
	if err, value = cacheGet(ss, key); nil != err {
		return
	}
	head := http.Header{}
	buf := &bytes.Buffer{}
	if err = binary.Write(buf, binary.BigEndian, value); nil != err {
		common.Errorf("binary Write error.=%s", obj)
		return
	}
	if err = binary.Read(buf, binary.BigEndian, head); nil != err {
		common.Errorf("binary read error.=%s", obj)
		return
	}
	return err, head
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

func CacheTcpObjExist(obj string) bool {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.%s", obj)
		return false
	}
	return cacheObjExist(ss)
}

func CacheHttpObjExist(obj string) bool {
	common.Infof("obj=%v", obj)
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.%s", obj)
		return false
	}
	return cacheObjExist(ss)
}

func cacheObjExist(obj []string) bool {
	for _, v := range obj {
		_, ok := mapCache[v]
		if ok {
			return true
		}
	}
	return false
}
