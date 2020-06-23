package protocol

import (
	"net"
)

type (
	//协议接口
	Protocol interface {
		NewCodec(conn net.Conn, timeOut, heartbeat int) interface{}
		Receive(interface{}) (interface{}, interface{}, error, int)
		Send(interface{}, interface{}) error
		Close(interface{}) error
		NoticeClose(interface{})
	}
)
