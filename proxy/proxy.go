package proxy

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type stTcpProxy struct {
	timeout      int    //读写超时时间
	heartbeat    int    //心跳间隔
	maxConn      int64  //最大连接数
	strAddrTcp   string //tcp监听地址
	maxRate      int64  //最大接收字节数
	maxRatePer   int64  //每个连接最大接收字节数
	connCount    int64  //已连接数
	rateCount    int64  //已接收字节数
	ratePerCount int64  //每个连接已接收字节数
	per          int64  //限速统计间隔
}

var lstTcpProxy stTcpProxy

const ltarsTcp = "tarsTcp"

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
	strPath, err := common.Conf.GetValue("log", "path")
	if nil != err {
		fmt.Printf("fail to get log path.%v", err)
		return
	}

	strUnit, err := common.Conf.GetValue("log", "unit")
	if nil != err {
		fmt.Printf("fail to get log unit.%v", err)
		return
	}
	//minute hour day week month
	if "minute" == strUnit {
		defer common.Start(common.LogFilePath(strPath), common.EveryMinute).Stop()
	} else if "hour" == strUnit {
		defer common.Start(common.LogFilePath(strPath), common.EveryHour).Stop()
	} else if "day" == strUnit {
		defer common.Start(common.LogFilePath(strPath), common.EveryDay).Stop()
	} else if "week" == strUnit {
		defer common.Start(common.LogFilePath(strPath), common.EveryWeek).Stop()
	} else if "month" == strUnit {
		defer common.Start(common.LogFilePath(strPath), common.EveryMonth).Stop()
	} else {
		defer common.Start(common.LogFilePath(strPath), common.EveryHour).Stop()
	}

	level, err := common.Conf.Int("log", "level")
	if nil != err {
		fmt.Printf("fail to get log level.%v", err)
		return
	}

	common.LevelSet(level)

	//tcp连接配置读取
	lstTcpProxy.timeout, err = common.Conf.Int("tcp", "timeout")
	if nil != err {
		common.Errorf("fail to get tcp timeout.%v", err)
		return
	}

	lstTcpProxy.maxConn, err = common.Conf.Int64("tcp", "max_conn")
	if nil != err {
		common.Errorf("fail to get tcp maxConn.%v", err)
		return
	}

	lstTcpProxy.maxRate, err = common.Conf.Int64("tcp", "rate")
	if nil != err {
		common.Errorf("fail to get tcp maxConn.%v", err)
		return
	}

	lstTcpProxy.maxRatePer, err = common.Conf.Int64("tcp", "rate_per")
	if nil != err {
		common.Errorf("fail to get tcp maxConn.%v", err)
		return
	}

	lstTcpProxy.heartbeat, err = common.Conf.Int("tcp", "heartbeat")
	if nil != err {
		common.Errorf("fail to get tcp heartbeat.%v", err)
		return
	}

	lstTcpProxy.per, err = common.Conf.Int64("tcp", "per")
	if nil != err {
		common.Errorf("fail to get tcp per.%v", err)
		return
	}

	lstTcpProxy.strAddrTcp, err = common.Conf.GetValue("tcp", "addr")
	if nil != err {
		common.Errorf("fail to get tcp addr.%v", err)
		return
	}

	util.InitRateLimit(ltarsTcp, lstTcpProxy.maxRate, lstTcpProxy.maxRatePer, lstTcpProxy.per)
}

func ProxyTcpHandle(session *Session, conn net.Conn) {
	//粗略限连接数，不做复杂的使用锁机制精确限连接数
	if lstTcpProxy.connCount >= lstTcpProxy.maxConn {
		common.Warnf("The max connect is reached.%d %d", lstTcpProxy.connCount, lstTcpProxy.maxConn)
		return
	}
	atomic.AddInt64(&lstTcpProxy.connCount, 1)
	defer atomic.AddInt64(&lstTcpProxy.connCount, -1)

	var wg sync.WaitGroup
	infoTcpProxy := session.controller.TcpProxyGet()
	infoProtocol := session.codec.NewCodec(conn, lstTcpProxy.timeout, lstTcpProxy.heartbeat)
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
		if err = util.AddRate(ltarsTcp, int64(n)); nil != err {
			common.Warnf("More than the size of the max rate limit.%d", n)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
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
