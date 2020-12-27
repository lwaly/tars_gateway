package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type StHttpProxyConf struct {
	Addr                string          `json:"addr,omitempty"`            //监听地址
	LimitObj            string          `json:"limitObj,omitempty"`        //http对象
	Switch              uint32          `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch     uint32          `json:"rateLimitSwitch,omitempty"` //1开启服务
	MaxConn             int64           `json:"maxConn,omitempty"`         //最大连接数
	MaxRate             int64           `json:"maxRate,omitempty"`         //最大接收字节数
	MaxRatePer          int64           `json:"maxRatePer,omitempty"`      //每个连接最大接收字节数
	ConnCount           int64           `json:"connCount,omitempty"`       //已连接数
	RateCount           int64           `json:"rateCount,omitempty"`       //已接收字节数
	RatePerCount        int64           `json:"ratePerCount,omitempty"`    //每个连接已接收字节数
	Per                 int64           `json:"per,omitempty"`             //限速统计间隔
	App                 []StHttpAppConf `json:"app,omitempty"`             //限速统计间隔
	BlackList           []string        `json:"blackList,omitempty"`       //
	WhiteList           []string        `json:"whiteList,omitempty"`       //
	CacheSwitch         int64           `json:"cacheSwitch,omitempty"`
	CacheSize           int64           `json:"cacheSize,omitempty"`
	CacheExpirationTime int64           `json:"cacheExpirationTime,omitempty"`
}

type StHttpAppConf struct {
	Switch              uint32                `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch     uint32                `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name                string                `json:"name,omitempty"`
	MaxConn             int64                 `json:"maxConn,omitempty"`      //最大连接数
	MaxRate             int64                 `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer          int64                 `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount           int64                 `json:"connCount,omitempty"`    //已连接数
	RateCount           int64                 `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount        int64                 `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per                 int64                 `json:"per,omitempty"`          //限速统计间隔
	Server              []StHttpAppServerConf `json:"server,omitempty"`       //限速统计间隔
	BlackList           []string              `json:"blackList,omitempty"`    //
	WhiteList           []string              `json:"whiteList,omitempty"`    //
	CacheSwitch         int64                 `json:"cacheSwitch,omitempty"`
	CacheSize           int64                 `json:"cacheSize,omitempty"`
	CacheExpirationTime int64                 `json:"cacheExpirationTime,omitempty"`
}

type StHttpAppServerConf struct {
	Switch              uint32   `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch     uint32   `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name                string   `json:"name,omitempty"`
	MaxConn             int64    `json:"maxConn,omitempty"`      //最大连接数
	MaxRate             int64    `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer          int64    `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount           int64    `json:"connCount,omitempty"`    //已连接数
	RateCount           int64    `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount        int64    `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per                 int64    `json:"per,omitempty"`          //限速统计间隔
	BlackList           []string `json:"blackList,omitempty"`    //
	WhiteList           []string `json:"whiteList,omitempty"`    //
	CacheSwitch         int64    `json:"cacheSwitch,omitempty"`
	CacheSize           int64    `json:"cacheSize,omitempty"`
	CacheExpirationTime int64    `json:"cacheExpirationTime,omitempty"`
}

type HttpController interface {
	ReloadConf() (err error)
	InitProxyHTTP(p interface{}, f func(interface{}, *http.Request, *http.Response) error) (err error)
	ServeHTTP(w http.ResponseWriter, r *http.Request) (err error)
}

type StHttpController struct {
	controller  HttpController
	stHttpProxy *StHttpProxyConf
}

func InitHttpProxy() (stHttpProxy *StHttpProxyConf, err error) {
	//tcp连接配置读取
	stHttpProxy = new(StHttpProxyConf)

	err = common.Conf.GetStruct("http", stHttpProxy)
	if err != nil {
		common.Errorf("fail to get http conf.%v", err)
		return
	}

	if common.SWITCH_ON == stHttpProxy.Switch {
		if common.SWITCH_ON == stHttpProxy.RateLimitSwitch {
			util.InitRateLimit(stHttpProxy.LimitObj, stHttpProxy.MaxRate, stHttpProxy.MaxRatePer, stHttpProxy.MaxConn, stHttpProxy.Per)
		}
		for _, v := range stHttpProxy.App {
			if common.SWITCH_ON == v.RateLimitSwitch {
				util.InitRateLimit(stHttpProxy.LimitObj+"."+v.Name, v.MaxRate, v.MaxRatePer, v.MaxConn, v.Per)
			}
			for _, v1 := range v.Server {
				if common.SWITCH_ON == v.RateLimitSwitch {
					util.InitRateLimit(stHttpProxy.LimitObj+"."+v.Name+"."+v1.Name, v1.MaxRate, v1.MaxRatePer, v1.MaxConn, v1.Per)
				}
			}
		}

		if common.SWITCH_ON == stHttpProxy.CacheSwitch {
			util.InitCache(stHttpProxy.LimitObj, time.Duration(stHttpProxy.CacheExpirationTime)*time.Millisecond,
				time.Duration(60)*time.Second, stHttpProxy.CacheSize)
		}
		for _, v := range stHttpProxy.App {
			if common.SWITCH_ON == v.CacheSwitch {
				util.InitCache(stHttpProxy.LimitObj+"."+v.Name,
					time.Duration(v.CacheExpirationTime)*time.Millisecond,
					time.Duration(60)*time.Second, v.CacheSize)
			}
			for _, v1 := range v.Server {
				if common.SWITCH_ON == v.CacheSwitch {
					util.InitCache(stHttpProxy.LimitObj+"."+v.Name+"."+v1.Name,
						time.Duration(v1.CacheExpirationTime)*time.Millisecond,
						time.Duration(60)*time.Second, v1.CacheSize)
				}
			}
		}
	}

	return
}

func StartContentHttpProxy(stHttpProxy *StHttpProxyConf, h HttpController) (err error) {
	//监听端口
	controller := &StHttpController{stHttpProxy: stHttpProxy, controller: h}
	controller.controller.InitProxyHTTP(stHttpProxy, ModifyResponse)
	err = http.ListenAndServe(stHttpProxy.Addr, controller)

	if err != nil {
		common.Errorf("ListenAndServe:%s ", err.Error())
		return
	}

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			select {
			case <-ticker.C:
				if nil != stHttpProxy {
					err := common.Conf.GetStruct("http", stHttpProxy)
					if err != nil {
						common.Errorf("fail to get http conf.%v", err)
					}
				}
				if nil != controller.controller {
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

	if err := util.AddHttpConnLimit(h.stHttpProxy.LimitObj+r.URL.Path, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.AddHttpConnLimit(h.stHttpProxy.LimitObj+r.URL.Path, -1)

	h.controller.ServeHTTP(w, r)
}

func ModifyResponse(p interface{}, req *http.Request, rsp *http.Response) (err error) {
	stHttpProxy := p.(*StHttpProxyConf)
	if nil != stHttpProxy {
		temp := fmt.Sprintf("%s%s", stHttpProxy.LimitObj, rsp.Request.URL.Path)
		n := rsp.ContentLength
		if rsp.ContentLength < 0 {
			n = 0
		}
		for k, v := range rsp.Header {
			n += int64(len(k)) + int64(len(v))
		}
		if err = util.AddHttpRate(temp, n, 0); nil != err {
			common.Warnf("More than the size of the max rate limit.%s.%d", temp, n)
			return
		}
	}

	return nil
}
