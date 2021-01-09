package proxy

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type StTcpProxyConf struct {
	LimitObj                 string         `json:"limitObj,omitempty"`
	Timeout                  int            `json:"timeout,omitempty"`         //读写超时时间
	Heartbeat                int            `json:"heartbeat,omitempty"`       //心跳间隔
	MaxConn                  int64          `json:"maxConn,omitempty"`         //最大连接数
	Addr                     string         `json:"addr,omitempty"`            //tcp监听地址
	MaxRate                  int64          `json:"maxRate,omitempty"`         //最大接收字节数
	MaxRatePer               int64          `json:"maxRatePer,omitempty"`      //每个连接最大接收字节数
	ConnCount                int64          `json:"connCount,omitempty"`       //已连接数
	RateCount                int64          `json:"rateCount,omitempty"`       //已接收字节数
	RatePerCount             int64          `json:"ratePerCount,omitempty"`    //每个连接已接收字节数
	Per                      int64          `json:"per,omitempty"`             //限速统计间隔
	Switch                   uint32         `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch          uint32         `json:"rateLimitSwitch,omitempty"` //1开启服务
	App                      []StTcpAppConf `json:"app,omitempty"`             //
	BlackList                []string       `json:"blackList,omitempty"`       //
	WhiteList                []string       `json:"whiteList,omitempty"`       //
	CacheSwitch              int64          `json:"cacheSwitch,omitempty"`
	CacheSize                int64          `json:"cacheSize,omitempty"`
	CacheExpirationTime      int64          `json:"cacheExpirationTime,omitempty"`
	CacheExpirationCleanTime string         `json:"cacheExpirationCleanTime,omitempty"`
}

type StTcpAppConf struct {
	Switch                   uint32               `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch          uint32               `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name                     string               `json:"name,omitempty"`
	MaxConn                  int64                `json:"maxConn,omitempty"`      //最大连接数
	MaxRate                  int64                `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer               int64                `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount                int64                `json:"connCount,omitempty"`    //已连接数
	RateCount                int64                `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount             int64                `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per                      int64                `json:"per,omitempty"`          //限速统计间隔
	Server                   []StTcpAppServerConf `json:"server,omitempty"`       //
	BlackList                []string             `json:"blackList,omitempty"`    //
	WhiteList                []string             `json:"whiteList,omitempty"`    //
	CacheSwitch              int64                `json:"cacheSwitch,omitempty"`
	CacheSize                int64                `json:"cacheSize,omitempty"`
	CacheExpirationTime      int64                `json:"cacheExpirationTime,omitempty"`
	CacheExpirationCleanTime string               `json:"cacheExpirationCleanTime,omitempty"`
}

type StTcpAppServerConf struct {
	Switch                   uint32   `json:"switch,omitempty"`          //1开启服务
	RateLimitSwitch          uint32   `json:"rateLimitSwitch,omitempty"` //1开启服务
	Name                     string   `json:"name,omitempty"`
	MaxConn                  int64    `json:"maxConn,omitempty"`      //最大连接数
	MaxRate                  int64    `json:"maxRate,omitempty"`      //最大接收字节数
	MaxRatePer               int64    `json:"maxRatePer,omitempty"`   //每个连接最大接收字节数
	ConnCount                int64    `json:"connCount,omitempty"`    //已连接数
	RateCount                int64    `json:"rateCount,omitempty"`    //已接收字节数
	RatePerCount             int64    `json:"ratePerCount,omitempty"` //每个连接已接收字节数
	Per                      int64    `json:"per,omitempty"`          //限速统计间隔
	BlackList                []string `json:"blackList,omitempty"`    //
	WhiteList                []string `json:"whiteList,omitempty"`    //
	CacheSwitch              int64    `json:"cacheSwitch,omitempty"`
	CacheSize                int64    `json:"cacheSize,omitempty"`
	CacheExpirationTime      int64    `json:"cacheExpirationTime,omitempty"`
	CacheExpirationCleanTime string   `json:"cacheExpirationCleanTime,omitempty"`
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
	HandlePre(info, reqHead, reqBody interface{}) (limitObj string, reqHeadOut, reqBodyOut interface{}, err error)
	Handle(info, reqHead, reqBody interface{}) (msg []byte, err error)
	IsExit(info interface{}) int
	Close(info interface{})
}

func reloadConf(stTcpProxy *StTcpProxyConf) (err error) {
	//tcp连接配置读取
	err = common.Conf.GetStruct("tcp", stTcpProxy)
	if err != nil {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	//限速初始化
	if common.SWITCH_ON == stTcpProxy.RateLimitSwitch {
		util.RateLimitInit(stTcpProxy.LimitObj, stTcpProxy.MaxRate, stTcpProxy.MaxRatePer, stTcpProxy.MaxConn, stTcpProxy.Per)
	} else if "" != stTcpProxy.LimitObj {
		util.RateLimitInit(stTcpProxy.LimitObj, 0, 0, 0, 0)
	}
	for _, v := range stTcpProxy.App {
		if common.SWITCH_ON == v.RateLimitSwitch {
			util.RateLimitInit(stTcpProxy.LimitObj+"."+v.Name, v.MaxRate, v.MaxRatePer, v.MaxConn, v.Per)
		} else if "" != stTcpProxy.LimitObj {
			util.RateLimitInit(stTcpProxy.LimitObj+"."+v.Name, 0, 0, 0, 0)
		}
		for _, v1 := range v.Server {
			if common.SWITCH_ON == v.RateLimitSwitch {
				util.RateLimitInit(stTcpProxy.LimitObj+"."+v.Name+"."+v1.Name, v1.MaxRate, v1.MaxRatePer, v1.MaxConn, v1.Per)
			} else if "" != stTcpProxy.LimitObj {
				util.RateLimitInit(stTcpProxy.LimitObj+"."+v.Name+"."+v1.Name, 0, 0, 0, 0)
			}
		}
	}

	//缓存初始化
	if common.SWITCH_ON == stTcpProxy.CacheSwitch {
		util.InitCache(stTcpProxy.LimitObj, stTcpProxy.CacheExpirationCleanTime,
			time.Duration(stTcpProxy.CacheExpirationTime)*time.Millisecond, stTcpProxy.CacheSize)
	} else if "" != stTcpProxy.LimitObj {
		util.InitCache(stTcpProxy.LimitObj, "", 0, 0)
	}
	for _, v := range stTcpProxy.App {
		if common.SWITCH_ON == v.CacheSwitch {
			util.InitCache(stTcpProxy.LimitObj+"."+v.Name, v.CacheExpirationCleanTime,
				time.Duration(v.CacheExpirationTime)*time.Millisecond, v.CacheSize)
		} else if "" != stTcpProxy.LimitObj {
			util.InitCache(stTcpProxy.LimitObj+"."+v.Name, "", 0, 0)
		}
		for _, v1 := range v.Server {
			if common.SWITCH_ON == v.CacheSwitch {
				util.InitCache(stTcpProxy.LimitObj+"."+v.Name+"."+v1.Name, v1.CacheExpirationCleanTime,
					time.Duration(v1.CacheExpirationTime)*time.Millisecond, v1.CacheSize)
			} else if "" != stTcpProxy.LimitObj {
				util.InitCache(stTcpProxy.LimitObj+"."+v.Name+"."+v1.Name, "", 0, 0)
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

func ProxyTcpHandle(session *Session) {
	infoProtocol := session.codec.NewCodec(session.conn, session.stTcpProxyConf.Timeout, session.stTcpProxyConf.Heartbeat)
	defer session.Close(infoProtocol)
	if 0 != len(session.stTcpProxyConf.WhiteList) {
		if !common.IpIsInlist(session.conn.RemoteAddr().String(), session.stTcpProxyConf.WhiteList) {
			common.Errorf("addr not in WhiteList.%v", session.conn.RemoteAddr().String())
			return
		}
	} else if 0 != len(session.stTcpProxyConf.BlackList) && common.IpIsInlist(session.conn.RemoteAddr().String(), session.stTcpProxyConf.BlackList) {
		common.Errorf("addr in BlackList.%v", session.conn.RemoteAddr().String())
		return
	}

	if err := util.TcpConnLimitAdd(session.stTcpProxyConf.LimitObj, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.TcpConnLimitAdd(session.stTcpProxyConf.LimitObj, -1)
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

		temp := fmt.Sprintf("%s.%s", session.stTcpProxyConf.LimitObj, limitObj)
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
		if "" != session.stTcpProxyConf.LimitObj {
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
