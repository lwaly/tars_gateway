#! /bin/bash

cd $GOPATH/src/github.com/lwaly/tars_gateway/protocol
protoc --go_out=plugins=tarsrpc:. *.proto
cd $GOPATH/src/github.com/lwaly/tars_gateway/proxy/proxy_tars
protoc --go_out=plugins=tarsrpc:. *.proto