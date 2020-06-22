package proxy

import (
	"fmt"
	"sync"

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

func ProxyTcpHandle(session *Session) {
	defer session.Close()
	var wg sync.WaitGroup
	var uid int64
	info := session.controller.TcpProxyGet()
	for {
		body, reqTemp, err, n := session.Receive()
		if 0 != session.controller.IsExit(info) {
			common.Infof("connect close.uid=%d.len=%d", uid, n)
			session.NoticeClose()
			break
		}
		if err != nil && "EOF" != err.Error() {
			common.Errorf("fail to get msg.%s", err.Error())
			break
		} else if err != nil && "EOF" == err.Error() {
			common.Infof("msg end.")
			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			output, reqOut, err := session.controller.HandleReq(info, body, reqTemp)
			if err != nil {
				return
			}
			msg, err := session.controller.HandleRsp(info, output, reqOut)
			if err != nil {
				common.Infof("fail to handle req msg.%v", err)
				return
			}
			if err = session.Send(msg); nil != err {
				common.Infof("fail to send msg.%v", err)
				return
			}
		}()
	}

	//关闭代理
	session.controller.Close(info)
	wg.Wait()
}
