package proxy

import (
	"sync"

	"github.com/lwaly/tars_gateway/common"
)

var secret = ""
var machine int
var mapKey24 map[string]int64
var mutex sync.Mutex
var g_seq uint32

func GetSeq() uint32 {
	mutex.Lock()
	defer mutex.Unlock()
	g_seq++
	return g_seq
}

type StLogConf struct {
	Path  string `json:"path,omitempty"`
	Level int    `json:"level,omitempty"`
	Unit  string `json:"unit,omitempty"`
}

func InitProxy() {
	//日志系统初始化
	stLogConf := StLogConf{}

	err := common.Conf.GetStruct("log", &stLogConf)
	if err != nil {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	//minute hour day week month
	if "minute" == stLogConf.Unit {
		defer common.Start(common.LogFilePath(stLogConf.Path), common.EveryMinute).Stop()
	} else if "hour" == stLogConf.Unit {
		defer common.Start(common.LogFilePath(stLogConf.Path), common.EveryHour).Stop()
	} else if "day" == stLogConf.Unit {
		defer common.Start(common.LogFilePath(stLogConf.Path), common.EveryDay).Stop()
	} else if "week" == stLogConf.Unit {
		defer common.Start(common.LogFilePath(stLogConf.Path), common.EveryWeek).Stop()
	} else if "month" == stLogConf.Unit {
		defer common.Start(common.LogFilePath(stLogConf.Path), common.EveryMonth).Stop()
	} else {
		defer common.Start(common.LogFilePath(stLogConf.Path), common.EveryHour).Stop()
	}

	common.LevelSet(stLogConf.Level)
}
