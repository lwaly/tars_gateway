package util

import (
	"errors"
	"sync"

	"github.com/lwaly/tars_gateway/common"
)

var mapMaxConn map[string]int32
var mapObjConn map[string]int32
var mapObjMutex map[string]*sync.Mutex

const (
	defaultConn = 10000
)

func init() {
	mapMaxConn = make(map[string]int32)
	mapObjConn = make(map[string]int32)
	mapObjMutex = make(map[string]*sync.Mutex)
}

func InitRateLimit(obj string, maxConn int32) (err error) {
	if _, ok := mapMaxConn[obj]; ok == true {
		common.Errorf("repeat key.key=%s", obj)
		return errors.New("repeat key")
	}

	if 0 >= maxConn {
		maxConn = defaultConn
	}
	mapMaxConn[obj] = maxConn
	mapObjConn[obj] = 0
	mapObjMutex[obj] = new(sync.Mutex)
	return
}

func AddConn(obj string) (err error) {
	if maxConn, ok := mapMaxConn[obj]; ok != true {
		common.Errorf("fail to find rate limit obj.key=%s", obj)
		return errors.New("fail to find rate limit obj")
	} else {
		m, _ := mapObjMutex[obj]
		m.Lock()
		defer m.Unlock()
		conn, _ := mapObjConn[obj]
		if conn > maxConn {
			common.Errorf("More than the size of the max rate limit.key=%s,conn=%d", obj, maxConn)
			return errors.New("More than the size of the max rate limit")
		}
		mapObjConn[obj] = conn + 1
	}

	return
}

func TryAddConn(obj string) (err error) {
	if maxConn, ok := mapMaxConn[obj]; ok != true {
		common.Errorf("fail to find rate limit obj.key=%s", obj)
		return errors.New("fail to find rate limit obj")
	} else {
		m, _ := mapObjMutex[obj]
		m.Lock()
		defer m.Unlock()
		conn, _ := mapObjConn[obj]
		if conn > maxConn {
			common.Errorf("More than the size of the max rate limit.key=%s,conn=%d", obj, maxConn)
			return errors.New("More than the size of the max rate limit")
		}
	}

	return
}

func SubConn(obj string) (err error) {
	if maxConn, ok := mapMaxConn[obj]; ok != true {
		common.Errorf("fail to find rate limit obj.key=%s", obj)
		return errors.New("fail to find rate limit obj")
	} else {
		m, _ := mapObjMutex[obj]
		m.Lock()
		defer m.Unlock()
		conn, _ := mapObjConn[obj]
		if conn > maxConn {
			common.Errorf("More than the size of the max rate limit.key=%s,conn=%d", obj, maxConn)
			return errors.New("More than the size of the max rate limit")
		}
		mapObjConn[obj] = conn - 1
	}

	return
}
