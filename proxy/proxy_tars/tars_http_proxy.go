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

type HttpControllerTars struct {
	mapHttpEndpoint map[string]tars.EndpointManager
	fileLock        *flock.Flock
	Secret          string `json:"secret,omitempty"`
	RouteType       int    `json:"routeType,omitempty"`
	RouteId         uint64
}

func (h *HttpControllerTars) ReloadConf() (err error) {
	if err = common.Conf.GetStruct("http", h); nil != err {
		common.Errorf("fail to get app info")
		return
	}
	return
}

func (h *HttpControllerTars) InitProxyHTTP() (err error) {
	h.mapHttpEndpoint = make(map[string]tars.EndpointManager)
	h.fileLock = flock.New("/var/lock/gateway-lock-http.lock")

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

	var uid uint64
	if "" != h.Secret {
		if 0 == strings.Compare(a[3], "Login") || 0 == strings.Compare(a[3], "Register") || 0 == strings.Compare(a[3], "Verify") {
			common.Infof("user login")
			uid = 1
		} else {
			token := r.Header.Get("Token")
			claims, err := util.TokenAuth(token, h.Secret)

			if err != nil {
				common.Errorf("authentication token fail.%v.%v.%v", token, h.Secret, err)
				w.Write(common.NewErrorInfo(common.ERR_NO_USER, common.ErrNoUser))
				return err
			}
			uid = claims.Uid
		}
	}

	obj := fmt.Sprintf("%s.%s.%sObj", a[1], a[2], a[2])
	common.Infof("init EndpointManager.%s %s,uid=%d", r.URL.Path, obj, uid)
	// manager, ok := h.mapHttpEndpoint[obj]
	// if !ok {
	// 	bLock := false
	// 	bLock, err = h.fileLock.TryLock()
	// 	if nil != err || false == bLock {
	// 		err = errors.New("fail to TryLock")
	// 		return
	// 	}
	// 	defer h.fileLock.Unlock()
	// 	manager = new(tars.EndpointManager)
	// 	manager.Init(obj, comm)
	// 	h.mapHttpEndpoint[obj] = manager
	// 	common.Infof("new EndpointManager. %s", obj)
	// }

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
			}

			proxy.ServeHTTP(w, r)
		}
	}
	return
}
