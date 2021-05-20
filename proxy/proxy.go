package proxy

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type Controller interface {
	ReloadConf() (err error)
	InitProxy(key string) error
	TcpProxyGet(ip string) (info interface{})
	Verify(info interface{}) error
	HandlePre(info, reqHead, reqBody interface{}) (limitObj string, reqHeadOut, reqBodyOut interface{}, err error)
	Handle(info, reqHead, reqBody interface{}) (msg []byte, err error)
	IsExit(info interface{}) int
	Close(info interface{})
}

func InitTcpProxy(key string) (stProxyConf *StProxyConf, err error) {
	//tcp连接配置读取
	stProxyConf = new(StProxyConf)
	stProxyConf.key = key
	return stProxyConf, stProxyConf.reloadConf()
}

func ProxyTcpHandle(session *Session) {
	infoProtocol := session.codec.NewCodec(session.conn, session.stProxyConf.Timeout, session.stProxyConf.Heartbeat)
	defer session.Close(infoProtocol)
	if 0 != len(session.stProxyConf.WhiteList) {
		if !common.IpIsInlist(session.conn.RemoteAddr().String(), session.stProxyConf.WhiteList) {
			common.Errorf("addr not in WhiteList.%v", session.conn.RemoteAddr().String())
			return
		}
	} else if 0 != len(session.stProxyConf.BlackList) && common.IpIsInlist(session.conn.RemoteAddr().String(), session.stProxyConf.BlackList) {
		common.Errorf("addr in BlackList.%v", session.conn.RemoteAddr().String())
		return
	}

	if err := util.TcpConnLimitAdd(session.stProxyConf.LimitObj, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.TcpConnLimitAdd(session.stProxyConf.LimitObj, -1)
	var wg sync.WaitGroup
	infoTcpProxy := session.controller.TcpProxyGet(session.conn.RemoteAddr().String())
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

	connId, _ := strconv.ParseInt(fmt.Sprintf("%d", &session.conn), 10, 64)
	pHand := func(isFirstMsg *int, reqHead, reqBody interface{}) (limitObj string, err error) {
		defer wg.Done()
		tempLimitObj, reqHeadTemp, reqBodyTemp, tempErr := session.controller.HandlePre(infoTcpProxy, reqHead, reqBody)
		if tempErr != nil {
			return "", tempErr
		}
		limitObj = tempLimitObj

		temp := fmt.Sprintf("%s.%s", session.stProxyConf.LimitObj, limitObj)
		if 0 == *isFirstMsg {
			if err = util.TcpConnLimitAdd(temp, 1); nil != err {
				common.Warnf("over limit connect.%s", limitObj)
				return
			}
			*isFirstMsg = 1
		}

		msg, tempErr := session.controller.Handle(infoTcpProxy, reqHeadTemp, reqBodyTemp)
		if tempErr != nil {
			common.Infof("fail to handle req msg.%v", err)
			return "", tempErr
		}

		//达到限速阀值,直接丢弃消息
		if "" != session.stProxyConf.LimitObj {
			if err = util.RateAdd(temp, int64(len(msg)), connId); nil != err {
				common.Warnf("More than the size of the max rate limit.%s.%d", temp, len(msg))
				return
			}
		}
		if err = session.Send(infoProtocol, msg); nil != err {
			common.Infof("fail to send msg.%v", err)
			return
		}
		return
	}

	isFirstMsg := 0
	for {
		reqHead, reqBody, err, n := session.Receive(infoProtocol)
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
		if 0 == isFirstMsg {
			if limitObj, err := pHand(&isFirstMsg, reqHead, reqBody); nil != err {
				common.Errorf("connect close.n=%d", n)
				if 1 == isFirstMsg {
					util.TcpConnLimitAdd(limitObj, -1)
				}
				session.NoticeClose(infoProtocol)
				break
			} else if 1 == isFirstMsg {
				defer util.TcpConnLimitAdd(limitObj, -1)
			}
		} else {
			go pHand(&isFirstMsg, reqHead, reqBody)
		}

	}

	//关闭代理
	session.controller.Close(infoTcpProxy)
	wg.Wait()
}
