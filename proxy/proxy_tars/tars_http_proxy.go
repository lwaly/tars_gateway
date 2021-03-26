package proxy_tars

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

type StHttpServer struct {
	Id        uint32   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Secret    string   `json:"secret,omitempty"`
	RouteType int      `json:"routeType,omitempty"`
	Switch    uint32   `json:"switch,omitempty"`    //1开启服务
	BlackList []string `json:"blackList,omitempty"` //
	WhiteList []string `json:"whiteList,omitempty"` //
}

type StHttpApp struct {
	Id        uint32         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Server    []StHttpServer `json:"server,omitempty"`
	Secret    string         `json:"secret,omitempty"`
	RouteType int            `json:"routeType,omitempty"`
	Switch    uint32         `json:"switch,omitempty"`    //1开启服务
	BlackList []string       `json:"blackList,omitempty"` //
	WhiteList []string       `json:"whiteList,omitempty"` //
}

type StTarsHttpProxy struct {
	Secret             string      `json:"secret,omitempty"`
	RouteType          int         `json:"routeType,omitempty"`
	App                []StHttpApp `json:"app,omitempty"`
	LimitObj           string      `json:"limitObj,omitempty"` //http对象
	rwMutex            sync.RWMutex
	mapHttpEndpoint    map[string]tars.EndpointManager
	mapSecret          map[string]string
	RouteId            uint64
	mapAppBlackList    map[string][]string
	mapServerBlackList map[string][]string
	mapAppWhiteList    map[string][]string
	mapServerWhiteList map[string][]string
}

var pCallBackStruct interface{}
var pCallResponseFunc func(p interface{}, rsp *http.Response) error
var pCallRequestFunc func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)

func (h *StTarsHttpProxy) ReloadConf() (err error) {
	if err = common.Conf.GetStruct("http", h); nil != err {
		common.Errorf("fail to get app info")
		return
	}
	h.rwMutex.Lock()
	defer h.rwMutex.Unlock()
	for _, value := range h.App {
		for _, v := range value.Server {
			if "" != v.Secret && "empty" != v.Secret {
				h.mapSecret[value.Name+"."+v.Name] = v.Secret
			} else if "" != value.Secret && "empty" != value.Secret {
				h.mapSecret[value.Name+"."+v.Name] = value.Secret
			} else if "" != h.Secret && "empty" != h.Secret {
				h.mapSecret[value.Name+"."+v.Name] = h.Secret
			} else {
				h.mapSecret[value.Name+"."+v.Name] = ""
			}
		}
	}

	for _, value := range h.App {
		mTemp := make(map[string]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Name]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}

			h.mapAppWhiteList[value.Name] = value.WhiteList
			h.mapServerWhiteList[value.Name+"."+v.Name] = v.WhiteList
		}
	}

	for _, value := range h.App {
		mTemp := make(map[string]string)
		for _, v := range value.Server {
			_, ok := mTemp[v.Name]
			if ok {
				common.Errorf("repeat serverId.%v", v)
				continue
			}

			h.mapAppBlackList[value.Name] = value.BlackList
			h.mapServerBlackList[value.Name+"."+v.Name] = v.BlackList
		}
	}
	return
}

func (h *StTarsHttpProxy) InitProxyHTTP(p interface{}, ResponseFunc func(p interface{}, rsp *http.Response) error,
	RequestFunc func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)) (err error) {
	h.mapHttpEndpoint = make(map[string]tars.EndpointManager)

	h.mapSecret = make(map[string]string)
	h.mapAppBlackList = make(map[string][]string)
	h.mapServerBlackList = make(map[string][]string)
	h.mapAppWhiteList = make(map[string][]string)
	h.mapServerWhiteList = make(map[string][]string)

	pCallBackStruct = p
	pCallResponseFunc = ResponseFunc
	pCallRequestFunc = RequestFunc
	h.ReloadConf()
	return
}

//实现Handler的接口
func (h *StTarsHttpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Token")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	if 0 == strings.Compare(r.Method, "OPTIONS") {
		return
	}

	a := strings.Split(r.URL.Path, "/")

	if len(a) < 4 {
		common.Errorf("fail to Split url.%s", r.URL.Path)
		w.Write(common.NewErrorInfo(common.ERR_UNKNOWN, common.ErrUnknown))
		return
	}

	h.rwMutex.RLock()
	if err = h.whiteBlackQuery(a, r); nil != err {
		w.WriteHeader(http.StatusBadGateway)
		h.rwMutex.RUnlock()
		return
	}
	h.rwMutex.RUnlock()

	//verify Token
	if err = h.verifyToken(a, w, r); nil != err {
		return
	}
	//query cache
	if nil != pCallRequestFunc {
		if code, errTemp := pCallRequestFunc(pCallBackStruct, w, r); common.OK == code && nil != errTemp {
			if "CACHE" == errTemp.Error() {
				common.Infof("cache")
				return
			}
		} else if common.ERR_LIMIT == code {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}

	return h.serveHTTP(a, w, r)
}

