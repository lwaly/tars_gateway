package proxy

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type StTcpProxyConf struct {
	StrObj       string
	Timeout      int    //读写超时时间
	Heartbeat    int    //心跳间隔
	MaxConn      int64  //最大连接数
	StrAddr      string //tcp监听地址
	MaxRate      int64  //最大接收字节数
	MaxRatePer   int64  //每个连接最大接收字节数
	ConnCount    int64  //已连接数
	RateCount    int64  //已接收字节数
	RatePerCount int64  //每个连接已接收字节数
	Per          int64  //限速统计间隔
}

type Controller interface {
	InitProxy()
	TcpProxyGet() (info interface{})
	Verify(info interface{}) error
	HandleReq(info, body, reqTemp interface{}) (output, reqOut interface{}, err error)
	HandleRsp(info, output, reqOut interface{}) (outHeadRsp []byte, err error)
	IsExit(info interface{}) int
	Close(info interface{})
}

func ProxyTcpHandle(session *Session, conn net.Conn, stTcpProxyConf *StTcpProxyConf) {
	//粗略限连接数，不做复杂的使用锁机制精确限连接数
	if 0 < stTcpProxyConf.MaxConn {
		if stTcpProxyConf.ConnCount >= stTcpProxyConf.MaxConn {
			common.Warnf("The max connect is reached.%d %d", stTcpProxyConf.ConnCount, stTcpProxyConf.MaxConn)
			return
		}
		atomic.AddInt64(&stTcpProxyConf.ConnCount, 1)
		defer atomic.AddInt64(&stTcpProxyConf.ConnCount, -1)
	}

	var wg sync.WaitGroup
	infoTcpProxy := session.controller.TcpProxyGet()
	infoProtocol := session.codec.NewCodec(conn, stTcpProxyConf.Timeout, stTcpProxyConf.Heartbeat)
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

		//达到限速阀值,直接丢弃消息,短暂休眠继续消息读取
		if "" != stTcpProxyConf.StrObj {
			if err = util.AddRate(stTcpProxyConf.StrObj, int64(n)); nil != err {
				common.Warnf("More than the size of the max rate limit.%d", n)
				time.Sleep(time.Duration(100) * time.Millisecond)
				continue
			}
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
