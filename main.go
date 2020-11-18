package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/proxy"
	"github.com/lwaly/tars_gateway/proxy/proxy_tars"
)

func main() {
	var strConfPath string
	flag.StringVar(&strConfPath, "c", "./conf/conf", "The gateway config path")
	if err := common.NewConf(strConfPath); nil != err {
		fmt.Println(err)
		fmt.Println(`
		Usage: gateway [options]
		
		Options:
			-c, <config path>
		`)
		os.Exit(1)
	}
	proxy.InitProxy()

	stHttpProxy, err := proxy.InitHttpProxy()
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	if 1 == stHttpProxy.Switch {
		go proxy.StartContentHttpProxy(stHttpProxy, &proxy_tars.HttpControllerTars{})
	}

	stTcpProxy, err := proxy.InitTcpProxy()
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	if 1 == stTcpProxy.Switch {
		server, err := proxy.Listen("tcp", stTcpProxy, &proxy_tars.ProtoProtocol{}, proxy.HandlerFunc(proxy.ProxyTcpHandle), &proxy_tars.StTarsTcpProxy{})
		if err != nil {
			common.Errorf("%v", err)
			os.Exit(1)
		}

		go server.Serve()
	}

	common.InstanceGet().Watch()
}
