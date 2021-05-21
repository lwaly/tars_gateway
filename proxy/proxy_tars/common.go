package proxy_tars

import (
	fmt "fmt"
	"net/http"
	"github.com/lwaly/tars_gateway/common"
	"sync"

	tars "github.com/TarsCloud/TarsGo/tars"
)

type StServer struct {
	Id        uint32   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	RouteType int      `json:"routeType,omitempty"`
	Switch    uint32   `json:"switch,omitempty"` //1开启服务
	Secret    string   `json:"secret,omitempty"`
	BlackList []string `json:"blackList,omitempty"` //
	WhiteList []string `json:"whiteList,omitempty"` //
}

type StApp struct {
	Id            uint32     `json:"id,omitempty"`
	Name          string     `json:"name,omitempty"`
	Server        []StServer `json:"server,omitempty"`
	Secret        string     `json:"secret,omitempty"`
	RouteType     int        `json:"routeType,omitempty"`
	Switch        uint32     `json:"switch,omitempty"`    //1开启服务
	BlackList     []string   `json:"blackList,omitempty"` //
	WhiteList     []string   `json:"whiteList,omitempty"` //
	GatewayObject string     `json:"gatewayObject,omitempty"`
}

type StTarsProxy struct {
	mapEndpoint map[string]tars.EndpointManager
	rwMutex     sync.RWMutex
	Secret      string  `json:"secret,omitempty"`
	RouteType   int     `json:"routeType,omitempty"`
	App         []StApp `json:"app,omitempty"`
	LimitObj    string  `json:"limitObj,omitempty"`
	RouteId     uint64
	key         string
}

var (
	comm            *tars.Communicator
	mapHttpCallBack map[string]*StTarsHttpProxyCommon
)

const MODIFY_RESPONSE = "ModifyResponse_key"

func init() {
	mapHttpCallBack = make(map[string]*StTarsHttpProxyCommon)
}

func (info *StTarsProxy) InitProxy(key string) (err error) {
	info.key = key

	err = common.Conf.GetStruct("tars", info)
	if err != nil {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}

	type StTarsAddr struct {
		Ip   string `json:"ip,omitempty"`
		Port int    `json:"port,omitempty"`
	}
	stTarsAddr := StTarsAddr{}
	err = common.Conf.GetStruct("tars", &stTarsAddr)
	if err != nil {
		common.Errorf("fail to get tcp conf.%v", err)
		return
	}
	addr := fmt.Sprintf("tars.tarsregistry.QueryObj@tcp -h %s -p %d -t 10000", stTarsAddr.Ip, stTarsAddr.Port)
	comm = tars.NewCommunicator()
	comm.SetLocator(addr)
	return
}

func (info *StTarsProxy) reloadConf() (err error) {
	if err = common.Conf.GetStruct(info.key, info); nil != err {
		common.Errorf("fail to get app info")
		return
	}
	return
}

type StTarsProxyTcpCommon struct {
	*StTarsProxy
	userCount          int64
	iSign              uint64
	mapApp             map[uint32]string
	mapServer          map[string]map[uint32]string
	mapEndpoint        map[string]tars.EndpointManager
	mapAppBlackList    map[uint32][]string
	mapServerBlackList map[uint32]map[uint32][]string
	mapAppWhiteList    map[uint32][]string
	mapServerWhiteList map[uint32]map[uint32][]string
	mapSecret          map[uint32]string
	mapAppQueueObject  map[uint32]string
}

func (info *StTarsProxyTcpCommon) InitProxy(key string) (err error) {
	info.StTarsProxy = new(StTarsProxy)
	info.StTarsProxy.InitProxy(key)
	info.mapApp = make(map[uint32]string)
	info.mapServer = make(map[string]map[uint32]string)
	info.mapSecret = make(map[uint32]string)
	info.mapAppQueueObject = make(map[uint32]string)
	info.mapAppBlackList = make(map[uint32][]string)
	info.mapServerBlackList = make(map[uint32]map[uint32][]string)
	info.mapAppWhiteList = make(map[uint32][]string)
	info.mapServerWhiteList = make(map[uint32]map[uint32][]string)

	return
}

