package proxy

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/protocol"
)

type ReflectServer struct {
	listener     net.Listener
	protocol     protocol.Protocol //数据收发
	handler      Handler
	controller   Controller //业务处理
	sendChSize   int
	timeOut      int
	heartbeat    int
	rateLimitObj string
}

func NewReflectServer(listener net.Listener, protocol protocol.Protocol, handler Handler, controller Controller) (*ReflectServer, error) {
	if nil == controller {
		panic("Controller is nil")
	}

	controller.InitProxy()
	return &ReflectServer{
		listener:   listener,
		protocol:   protocol,
		handler:    handler,
		controller: controller,
	}, nil
}

func (server *ReflectServer) Serve() error {
	for {
		conn, err := Accept(server.listener)
		if err != nil {
			common.Errorf("fail to Accept.%v", err)
			continue
		}

		go server.handleConnection(conn)
	}
}

func (server *ReflectServer) Listener() net.Listener {
	return server.listener
}

func (server *ReflectServer) handleConnection(conn net.Conn) {
	codec, err := server.protocol.NewCodec(conn, 2, 45)
	if err != nil {
		conn.Close()
		return
	}
	if session := NewSession(codec, server.controller); nil != session {
		server.handler.HandleSession(session)
	} else {
		common.Errorf("fail to create session.")
	}
	return
}

func (server *ReflectServer) Stop() {
	server.listener.Close()
}

func Listen(network, address string, protocol protocol.Protocol, handler Handler, controller Controller) (*ReflectServer, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return NewReflectServer(listener, protocol, handler, controller)
}

func Accept(listener net.Listener) (net.Conn, error) {
	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil, io.EOF
			}
			return nil, err
		}
		return conn, nil
	}
}