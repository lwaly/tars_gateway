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
	flag.Parse()
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

	stHttpProxy, err := proxy.InitHttpProxy("http")
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	stTcpProxy, err := proxy.InitTcpProxy("tcp")
	if nil != err {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	if 1 == stHttpProxy.Switch {
		go proxy.StartHttpProxy(stHttpProxy, &proxy_tars.StTarsHttpProxy{})
	}

	if 1 == stTcpProxy.Switch {
		go proxy.Run("tcp", stTcpProxy, &proxy_tars.ProtoProtocol{}, &proxy_tars.StTarsTcpProxy{})
	}

	common.InstanceGet().Watch()
}
