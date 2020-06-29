package util

import (
	"errors"
	"sync"
	"time"

	"github.com/lwaly/tars_gateway/common"
)

type stRateLimit struct {
	maxRate    int64
	maxRatePer int64
	objRate    int64
	objRatePer int64
	per        time.Duration
	reset      time.Time
	objMutex   *sync.Mutex
}

var mapRateLimit map[string]*stRateLimit

func init() {
	mapRateLimit = make(map[string]*stRateLimit)
}

func InitRateLimit(obj string, maxRate, maxRatePer, per int64) (err error) {
	if 0 >= maxRate || 0 >= maxRatePer || 0 >= per || "" == obj {
		common.Errorf("param error.obj=%s,maxRate=%d,maxRatePer=%d,per=%d", obj, maxRate, maxRatePer, per)
		return errors.New("param error")
	}

	if _, ok := mapRateLimit[obj]; ok {
		common.Errorf("repeat key.key=%s", obj)
		return errors.New("repeat key")
	}

	lstRateLimit := stRateLimit{}
	lstRateLimit.maxRate = maxRate
	lstRateLimit.maxRatePer = maxRatePer
	lstRateLimit.objRate = 0
	lstRateLimit.objRatePer = 0
	lstRateLimit.per = time.Duration(per)
	lstRateLimit.reset = time.Now().Add(lstRateLimit.per * time.Millisecond)
	lstRateLimit.objMutex = new(sync.Mutex)

	mapRateLimit[obj] = &lstRateLimit
	return
}

func AddRate(obj string, n int64) (err error) {
	if v, ok := mapRateLimit[obj]; ok {
		v.objMutex.Lock()
		defer v.objMutex.Unlock()

		if time.Now().After(v.reset) {
			v.reset = time.Now().Add(v.per * time.Millisecond)
			v.objRate = 0
			v.objRatePer = 0
		}

		if v.objRate > v.maxRate || v.objRatePer > v.maxRatePer {
			common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
				"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d", obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer)
			return errors.New("More than the size of the max rate limit")
		}

		v.objRate += n
		v.objRatePer += n
	} else {
		common.Errorf("fail to find rate limit obj.key=%s", obj)
		return errors.New("fail to find rate limit obj")
	}

	return
}

func TryAddRate(obj string, n int64) (err error) {
	if v, ok := mapRateLimit[obj]; ok {
		v.objMutex.Lock()
		defer v.objMutex.Unlock()

		if time.Now().After(v.reset) {
			v.reset = time.Now().Add(v.per * time.Millisecond)
			v.objRate = 0
			v.objRatePer = 0
		}

		if v.objRate > v.maxRate || v.objRatePer > v.maxRatePer {
			common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
				"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d", obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer)
			return errors.New("More than the size of the max rate limit")
		}
	} else {
		common.Errorf("fail to find rate limit obj.key=%s", obj)
		return errors.New("fail to find rate limit obj")
	}

	return
}
