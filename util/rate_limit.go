package util

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lwaly/tars_gateway/common"
)

type stRateLimit struct {
	maxConn       int64
	maxRate       int64
	maxRatePer    int64
	objRate       int64
	objRatePer    int64
	objConn       int64
	mapobjRatePer map[int64]int64
	per           time.Duration
	reset         time.Time
	objMutex      *sync.Mutex
}

var mapRateLimit map[string]*stRateLimit

func init() {
	mapRateLimit = make(map[string]*stRateLimit)
}

func InitRateLimit(obj string, maxRate, maxRatePer, maxConn, per int64) (err error) {
	if 0 >= per || "" == obj {
		common.Warnf("param error.obj=%s,maxRate=%d,maxRatePer=%d,per=%d", obj, maxRate, maxRatePer, per)
		return errors.New("param error")
	}

	if _, ok := mapRateLimit[obj]; ok {
		common.Errorf("repeat key.key=%s", obj)
		return errors.New("repeat key")
	}

	lstRateLimit := stRateLimit{}
	lstRateLimit.maxConn = maxConn
	lstRateLimit.maxRate = maxRate
	lstRateLimit.maxRatePer = maxRatePer
	lstRateLimit.objRate = 0
	lstRateLimit.objRatePer = 0
	lstRateLimit.objConn = 0
	lstRateLimit.per = time.Duration(per)
	lstRateLimit.reset = time.Now().Add(lstRateLimit.per * time.Millisecond)
	lstRateLimit.objMutex = new(sync.Mutex)
	lstRateLimit.mapobjRatePer = make(map[int64]int64)

	mapRateLimit[obj] = &lstRateLimit
	return
}

func addRate(obj string, n, conn int64) (err error) {
	if v, ok := mapRateLimit[obj]; ok {
		v.objMutex.Lock()
		defer v.objMutex.Unlock()

		if time.Now().After(v.reset) {
			v.reset = time.Now().Add(v.per * time.Millisecond)
			v.objRate = 0
			v.objRatePer = 0
		}

		//http的每次数据访问都是一次tcp连接，判断当次即可
		if 0 == conn {
			if v.objRate+n > v.maxRate || v.objRatePer+n > v.maxRatePer {
				common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
					"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d", obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer)
				return errors.New("More than the size of the max rate limit")
			}
			v.objRatePer += n
		} else {
			vObjRatePer, _ := v.mapobjRatePer[conn]
			if v.objRate+n > v.maxRate || vObjRatePer+n > v.maxRatePer {
				common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
					"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d", obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer)
				return errors.New("More than the size of the max rate limit")
			}
			v.mapobjRatePer[conn] = vObjRatePer + n
		}

		v.objRate += n
	} else {
		common.Infof("unlimit obj.key=%s", obj)
		return
	}

	return
}

func AddRate(obj string, n, conn int64) (err error) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}

	ss = ss[0:3]
	tempObj := ""
	for _, v := range ss {
		tempObj += v
		if err = addRate(tempObj, n, conn); nil != err {
			common.Warnf("More than the size of the max rate limit.%d", n)
			return
		}
		tempObj += "."
	}
	return
}

func AddTcpConnLimit(obj string, count int64) (err error) {
	ss := strings.Split(obj, ".")
	ssTemp := []string{}
	if 1 == len(ss) {
		ssTemp = append(ssTemp, ss[0])
	} else if 3 <= len(ss) {
		ssTemp = append(ssTemp, ss[0]+"."+ss[1])
		ssTemp = append(ssTemp, ss[0]+"."+ss[1]+"."+ss[2])
	} else {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}

	for _, v := range ssTemp {
		if err = addConnLimit(v, count); nil != err {
			common.Warnf("More than the size of the max rate limit.%d", n)
			return
		}
	}
	return
}

func AddHttpConnLimit(obj string, count int64) (err error) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}

	ss = ss[0:3]
	tempObj := ""
	for _, v := range ss {
		tempObj += v
		if err = addConnLimit(tempObj, count); nil != err {
			common.Warnf("More than the size of the max rate limit.%d", n)
			return
		}
		tempObj += "."
	}
	return
}

func addConnLimit(obj string, count int64) (err error) {
	//粗略限连接数，不做复杂的使用锁机制精确限连接数
	if v, ok := mapRateLimit[obj]; ok {
		v.objMutex.Lock()
		defer v.objMutex.Unlock()

		if 0 < count {
			if v.objConn > v.maxConn {
				common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
					"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d", obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer)
				return errors.New("More than the size of the max rate limit")
			}
		}
		atomic.AddInt64(&v.objConn, count)
	} else {
		common.Infof("unlimit obj.key=%s", obj)
		return
	}

	return
}
