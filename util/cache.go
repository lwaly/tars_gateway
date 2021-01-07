package util

import (
	"encoding/json"
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

func InitCache(obj, cacheExpirationCleanTime string, defaultExpiration time.Duration, maxCacheSize int64) {
	_, ok := mapCache[obj]
	if ok {
		common.Warnf("repeat cache obj.%v", obj)
	} else {
		mapCache[obj] = cache.New(cacheExpirationCleanTime, defaultExpiration, maxCacheSize)
	}

	return
}

func CacheTcpAdd(obj string, key string, value []byte) (err error) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj")
	}
	return cacheAdd(ss, key, value)
}

func CacheHttpBodyAdd(obj string, key string, value []byte) (err error) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj")
	}

	return cacheAdd(ss, key, value)
}

func CacheHttpHeadAdd(obj string, key string, head *http.Header) (err error) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj")
	}
	var b []byte
	if b, err = json.Marshal(head); nil != err {
		common.Errorf("error obj.=%s", obj)
		return
	}

	return cacheAdd(ss, key, b)
}

func cacheAdd(obj []string, key string, b []byte) (err error) {
	for i := len(obj); i > 0; i-- {
		tempObj := ""
		for j := 0; j < i; j++ {
			tempObj += obj[j] + "."
		}
		tempObj = tempObj[0 : len(tempObj)-1]
		v, ok := mapCache[tempObj]
		if ok {
			return v.SetDefault(key, b, int64(len(b)))
		}
	}

	return errors.New("not find")
}

func CacheTcpGet(obj string, key string) (err error, value interface{}) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj"), nil
	}
	return cacheGet(ss, key)
}

func cacheHttpGet(obj string, key string) (err error, value interface{}) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.=%s", obj)
		return errors.New("error obj"), nil
	}
	return cacheGet(ss, key)
}

func CacheHttpBodyGet(obj string, key string) (err error, value interface{}) {
	return cacheHttpGet(obj, key)
}

func CacheHttpHeadGet(obj string, key string) (err error, head http.Header, n int64) {
	if errTemp, value := cacheHttpGet(obj, key); nil != err {
		return errTemp, nil, 0
	} else {
		if b, ok := value.([]byte); ok {
			err = json.Unmarshal(b, &head)
			return err, head, int64(len(b))
		} else {
			return errors.New("not find"), nil, 0
		}
	}
}

func cacheGet(obj []string, key string) (err error, value interface{}) {
	for i := len(obj); i > 0; i-- {
		tempObj := ""
		for j := 0; j < i; j++ {
			tempObj += obj[j] + "."
		}
		tempObj = tempObj[0 : len(tempObj)-1]
		v, ok := mapCache[tempObj]
		if ok {
			return v.Get(key)
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
