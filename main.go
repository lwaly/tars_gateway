package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/proxy"
	"github.com/lwaly/tars_gateway/proxy/proxy_tars"
)

func main() {
	proxy.InitProxy()
	str, err := common.Conf.GetValue("tcp", "addr")
	if nil != err {
		fmt.Printf("fail to get log path")
		return
	}
	server, err := proxy.Listen("tcp", str, &proxy_tars.ProtoProtocol{}, proxy.HandlerFunc(proxy.ProxyTcpHandle), &proxy_tars.StTarsTcpProxy{})
	if err != nil {
		common.Errorf("%v", err)
		os.Exit(1)
	}

	server.Serve()
}
