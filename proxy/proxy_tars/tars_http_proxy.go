package proxy_tars

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gofrs/flock"
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

type HttpControllerTars struct {
	Secret             string      `json:"secret,omitempty"`
	RouteType          int         `json:"routeType,omitempty"`
	App                []StHttpApp `json:"app,omitempty"`
	LimitObj           string      `json:"limitObj,omitempty"` //http对象
	mapHttpEndpoint    map[string]tars.EndpointManager
	fileLock           *flock.Flock
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

func (h *HttpControllerTars) ReloadConf() (err error) {
	if err = common.Conf.GetStruct("http", h); nil != err {
		common.Errorf("fail to get app info")
		return
	}
	for _, value := range h.App {
		for _, v := range value.Server {
			common.Infof("#####.%v.%v.%v", v.Secret, value.Secret, h.Secret)
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

func (h *HttpControllerTars) InitProxyHTTP(p interface{}, ResponseFunc func(p interface{}, rsp *http.Response) error,
	RequestFunc func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)) (err error) {
	h.mapHttpEndpoint = make(map[string]tars.EndpointManager)
	h.fileLock = flock.New("/var/lock/gateway-lock-http.lock")
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
func (h *HttpControllerTars) ServeHTTP(w http.ResponseWriter, r *http.Request) (err error) {
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

	appWhiteList, _ := h.mapAppWhiteList[a[1]]
	serverWhiteList, _ := h.mapServerWhiteList[a[1]+"."+a[2]]
	if 0 != len(appWhiteList) || 0 != len(serverWhiteList) {
		if !common.IpIsInlist(r.RemoteAddr, appWhiteList) || !common.IpIsInlist(r.RemoteAddr, serverWhiteList) {
			common.Errorf("addr not in WhiteList.%v", r.RemoteAddr)
			return
		}
	} else {
		appBlackList, _ := h.mapAppBlackList[a[1]]
		serverBlackList, _ := h.mapServerBlackList[a[1]+"."+a[2]]
		if common.IpIsInlist(r.RemoteAddr, appBlackList) || common.IpIsInlist(r.RemoteAddr, serverBlackList) {
			common.Errorf("addr not in WhiteList.%v", r.RemoteAddr)
			return
		}
	}
	var uid uint64
	secret, _ := h.mapSecret[a[1]+"."+a[2]]
	if "" != secret {
		if 0 == strings.Compare(a[3], "Login") || 0 == strings.Compare(a[3], "Register") || 0 == strings.Compare(a[3], "Verify") {
			common.Infof("user login")
			uid = 1
		} else {
			token := r.Header.Get("Token")
			claims, err := util.TokenAuth(token, secret)
			if err != nil {
				common.Errorf("authentication token fail.%v.%v", token, err)
				w.Write(common.NewErrorInfo(common.ERR_NO_USER, common.ErrNoUser))
				return err
			}
			uid = claims.Uid
		}
	}

	obj := fmt.Sprintf("%s.%s.%sObj", a[1], a[2], a[2])
	common.Infof("init EndpointManager.%s %s,uid=%d", r.URL.Path, obj, uid)

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

			if nil != pCallRequestFunc {
				if code, errTemp := pCallRequestFunc(pCallBackStruct, w, r); common.OK == code && nil != errTemp {
					if "CACHE" == errTemp.Error() {
						common.Infof("cache")
						return
					}
				}
			}

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
