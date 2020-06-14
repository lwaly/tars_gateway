package protocol

import (
	"io"
	"net"
)

type (
	//协议接口
	Protocol interface {
		NewCodec(conn net.Conn, timeOut, heartbeat int) (Codec, error)
	}

	Codec interface {
		Receive() (interface{}, interface{}, error, int)
		Send(interface{}) error
		Close() error
		NoticeClose()
	}

	ProtocolFunc func(rw io.ReadWriter) (Codec, error)

	proxyProtocol struct {
		maxPacketSize int
	}
)
