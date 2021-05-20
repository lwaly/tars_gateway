package proxy

import (
	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
	"time"
)

type StProxyConf struct {
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
	key                      string
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

func (stTcpProxy *StProxyConf) reloadConf() (err error) {
	//tcp连接配置读取
	err = common.Conf.GetStruct(stTcpProxy.key, stTcpProxy)
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
