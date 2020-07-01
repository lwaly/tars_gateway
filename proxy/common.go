package proxy

import (
	"fmt"
	"sync"

	"github.com/lwaly/tars_gateway/common"
)

var secret = ""
var machine int
var mapKey24 map[string]int64
var mutex sync.Mutex
var g_seq uint32

func init() {
}

func GetSeq() uint32 {
	mutex.Lock()
	defer mutex.Unlock()
	g_seq++
	return g_seq
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
}
