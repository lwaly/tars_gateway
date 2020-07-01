package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/proxy"
	"github.com/lwaly/tars_gateway/proxy/proxy_tars"
)

func main() {
	proxy.InitProxy()
	stTcpProxy, err := InitTarsTcpConf()
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	stHttpProxy, err := InitTarsHttpConf()
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	server, err := proxy.Listen("tcp", stTcpProxy, &proxy_tars.ProtoProtocol{}, proxy.HandlerFunc(proxy.ProxyTcpHandle), &proxy_tars.StTarsTcpProxy{})
	if err != nil {
		common.Errorf("%v", err)
		os.Exit(1)
	}

	go proxy.StartContentHttpProxy(stHttpProxy, &proxy_tars.HttpControllerTars{})
	server.Serve()
}

func InitTarsTcpConf() (stTcpProxy *proxy.StTcpProxyConf, err error) {
	//tcp连接配置读取
	stTcpProxy = new(proxy.StTcpProxyConf)
	stTcpProxy.Timeout, err = common.Conf.Int("tcp", "timeout")
	if nil != err {
		common.Errorf("fail to get tcp timeout.%v", err)
		return
	}

	stTcpProxy.MaxConn, err = common.Conf.Int64("tcp", "max_conn")
	if nil != err {
		common.Errorf("fail to get tcp maxConn.%v", err)
		return
	}

	stTcpProxy.MaxRate, err = common.Conf.Int64("tcp", "rate")
	if nil != err {
		common.Errorf("fail to get tcp rate.%v", err)
		return
	}

	stTcpProxy.MaxRatePer, err = common.Conf.Int64("tcp", "rate_per")
	if nil != err {
		common.Errorf("fail to get tcp rate_per.%v", err)
		return
	}

	stTcpProxy.Heartbeat, err = common.Conf.Int("tcp", "heartbeat")
	if nil != err {
		common.Errorf("fail to get tcp heartbeat.%v", err)
		return
	}

	stTcpProxy.Per, err = common.Conf.Int64("tcp", "per")
	if nil != err {
		common.Errorf("fail to get tcp per.%v", err)
		return
	}

	stTcpProxy.StrAddr, err = common.Conf.GetValue("tcp", "addr")
	if nil != err {
		common.Errorf("fail to get tcp addr.%v", err)
		return
	}

	stTcpProxy.StrObj, err = common.Conf.GetValue("tcp", "obj")
	if nil != err {
		common.Errorf("fail to get tcp addr.%v", err)
		return
	}
	return
}

func InitTarsHttpConf() (stHttpProxy *proxy.StHttpProxyConf, err error) {
	//tcp连接配置读取
	stHttpProxy = new(proxy.StHttpProxyConf)

	stHttpProxy.MaxConn, err = common.Conf.Int64("http", "max_conn")
	if nil != err {
		common.Errorf("fail to get http maxConn.%v", err)
		return
	}

	stHttpProxy.Per, err = common.Conf.Int64("http", "per")
	if nil != err {
		common.Errorf("fail to get http per.%v", err)
		return
	}

	stHttpProxy.StrAddr, err = common.Conf.GetValue("http", "addr")
	if nil != err {
		common.Errorf("fail to get http addr.%v", err)
		return
	}
	return
}