func (info *StTarsProxyTcpCommon) reloadConf() (err error) {
	info.StTarsProxy.reloadConf()
	info.mapEndpoint = make(map[string]tars.EndpointManager)
	mapTemp := make(map[uint32]string)
	for _, value := range info.App {
		_, ok := mapTemp[value.Id]
		if ok {
			common.Errorf("repeat app.%v", value)
			continue
		} else {
			mapTemp[value.Id] = value.Name
		}

		_, ok = info.mapApp[value.Id]
		if !ok {
			info.mapApp[value.Id] = value.Name
		}

		mTemp := make(map[uint32]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Id]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}
			mTemp[v.Id] = v.Name
		}

		info.mapServer[value.Name] = mTemp
	}

	for _, value := range info.App {
		if "" != value.Secret && "empty" != value.Secret {
			info.mapSecret[value.Id] = value.Secret
		} else if "" != info.Secret && "empty" != info.Secret {
			info.mapSecret[value.Id] = info.Secret
		} else {
			info.mapSecret[value.Id] = ""
		}
	}

	for _, value := range info.App {
		if "" != value.GatewayObject && "empty" != value.GatewayObject {
			info.mapAppQueueObject[value.Id] = value.GatewayObject
		} else {
			info.mapAppQueueObject[value.Id] = ""
		}
	}

	for _, value := range info.App {
		_, ok := info.mapAppWhiteList[value.Id]
		if !ok {
			info.mapAppWhiteList[value.Id] = value.WhiteList
		}

		mTemp := make(map[uint32][]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Id]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}
			mTemp[v.Id] = v.WhiteList
		}

		info.mapServerWhiteList[value.Id] = mTemp
	}

	for _, value := range info.App {
		_, ok := info.mapAppBlackList[value.Id]
		if !ok {
			info.mapAppBlackList[value.Id] = value.BlackList
		}

		mTemp := make(map[uint32][]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Id]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}
			mTemp[v.Id] = v.BlackList
		}

		info.mapServerBlackList[value.Id] = mTemp
	}
	return
}

type StTarsHttpProxyCommon struct {
	*StTarsProxy
	userCount          int64
	iSign              uint64
	mapApp             map[uint32]string
	mapServer          map[string]map[uint32]string
	mapEndpoint        map[string]tars.EndpointManager
	mapAppBlackList    map[string][]string
	mapServerBlackList map[string][]string
	mapAppWhiteList    map[string][]string
	mapServerWhiteList map[string][]string
	mapSecret          map[string]string

	pCallBackStruct   interface{}
	pCallResponseFunc func(p interface{}, rsp *http.Response) error
	pCallRequestFunc  func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)
}

func (info *StTarsHttpProxyCommon) InitProxy(key string, p interface{}, ResponseFunc func(p interface{}, rsp *http.Response) error,
	RequestFunc func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)) (err error) {
	info.StTarsProxy = new(StTarsProxy)
	info.StTarsProxy.InitProxy(key)

	info.mapSecret = make(map[string]string)
	info.mapAppBlackList = make(map[string][]string)
	info.mapServerBlackList = make(map[string][]string)
	info.mapAppWhiteList = make(map[string][]string)
	info.mapServerWhiteList = make(map[string][]string)

	if _, ok := mapHttpCallBack[key]; !ok {
		info.pCallBackStruct = p
		info.pCallResponseFunc = ResponseFunc
		info.pCallRequestFunc = RequestFunc
		mapHttpCallBack[key] = info
	}
	return
}

func (info *StTarsHttpProxyCommon) reloadConf() (err error) {
	info.StTarsProxy.reloadConf()
	info.mapEndpoint = make(map[string]tars.EndpointManager)
	for _, value := range info.App {
		for _, v := range value.Server {
			if "" != v.Secret {
				info.mapSecret[value.Name+"."+v.Name] = v.Secret
			} else {
				info.mapSecret[value.Name+"."+v.Name] = value.Secret
			}
		}
	}

	for _, value := range info.App {
		mTemp := make(map[string]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Name]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}

			info.mapAppWhiteList[value.Name] = value.WhiteList
			info.mapServerWhiteList[value.Name+"."+v.Name] = v.WhiteList
		}
	}

	for _, value := range info.App {
		mTemp := make(map[string]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Name]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}

			info.mapAppBlackList[value.Name] = value.BlackList
			info.mapServerBlackList[value.Name+"."+v.Name] = v.BlackList
		}
	}
	return
}

func ModifyResponse(rsp *http.Response) (err error) {
	key := rsp.Header.Get(MODIFY_RESPONSE)
	if v, ok := mapHttpCallBack[key]; ok {
		if nil != v.pCallResponseFunc {
			if err = v.pCallResponseFunc(v.pCallBackStruct, rsp); nil != err {
				common.Warnf("More than the size of the max rate limit.")
				return
			}
		}
		rsp.Header.Del(MODIFY_RESPONSE)
	}

	return nil
}
