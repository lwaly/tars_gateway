package proxy

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type StTcpProxyConf struct {
	LimitObj        string         `json:"limitObj,omitempty"`
	Timeout         int            `json:"timeout,omitempty"`         //读写超时时间
	Heartbeat       int            `json:"heartbeat,omitempty"`       //心跳间隔
	MaxConn         int64          `json:"maxConn,omitempty"`         //最大连接数
	Addr            string         `json:"addr,omitempty"`            //tcp监听地址
	MaxRate         int64          `json:"maxRate,omitempty"`         //最大接收字节数
	MaxRatePer      int64          `json:"maxRatePer,omitempty"`      //每个连接最大接收字节数
	ConnCount       int64          `json:"connCount,omitempty"`       //已连接数
	RateCount       int64          `json:"rateCount,omitempty"`       //已接收字节数
	RatePerCount    int64          `json:"ratePerCount,omitempty"`    //每个连接已接收字节数
	Per             int64          `json:"per,omitempty"`             //限速统计间隔
	Switch          uint32         `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch uint32         `json:"rateLimitSwitch,omitempty"` //1开启服务
	App             []StTcpAppConf `json:"app,omitempty"`             //
	BlackList       []string       `json:"blackList,omitempty"`       //
	WhiteList       []string       `json:"whiteList,omitempty"`       //
}

type StTcpAppConf struct {
	Switch          uint32               `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch uint32               `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name            string               `json:"name,omitempty"`
	MaxConn         int64                `json:"maxConn,omitempty"`      //最大连接数
	MaxRate         int64                `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer      int64                `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount       int64                `json:"connCount,omitempty"`    //已连接数
	RateCount       int64                `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount    int64                `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per             int64                `json:"per,omitempty"`          //限速统计间隔
	Server          []StTcpAppServerConf `json:"server,omitempty"`       //
	BlackList       []string             `json:"blackList,omitempty"`    //
	WhiteList       []string             `json:"whiteList,omitempty"`    //
}

type StTcpAppServerConf struct {
	Switch          uint32   `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch uint32   `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name            string   `json:"name,omitempty"`
	MaxConn         int64    `json:"maxConn,omitempty"`      //最大连接数
	MaxRate         int64    `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer      int64    `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount       int64    `json:"connCount,omitempty"`    //已连接数
	RateCount       int64    `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount    int64    `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per             int64    `json:"per,omitempty"`          //限速统计间隔
	BlackList       []string `json:"blackList,omitempty"`    //
	WhiteList       []string `json:"whiteList,omitempty"`    //
}

type StLimit struct {
	Name string `json:"name,omitempty"`
	Rate int    `json:"rate,omitempty"` //读写超时时间
}

type Controller interface {
	ReloadConf() (err error)
	InitProxy() error
	TcpProxyGet(ip string) (info interface{})
	Verify(info interface{}) error
	HandleReq(info, body, reqTemp interface{}) (limitObj string, output, reqOut interface{}, err error)
	HandleRsp(info, output, reqOut interface{}) (outHeadRsp []byte, err error)
	IsExit(info interface{}) int
	Close(info interface{})
}

func ReloadConf(controller Controller, stTcpProxy *StTcpProxyConf) {
	//tcp连接配置读取
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			select {
			case <-ticker.C:
				if nil != stTcpProxy {
					reloadConf(stTcpProxy)
				}
				if nil != controller {
					controller.ReloadConf()
				}
			}
		}
	}()

	return
}

func reloadConf(stTcpProxy *StTcpProxyConf) (err error) {
	err = common.Conf.GetStruct("tcp", stTcpProxy)
	if err != nil {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	//限速初始化
	if common.SWITCH_ON == stTcpProxy.Switch {
		if common.SWITCH_ON == stTcpProxy.RateLimitSwitch {
			util.InitRateLimit(stTcpProxy.LimitObj, stTcpProxy.MaxRate, stTcpProxy.MaxRatePer, stTcpProxy.MaxConn, stTcpProxy.Per)
		} else {
			util.RateLimitDel(stTcpProxy.LimitObj)
		}
		for _, v := range stTcpProxy.App {
			if common.SWITCH_ON == v.RateLimitSwitch {
				util.InitRateLimit(stTcpProxy.LimitObj+"."+v.Name, v.MaxRate, v.MaxRatePer, v.MaxConn, v.Per)
			} else {
				util.RateLimitDel(stTcpProxy.LimitObj + "." + v.Name)
			}
			for _, v1 := range v.Server {
				if common.SWITCH_ON == v.RateLimitSwitch {
					util.InitRateLimit(stTcpProxy.LimitObj+"."+v.Name+"."+v1.Name, v1.MaxRate, v1.MaxRatePer, v1.MaxConn, v1.Per)
				} else {
					util.RateLimitDel(stTcpProxy.LimitObj + "." + v.Name + "." + v1.Name)
				}
			}
		}
	}

	return
}

func InitTcpProxy() (stTcpProxy *StTcpProxyConf, err error) {
	//tcp连接配置读取
	stTcpProxy = new(StTcpProxyConf)

	return stTcpProxy, reloadConf(stTcpProxy)
}

func ProxyTcpHandle(session *Session, conn net.Conn, stTcpProxyConf *StTcpProxyConf) {
	infoProtocol := session.codec.NewCodec(conn, stTcpProxyConf.Timeout, stTcpProxyConf.Heartbeat)
	defer session.Close(infoProtocol)
	if 0 != len(stTcpProxyConf.WhiteList) {
		if !common.IpIsInlist(conn.RemoteAddr().String(), stTcpProxyConf.WhiteList) {
			common.Errorf("addr not in WhiteList.%v", conn.RemoteAddr().String())
			return
		}
	} else if 0 != len(stTcpProxyConf.BlackList) && common.IpIsInlist(conn.RemoteAddr().String(), stTcpProxyConf.BlackList) {
		common.Errorf("addr in BlackList.%v", conn.RemoteAddr().String())
		return
	}

	if err := util.AddTcpConnLimit(stTcpProxyConf.LimitObj, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.AddTcpConnLimit(stTcpProxyConf.LimitObj, -1)
	var wg sync.WaitGroup
	infoTcpProxy := session.controller.TcpProxyGet(conn.RemoteAddr().String())
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

	connId, _ := strconv.ParseInt(fmt.Sprintf("%d", &conn), 10, 64)
	pHand := func(isFirstMsg *int, body, reqTemp interface{}) (limitObj string, err error) {
		defer wg.Done()
		tempLimitObj, output, reqOut, tempErr := session.controller.HandleReq(infoTcpProxy, body, reqTemp)
		if tempErr != nil {
			return "", tempErr
		}
		limitObj = tempLimitObj

		temp := fmt.Sprintf("%s.%s", stTcpProxyConf.LimitObj, limitObj)
		if 0 == *isFirstMsg {
			if err = util.AddTcpConnLimit(temp, 1); nil != err {
				common.Warnf("over limit connect.%s", limitObj)
				return
			}
			*isFirstMsg = 1
		}

		msg, tempErr := session.controller.HandleRsp(infoTcpProxy, output, reqOut)
		if tempErr != nil {
			common.Infof("fail to handle req msg.%v", err)
			return "", tempErr
		}

		//达到限速阀值,直接丢弃消息
		if "" != stTcpProxyConf.LimitObj {
			if err = util.AddRate(temp, int64(len(msg)), connId); nil != err {
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
		if 0 == isFirstMsg {
			if limitObj, err := pHand(&isFirstMsg, body, reqTemp); nil != err {
				common.Errorf("connect close.n=%d", n)
				session.NoticeClose(infoProtocol)
				break
			} else if 1 == isFirstMsg {
				defer util.AddTcpConnLimit(limitObj, -1)
			}
		} else {
			go pHand(&isFirstMsg, body, reqTemp)
		}

	}

	//关闭代理
	session.controller.Close(infoTcpProxy)
	wg.Wait()
}
