package proxy

import (
	"net/http"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type StHttpProxyConf struct {
	Addr            string          `json:"addr,omitempty"`            //监听地址
	LimitObj        string          `json:"limitObj,omitempty"`        //监听地址
	Switch          uint32          `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch uint32          `json:"rateLimitSwitch,omitempty"` //1开启服务
	MaxConn         int64           `json:"maxConn,omitempty"`         //最大连接数
	MaxRate         int64           `json:"maxRate,omitempty"`         //最大接收字节数
	MaxRatePer      int64           `json:"maxRatePer,omitempty"`      //每个连接最大接收字节数
	ConnCount       int64           `json:"connCount,omitempty"`       //已连接数
	RateCount       int64           `json:"rateCount,omitempty"`       //已接收字节数
	RatePerCount    int64           `json:"ratePerCount,omitempty"`    //每个连接已接收字节数
	Per             int64           `json:"per,omitempty"`             //限速统计间隔
	App             []StHttpAppConf `json:"app,omitempty"`             //限速统计间隔
}

type StHttpAppConf struct {
	Switch          uint32                `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch uint32                `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name            string                `json:"name,omitempty"`
	MaxConn         int64                 `json:"maxConn,omitempty"`      //最大连接数
	MaxRate         int64                 `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer      int64                 `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount       int64                 `json:"connCount,omitempty"`    //已连接数
	RateCount       int64                 `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount    int64                 `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per             int64                 `json:"per,omitempty"`          //限速统计间隔
	Server          []StHttpAppServerConf `json:"server,omitempty"`       //限速统计间隔
}

type StHttpAppServerConf struct {
	Switch          uint32 `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch uint32 `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name            string `json:"name,omitempty"`
	MaxConn         int64  `json:"maxConn,omitempty"`      //最大连接数
	MaxRate         int64  `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer      int64  `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount       int64  `json:"connCount,omitempty"`    //已连接数
	RateCount       int64  `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount    int64  `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per             int64  `json:"per,omitempty"`          //限速统计间隔
}

type HttpController interface {
	ReloadConf() (err error)
	InitProxyHTTP(f func(*http.Response)) (err error)
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
	}
	return
}

func StartContentHttpProxy(stHttpProxy *StHttpProxyConf, h HttpController) (err error) {
	//监听端口
	controller := &StHttpController{stHttpProxy: stHttpProxy, controller: h}
	controller.controller.InitProxyHTTP()
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
	if err := util.AddHttpConnLimit("/"+h.stHttpProxy.LimitObj+r.URL.Path, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.AddHttpConnLimit("/"+h.stHttpProxy.LimitObj+r.URL.Path, -1)

	h.controller.ServeHTTP(w, r)
}

func ModifyResponse(rsp *http.Response) error {
	return nil
}
