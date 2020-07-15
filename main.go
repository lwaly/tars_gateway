package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/proxy"
	"github.com/lwaly/tars_gateway/proxy/proxy_tars"
)

func main() {
	common.NewConf("./conf/conf")
	proxy.InitProxy()
	stTcpProxy, err := proxy.InitTcpProxy()
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	stHttpProxy, err := proxy.InitHttpProxy()
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
