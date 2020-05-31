#! /bin/bash

cd $GOPATH/src/github.com/lwaly/tars_gateway/protocol
protoc --go_out=plugins=tarsrpc:. *.proto
