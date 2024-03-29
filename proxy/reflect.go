package proxy

import (
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/protocol"
)

type ReflectServer struct {
	listener    net.Listener
	protocol    protocol.Protocol //数据收发
	handler     Handler
	controller  Controller //业务处理
	stProxyConf *StProxyConf
}

func (server *ReflectServer) Serve() (err error) {
	common.Infof("start tcp server.addr=%s", server.stProxyConf.Addr)

	if err = server.controller.InitProxy(server.stProxyConf.key); nil != err {
		common.Errorf("fail init tcp server.addr=%s", server.stProxyConf.Addr)
		return
	}
	server.reloadConf()

	for {
		conn, err := Accept(server.listener)
		if err != nil {
			common.Errorf("fail to Accept.%v", err)
			continue
		}

		go server.handleConnection(conn)
	}
}

func (server *ReflectServer) reloadConf() {
	//tcp连接配置读取
	ticker := time.NewTicker(time.Second * 5)
	confLastUpdateTime := time.Now().UnixNano()
	go func() {
		for {
			select {
			case <-ticker.C:
				if confLastUpdateTime < common.Conf.LastUpdateTimeGet() {
					common.Infof("config update")
					confLastUpdateTime = common.Conf.LastUpdateTimeGet()
					server.stProxyConf.reloadConf()
					server.controller.ReloadConf()
				}
			}
		}
	}()

	return
}

func (server *ReflectServer) Listener() net.Listener {
	return server.listener
}

func (server *ReflectServer) handleConnection(conn net.Conn) {
	if session := NewSession(server.protocol, server.controller, conn, server.stProxyConf); nil != session {
		server.handler.HandleSession(session)
	} else {
		common.Errorf("fail to create session.")
	}
	return
}

func (server *ReflectServer) Stop() {
	server.listener.Close()
}

func Run(network string, stProxyConf *StProxyConf, protocol protocol.Protocol, controller Controller) (err error) {
	if nil == controller || nil == stProxyConf {
		return errors.New("Controller or stProxyConf is nil")
	}

	listener, err := net.Listen(network, stProxyConf.Addr)
	if err != nil {
		return err
	}

	server := &ReflectServer{
		listener:    listener,
		protocol:    protocol,
		handler:     HandlerFunc(ProxyTcpHandle),
		controller:  controller,
		stProxyConf: stProxyConf,
	}

	return server.Serve()
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