func (h *StTarsHttpProxy) verifyToken(a []string, w http.ResponseWriter, r *http.Request) (err error) {
	secret, _ := h.mapSecret[a[1]+"."+a[2]]
	if "" != secret {
		if 0 == strings.Compare(a[3], "Login") || 0 == strings.Compare(a[3], "Register") || 0 == strings.Compare(a[3], "Verify") {
			common.Infof("user login")
		} else {
			token := r.Header.Get("Token")
			_, err := util.TokenAuth(token, secret)
			if err != nil {
				common.Errorf("authentication token fail.%v.%v", token, err)
				w.Write(common.NewErrorInfo(common.ERR_NO_USER, common.ErrNoUser))
				return err
			}
		}
	}

	return
}

func (h *StTarsHttpProxy) serveHTTP(a []string, w http.ResponseWriter, r *http.Request) (err error) {
	obj := fmt.Sprintf("%s.%s.%sObj", a[1], a[2], a[2])
	manager, ok := h.mapHttpEndpoint[obj]
	if !ok {
		manager = tars.GetManager(comm, obj)
		h.mapHttpEndpoint[obj] = manager
		common.Infof("new EndpointManager.")
	}

	points := manager.GetAllEndpoint()
	if 0 != len(points) {
		var point *endpoint.Endpoint

		if 1 == h.RouteType {
			tempId := atomic.AddUint64(&h.RouteId, 1)
			point = points[tempId%uint64(len(points))]
		} else if 2 == h.RouteType {
			if route, err := strconv.ParseInt(r.Header.Get("route_type"), 10, 64); nil != err {
				common.Warnf("fail to get route.%s", r.Header.Get("route_type"))
				tempId := atomic.AddUint64(&h.RouteId, 1)
				point = points[tempId%uint64(len(points))]
			} else {
				point = points[route%int64(len(points))]
			}
		} else {
			tempId := atomic.AddUint64(&h.RouteId, 1)
			point = points[tempId%uint64(len(points))]
		}

		if nil != point {
			inner := fmt.Sprintf("%s:%d", point.Host, point.Port)
			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					//设置主机
					req.URL.Host = inner
					req.URL.Scheme = "http"
					//设置路径
					req.URL.Path = r.URL.Path
					//设置参数
					req.PostForm = r.PostForm
					req.URL.RawQuery = r.URL.RawQuery
					req.Form = r.Form
				},
				ModifyResponse: ModifyResponse,
			}

			proxy.ServeHTTP(w, r)
		}
	}
	return
}

func ModifyResponse(rsp *http.Response) (err error) {
	if nil != pCallResponseFunc {
		if err = pCallResponseFunc(pCallBackStruct, rsp); nil != err {
			common.Warnf("More than the size of the max rate limit.")
			return
		}
	}

	return nil
}

func (h *StTarsHttpProxy) whiteBlackQuery(a []string, r *http.Request) (err error) {
	appWhiteList, _ := h.mapAppWhiteList[a[1]]
	serverWhiteList, _ := h.mapServerWhiteList[a[1]+"."+a[2]]
	if 0 != len(appWhiteList) || 0 != len(serverWhiteList) {
		if !common.IpIsInlist(r.RemoteAddr, appWhiteList) || !common.IpIsInlist(r.RemoteAddr, serverWhiteList) {
			common.Errorf("addr not in WhiteList.%v", r.RemoteAddr)
			err = errors.New("it's not in whiteList")
			return
		}
	} else {
		appBlackList, _ := h.mapAppBlackList[a[1]]
		serverBlackList, _ := h.mapServerBlackList[a[1]+"."+a[2]]
		if common.IpIsInlist(r.RemoteAddr, appBlackList) || common.IpIsInlist(r.RemoteAddr, serverBlackList) {
			common.Errorf("addr not in WhiteList.%v", r.RemoteAddr)
			err = errors.New("it's in blackList")
			return
		}
	}
	return
}
