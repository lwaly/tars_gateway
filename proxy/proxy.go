package proxy

import (
	"fmt"
	"sync"

	"github.com/lwaly/tars_gateway/common"
)

type Controller interface {
	InitProxy()
	Verify() error
	HandleReq(body, reqTemp interface{}) (output, reqOut interface{}, err error)
	HandleRsp(output, reqOut interface{}) (outHeadRsp []byte, err error)
	IsExit() int
	Close()
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
	//var err error
	var uid int64
	// var server uint32
	// var servant uint32
	// isret := 0

	// if err = session.controller.Verify(); nil != err {
	// 	common.Errorf("error verify user.%d", uid)
	// 	isret = 2
	// 	return
	// }

	// ticker := time.NewTicker(time.Millisecond * 500)

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	tempV, ok := mapUser.Load(uid)
	// 	if !ok {
	// 		common.Infof("fail to find user.%d", uid)
	// 		return
	// 	}
	// 	v, ok := tempV.(*stUserInfo)
	// 	if ok {
	// 		//var seq uint32
	// 		for {
	// 			select {
	// 			case tokenRet = <-v.reader:
	// 				isret = 1
	// 				msg,err:=session.controller.Notice()
	// 				if err = session.Send(msg); nil != err {
	// 					common.Infof("fail to send msg.%v", err)
	// 					return
	// 				}
	// 				//休眠3秒等待发送退出消息成功
	// 				//time.Sleep(time.Duration(3) * time.Second)
	// 				common.Infof("Login from a different location.uid=%d", uid)
	// 				session.NoticeClose()
	// 				return
	// 			case <-ticker.C:
	// 				if 0 != isret {
	// 					common.Infof("connect close.uid=%d", uid)
	// 					return
	// 				}
	// 				//PushMsg(session, uid, seq)
	// 			}
	// 			if 0 != isret {
	// 				common.Infof("connect close.uid=%d", uid)
	// 				return
	// 			}
	// 		}
	// 	}
	// }()

	for {
		body, reqTemp, err, n := session.Receive()
		if 0 != session.controller.IsExit() {
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
			output, reqOut, err := session.controller.HandleReq(body, reqTemp)
			if err != nil {
				return
			}
			msg, err := session.controller.HandleRsp(output, reqOut)
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
	wg.Wait()
}
