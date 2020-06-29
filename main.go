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
	server, err := proxy.Listen(&proxy_tars.ProtoProtocol{}, proxy.HandlerFunc(proxy.ProxyTcpHandle), &proxy_tars.StTarsTcpProxy{})
	if err != nil {
		common.Errorf("%v", err)
		os.Exit(1)
	}

	server.Serve()
}
