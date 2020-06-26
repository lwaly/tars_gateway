package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lwaly/tars_gateway/common"
)

type Controller interface {
	InitProxy()
	TcpProxyGet() (info interface{})
	Verify(info interface{}) error
	HandleReq(info, body, reqTemp interface{}) (output, reqOut interface{}, err error)
	HandleRsp(info, output, reqOut interface{}) (outHeadRsp []byte, err error)
	IsExit(info interface{}) int
	Close(info interface{})
}

func InitProxy() {
	//日志系统初始化
	str, err := common.Conf.GetValue("log", "path")
	if nil != err {
		fmt.Printf("fail to get log path")
		return
	}
	defer common.Start(common.LogFilePath(str), common.EveryWeek).Stop()

	level, err := common.Conf.Int("log", "level")
	if nil != err {
		fmt.Printf("fail to get tcp level")
		return
	}

	common.LevelSet(level)
}

func ProxyTcpHandle(session *Session, conn net.Conn) {
	var wg sync.WaitGroup
	infoTcpProxy := session.controller.TcpProxyGet()
	infoProtocol := session.codec.NewCodec(conn, 1, 45)
	defer session.Close(infoProtocol)
	ticker := time.NewTicker(time.Millisecond * 500)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				if 0 != session.controller.IsExit(infoTcpProxy) {
					common.Infof("connect close.")
					session.NoticeClose(infoProtocol)
					return
				}
			}
		}
	}()

	for {
		body, reqTemp, err, n := session.Receive(infoProtocol)
		if 0 != session.controller.IsExit(infoTcpProxy) {
			common.Infof("connect close.n=%d", n)
			session.NoticeClose(infoProtocol)
			break
		}

		if err != nil && "EOF" != err.Error() {
			common.Errorf("fail to get msg.%s", err.Error())
			session.NoticeClose(infoProtocol)
			break
		} else if err != nil && "EOF" == err.Error() {
			common.Infof("msg end.")
			session.NoticeClose(infoProtocol)
			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			output, reqOut, err := session.controller.HandleReq(infoTcpProxy, body, reqTemp)
			if err != nil {
				return
			}
			msg, err := session.controller.HandleRsp(infoTcpProxy, output, reqOut)
			if err != nil {
				common.Infof("fail to handle req msg.%v", err)
				return
			}
			if err = session.Send(infoProtocol, msg); nil != err {
				common.Infof("fail to send msg.%v", err)
				return
			}
		}()
	}

	//关闭代理
	session.controller.Close(infoTcpProxy)
	wg.Wait()
}
