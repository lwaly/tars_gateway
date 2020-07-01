package proxy

import (
	"net/http"
	"sync/atomic"

	"ohplay_gateway/gateway/common"
)

type StHttpProxyConf struct {
	MaxConn   int64  //最大连接数
	ConnCount int64  //已连接数
	StrAddr   string //监听地址
	Per       int64  //限速统计间隔
}

type HttpController interface {
	InitProxyHTTP() (err error)
	ServeHTTP(w http.ResponseWriter, r *http.Request) (err error)
}

type StHttpController struct {
	controller  HttpController
	stHttpProxy *StHttpProxyConf
}

func StartContentHttpProxy(stHttpProxy *StHttpProxyConf, h HttpController) (err error) {
	//监听端口
	controller := &StHttpController{stHttpProxy: stHttpProxy, controller: h}
	controller.controller.InitProxyHTTP()
	err = http.ListenAndServe(stHttpProxy.StrAddr, controller)

	if err != nil {
		common.Errorf("ListenAndServe:%s ", err.Error())
		return
	}

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
