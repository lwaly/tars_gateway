package util

import (
	"errors"
	"net/http"
	"net/http/httputil"
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

func RateLimitInit(obj string, maxRate, maxRatePer, maxConn, per int64) (err error) {
	if 0 > per || "" == obj {
		common.Warnf("param error.obj=%s,maxRate=%d,maxRatePer=%d,per=%d", obj, maxRate, maxRatePer, per)
		return errors.New("param error")
	}

	if v, ok := mapRateLimit[obj]; ok {
		v.objMutex.Lock()
		common.Infof("update RateLimit.key=%s", obj)
		v.maxConn = maxConn
		v.maxRate = maxRate
		v.maxRatePer = maxRatePer

		if 0 == v.maxConn {
			v.objConn = 0
		}
		if 0 == v.maxRate {
			v.objRate = 0
		}
		if 0 == v.maxRatePer {
			v.objRatePer = 0
			for k := range v.mapobjRatePer {
				delete(v.mapobjRatePer, k)
			}
		}
		v.per = time.Duration(per)
		v.objMutex.Unlock()
		return
	} else {
		if 0 == maxRate && 0 == maxRatePer && 0 == maxConn {
			common.Warnf("param error.obj=%s,maxRate=%d,maxRatePer=%d,per=%d", obj, maxRate, maxRatePer, per)
			return errors.New("param error")
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
	}

	return
}

func rateAdd(ss []string, n, conn int64) (err error) {
	pFunc := func(obj string, n, conn int64) (err error) {
		if v, ok := mapRateLimit[obj]; ok {
			if 0 != v.per {
				v.objMutex.Lock()
				defer v.objMutex.Unlock()

				if time.Now().After(v.reset) {
					v.reset = time.Now().Add(v.per * time.Millisecond)
					v.objRate = 0
					v.objRatePer = 0
					if 0 != conn {
						v.mapobjRatePer[conn] = 0
					}
				}

				//http的每次数据访问都是一次tcp连接，判断当次即可
				if 0 != v.maxRate {
					if v.objRate+n > v.maxRate {
						common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
							"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d, v.maxConn=%d",
							obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer, v.maxConn)
						return errors.New("More than the size of the max rate limit")
					}

					v.objRate += n
				}
				if 0 != v.maxRatePer {
					if 0 == conn {
						if v.objRatePer+n > v.maxRatePer {
							common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
								"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d, v.maxConn=%d",
								obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer, v.maxConn)
							return errors.New("More than the size of the max rate limit")
						}
						v.objRatePer += n
					} else {
						vObjRatePer, _ := v.mapobjRatePer[conn]
						if vObjRatePer+n > v.maxRatePer {
							common.Warnf("More than the size of the max rate limit.obj=%s,v.objRate=%d, "+
								"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d, v.maxConn=%d",
								obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer, v.maxConn)
							return errors.New("More than the size of the max rate limit")
						}
						v.mapobjRatePer[conn] = vObjRatePer + n
					}
				}
			}
		} else {
			common.Infof("unlimit obj.key=%s", obj)
			return
		}

		return
	}

	tempObj := ""
	ssTemp := []string{}
	for _, v := range ss {
		tempObj += v
		if err = pFunc(tempObj, n, conn); nil != err {
			for _, v1 := range ssTemp {
				if errTemp := pFunc(v1, -n, conn); nil != errTemp {
					common.Warnf("More than the size of the max rate limit.%d", n)
					continue
				}
			}
			return
		}
		ssTemp = append(ssTemp, tempObj)
		tempObj += "."
	}
	return
}

func RateHttpAdd(obj string, rsp *http.Response, n int64) (err error) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}

	ss = ss[0:3]
	var b []byte
	if nil != rsp {
		if b, err = httputil.DumpResponse(rsp, false); nil != err {
			common.Errorf("rate limit err.key=%s", obj)
			return
		} else {
			n = rsp.ContentLength + int64(len(b))
		}
	}

	return rateAdd(ss, n, 0)
}

func RateAdd(obj string, n, conn int64) (err error) {
	ss := strings.Split(obj, ".")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}

	return rateAdd(ss, n, conn)
}

func TcpConnLimitAdd(obj string, count int64) (err error) {
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

	return connLimitAdd(ssTemp, count)
}

func HttpConnLimitAdd(obj string, count int64) (err error) {
	ss := strings.Split(obj, "/")
	if 3 > len(ss) {
		common.Errorf("error obj.key=%s", obj)
		return errors.New("error obj")
	}
	ss = ss[0:3]
	return connLimitAdd(ss, count)
}

func connLimitAdd(ss []string, count int64) (err error) {
	pFunc := func(obj string, count int64) (err error) {
		if v, ok := mapRateLimit[obj]; ok {
			if 0 != v.per {
				v.objMutex.Lock()
				defer v.objMutex.Unlock()
				if 0 != v.maxConn {
					if 0 < count {
						if v.objConn+count > v.maxConn {
							common.Warnf("More than the connect of the max rate limit.obj=%s,v.objRate=%d, "+
								"v.maxRate,=%d v.objRatePer=%d, v.maxRatePer=%d, v.maxConn=%d",
								obj, v.objRate, v.maxRate, v.objRatePer, v.maxRatePer, v.maxConn)
							return errors.New("More than the connect of the max rate limit")
						}
					}
					atomic.AddInt64(&v.objConn, count)
				}
			}
		} else {
			common.Infof("unlimit obj.key=%s", obj)
			return
		}

		return
	}

	tempObj := ""
	ssTemp := []string{}
	for _, v := range ss {
		tempObj += v
		if err = pFunc(tempObj, count); nil != err {
			for _, v1 := range ssTemp {
				if errTemp := pFunc(v1, -count); nil != errTemp {
					common.Warnf("More than the connect of the max rate limit.%d", count)
					continue
				}
			}
			return
		}
		ssTemp = append(ssTemp, tempObj)
		tempObj += "."
	}
	return
}
