package proxy

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/lwaly/tars_gateway/common"
)

type StHttpProxyConf struct {
	MaxConn   int64  `json:"maxConn,omitempty"` //最大连接数
	ConnCount int64  //已连接数
	Addr      string `json:"addr,omitempty"` //监听地址
	Per       int64  `json:"per,omitempty"`  //限速统计间隔
}

type HttpController interface {
	ReloadConf() (err error)
	InitProxyHTTP() (err error)
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
	//粗略限连接数，不做复杂的使用锁机制精确限连接数
	if 0 < h.stHttpProxy.MaxConn {
		if h.stHttpProxy.ConnCount >= h.stHttpProxy.MaxConn {
			common.Warnf("The max connect is reached.%d %d", h.stHttpProxy.ConnCount, h.stHttpProxy.MaxConn)
			return
		}
		atomic.AddInt64(&h.stHttpProxy.ConnCount, 1)
		defer atomic.AddInt64(&h.stHttpProxy.ConnCount, -1)
	}

	h.controller.ServeHTTP(w, r)
}
