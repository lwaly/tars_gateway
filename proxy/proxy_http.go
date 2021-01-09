package proxy

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type StHttpProxyConf struct {
	Addr                     string          `json:"addr,omitempty"`            //监听地址
	LimitObj                 string          `json:"limitObj,omitempty"`        //http对象
	Switch                   uint32          `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch          uint32          `json:"rateLimitSwitch,omitempty"` //1开启服务
	MaxConn                  int64           `json:"maxConn,omitempty"`         //最大连接数
	MaxRate                  int64           `json:"maxRate,omitempty"`         //最大接收字节数
	MaxRatePer               int64           `json:"maxRatePer,omitempty"`      //每个连接最大接收字节数
	ConnCount                int64           `json:"connCount,omitempty"`       //已连接数
	RateCount                int64           `json:"rateCount,omitempty"`       //已接收字节数
	RatePerCount             int64           `json:"ratePerCount,omitempty"`    //每个连接已接收字节数
	Per                      int64           `json:"per,omitempty"`             //限速统计间隔
	App                      []StHttpAppConf `json:"app,omitempty"`             //限速统计间隔
	BlackList                []string        `json:"blackList,omitempty"`       //
	WhiteList                []string        `json:"whiteList,omitempty"`       //
	CacheSwitch              int64           `json:"cacheSwitch,omitempty"`
	CacheSize                int64           `json:"cacheSize,omitempty"`
	CacheExpirationTime      int64           `json:"cacheExpirationTime,omitempty"`
	CacheExpirationCleanTime string          `json:"cacheExpirationCleanTime,omitempty"`
}

type StHttpAppConf struct {
	Switch                   uint32                `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch          uint32                `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name                     string                `json:"name,omitempty"`
	MaxConn                  int64                 `json:"maxConn,omitempty"`      //最大连接数
	MaxRate                  int64                 `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer               int64                 `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount                int64                 `json:"connCount,omitempty"`    //已连接数
	RateCount                int64                 `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount             int64                 `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per                      int64                 `json:"per,omitempty"`          //限速统计间隔
	Server                   []StHttpAppServerConf `json:"server,omitempty"`       //限速统计间隔
	BlackList                []string              `json:"blackList,omitempty"`    //
	WhiteList                []string              `json:"whiteList,omitempty"`    //
	CacheSwitch              int64                 `json:"cacheSwitch,omitempty"`
	CacheSize                int64                 `json:"cacheSize,omitempty"`
	CacheExpirationTime      int64                 `json:"cacheExpirationTime,omitempty"`
	CacheExpirationCleanTime string                `json:"cacheExpirationCleanTime,omitempty"`
}

type StHttpAppServerConf struct {
	Switch                   uint32   `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch          uint32   `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name                     string   `json:"name,omitempty"`
	MaxConn                  int64    `json:"maxConn,omitempty"`      //最大连接数
	MaxRate                  int64    `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer               int64    `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount                int64    `json:"connCount,omitempty"`    //已连接数
	RateCount                int64    `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount             int64    `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per                      int64    `json:"per,omitempty"`          //限速统计间隔
	BlackList                []string `json:"blackList,omitempty"`    //
	WhiteList                []string `json:"whiteList,omitempty"`    //
	CacheSwitch              int64    `json:"cacheSwitch,omitempty"`
	CacheSize                int64    `json:"cacheSize,omitempty"`
	CacheExpirationTime      int64    `json:"cacheExpirationTime,omitempty"`
	CacheExpirationCleanTime string   `json:"cacheExpirationCleanTime,omitempty"`
}

type HttpController interface {
	ReloadConf() (err error)
	InitProxyHTTP(p interface{}, ResponseFunc func(p interface{}, rsp *http.Response) error,
		RequestFunc func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)) (err error)
	ServeHTTP(w http.ResponseWriter, r *http.Request) (err error)
}

type StHttpController struct {
	controller  HttpController
	stHttpProxy *StHttpProxyConf
}

func reloadHttpConf(stHttpProxy *StHttpProxyConf) (err error) {
	err = common.Conf.GetStruct("http", stHttpProxy)
	if err != nil {
		common.Errorf("fail to get http conf.%v", err)
		return
	}

	if common.SWITCH_ON == stHttpProxy.RateLimitSwitch {
		util.RateLimitInit(stHttpProxy.LimitObj, stHttpProxy.MaxRate, stHttpProxy.MaxRatePer, stHttpProxy.MaxConn, stHttpProxy.Per)
	} else if "" != stHttpProxy.LimitObj {
		util.RateLimitInit(stHttpProxy.LimitObj, 0, 0, 0, 0)
	}
	for _, v := range stHttpProxy.App {
		if common.SWITCH_ON == v.RateLimitSwitch {
			util.RateLimitInit(stHttpProxy.LimitObj+"."+v.Name, v.MaxRate, v.MaxRatePer, v.MaxConn, v.Per)
		} else if "" != stHttpProxy.LimitObj {
			util.RateLimitInit(stHttpProxy.LimitObj+"."+v.Name, 0, 0, 0, 0)
		}
		for _, v1 := range v.Server {
			if common.SWITCH_ON == v.RateLimitSwitch {
				util.RateLimitInit(stHttpProxy.LimitObj+"."+v.Name+"."+v1.Name, v1.MaxRate, v1.MaxRatePer, v1.MaxConn, v1.Per)
			} else if "" != stHttpProxy.LimitObj {
				util.RateLimitInit(stHttpProxy.LimitObj+"."+v.Name+"."+v1.Name, 0, 0, 0, 0)
			}
		}
	}

	if common.SWITCH_ON == stHttpProxy.CacheSwitch {
		util.InitCache(stHttpProxy.LimitObj, stHttpProxy.CacheExpirationCleanTime,
			time.Duration(stHttpProxy.CacheExpirationTime)*time.Millisecond, stHttpProxy.CacheSize)
	} else if "" != stHttpProxy.LimitObj {
		util.InitCache(stHttpProxy.LimitObj, "", 0, 0)
	}
	for _, v := range stHttpProxy.App {
		if common.SWITCH_ON == v.CacheSwitch {
			util.InitCache(stHttpProxy.LimitObj+"."+v.Name, v.CacheExpirationCleanTime,
				time.Duration(v.CacheExpirationTime)*time.Millisecond, v.CacheSize)
		} else if "" != stHttpProxy.LimitObj {
			util.InitCache(stHttpProxy.LimitObj+"."+v.Name, "", 0, 0)
		}
		for _, v1 := range v.Server {
			if common.SWITCH_ON == v.CacheSwitch {
				util.InitCache(stHttpProxy.LimitObj+"."+v.Name+"."+v1.Name, v1.CacheExpirationCleanTime,
					time.Duration(v1.CacheExpirationTime)*time.Millisecond, v1.CacheSize)
			} else if "" != stHttpProxy.LimitObj {
				util.InitCache(stHttpProxy.LimitObj+"."+v.Name+"."+v1.Name, "", 0, 0)
			}
		}
	}

	return
}

func InitHttpProxy() (stHttpProxy *StHttpProxyConf, err error) {
	stHttpProxy = new(StHttpProxyConf)
	return stHttpProxy, reloadHttpConf(stHttpProxy)
}

func StartHttpProxy(stHttpProxy *StHttpProxyConf, h HttpController) (err error) {
	//监听端口

	controller := &StHttpController{stHttpProxy: stHttpProxy, controller: h}
	controller.controller.InitProxyHTTP(stHttpProxy, ModifyResponse, ModifyRequest)
	err = http.ListenAndServe(stHttpProxy.Addr, controller)

	if err != nil {
		common.Errorf("ListenAndServe:%s ", err.Error())
		return
	}

	ticker := time.NewTicker(time.Second * 5)
	confLastUpdateTime := time.Now().UnixNano()
	go func() {
		for {
			select {
			case <-ticker.C:
				if confLastUpdateTime < common.Conf.LastUpdateTimeGet() {
					confLastUpdateTime = common.Conf.LastUpdateTimeGet()
					reloadHttpConf(stHttpProxy)
					controller.controller.ReloadConf()
				}
			}
		}
	}()

	return err
}

//实现Handler的接口
func (h *StHttpController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if 0 != len(h.stHttpProxy.WhiteList) {
		if !common.IpIsInlist(r.RemoteAddr, h.stHttpProxy.WhiteList) {
			common.Errorf("addr not in WhiteList.%v", r.RemoteAddr)
			return
		}
	} else if 0 != len(h.stHttpProxy.BlackList) && common.IpIsInlist(r.RemoteAddr, h.stHttpProxy.BlackList) {
		common.Errorf("addr in BlackList.%v", r.RemoteAddr)
		return
	}

	if err := util.HttpConnLimitAdd(h.stHttpProxy.LimitObj+r.URL.Path, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.HttpConnLimitAdd(h.stHttpProxy.LimitObj+r.URL.Path, -1)

	h.controller.ServeHTTP(w, r)
}

func ModifyResponse(p interface{}, rsp *http.Response) (err error) {
	stHttpProxy := p.(*StHttpProxyConf)
	if nil != stHttpProxy {
		//计算是否可缓存，然后在最终得到限速的结果返回，限速不代表不能缓存
		cacheObj := rsp.Request.Header.Get("TARS_CACHE_OBJ")
		cacheKey := rsp.Request.Header.Get("TARS_CACHE_KEY")
		if "" != cacheObj && "" != cacheKey && util.CacheHttpObjExist(cacheObj) {
			defer rsp.Body.Close()
			if body, err := ioutil.ReadAll(rsp.Body); nil == err {
				cacheKeyHead := cacheKey + "head"
				cacheKeyBody := cacheKey + "body"
				rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

				if err = util.CacheHttpHeadAdd(cacheObj, cacheKeyHead, &rsp.Header); nil != err {
					common.Errorf("%v.%v", err, string(body))
				} else {
					if err = util.CacheHttpBodyAdd(cacheObj, cacheKeyBody, body); nil != err {
						common.Errorf("%v.%v", err, string(body))
					}
				}
			} else {
				common.Errorf("fail to read body.%v", err)
			}
		}

		temp := fmt.Sprintf("%s%s", stHttpProxy.LimitObj, rsp.Request.URL.Path)

		if err = util.RateHttpAdd(temp, rsp, 0); nil != err {
			common.Warnf("More than the size of the max rate limit.%s", temp)
			return
		}
	}

	return nil
}

func ModifyRequest(p interface{}, w http.ResponseWriter, r *http.Request) (code int, err error) {
	stHttpProxy := p.(*StHttpProxyConf)

	if nil != stHttpProxy {
		cacheObj := fmt.Sprintf("%s%s", stHttpProxy.LimitObj, r.URL.Path)

		//查询cache
		if "1" == r.Header.Get("TARS_CACHE") && util.CacheHttpObjExist(cacheObj) {
			defer r.Body.Close()
			if body, err := ioutil.ReadAll(r.Body); nil == err {
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				ha := md5.New()
				ha.Write([]byte(fmt.Sprintf("%s%s%v", r.Method, r.URL.Path, body)))
				cacheKey := base64.StdEncoding.EncodeToString(ha.Sum(nil))
				r.Header.Add("TARS_CACHE_KEY", cacheKey)
				r.Header.Add("TARS_CACHE_OBJ", cacheObj)
				cacheKeyHead := cacheKey + "head"
				cacheKeyBody := cacheKey + "body"
				if err, head, n := util.CacheHttpHeadGet(cacheObj, cacheKeyHead); nil == err {
					if err, v := util.CacheHttpBodyGet(cacheObj, cacheKeyBody); nil == err {
						body, okbody := v.([]byte)
						if okbody {
							temp := fmt.Sprintf("%s%s", stHttpProxy.LimitObj, r.URL.Path)

							if err = util.RateHttpAdd(temp, nil, n+int64(len(body))); nil != err {
								common.Warnf("More than the size of the max rate limit.%s", temp)
								return common.ERR_LIMIT, err
							}
							if _, err = w.Write(body); nil == err {
								copyHeader(w.Header(), head)
								return common.OK, errors.New("CACHE")
							} else {
								common.Errorf("fail to write body.%v", err)
								return common.OK, nil
							}
						}
					}
				}
			}
		}
	}

	return common.OK, nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
