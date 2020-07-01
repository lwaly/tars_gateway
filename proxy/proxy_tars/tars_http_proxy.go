package proxy_tars

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"

	"github.com/TarsCloud/TarsGo/tars"
)

var (
	mapHttpEndpoint map[string]*tars.EndpointManager
	secret_http     string
)

type HttpControllerTars struct {
}

func init() {
	mapHttpEndpoint = make(map[string]*tars.EndpointManager)
}

func (h *HttpControllerTars) InitProxyHTTP() error {
	secret_http, _ = common.Conf.GetValue("http", "secret")
	return nil
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
	if "" != secret_http {
		if 0 == strings.Compare(a[3], "Login") || 0 == strings.Compare(a[3], "Register") || 0 == strings.Compare(a[3], "Verify") {
			common.Infof("user login")
			uid = 1
		} else {
			token := r.Header.Get("Token")
			claims, err := util.TokenAuth(token, secret_http)

			if err != nil {
				common.Errorf("authentication token fail.%v.%v.%v", token, secret_http, err)
				w.Write(common.NewErrorInfo(common.ERR_NO_USER, common.ErrNoUser))
				return err
			}
			uid = claims.Uid
		}
	}

	obj := fmt.Sprintf("%s.%s.%sObj", a[1], a[2], a[2])
	common.Infof("init EndpointManager.%s %s,uid=%d", r.URL.Path, obj, uid)
	manager, ok := mapHttpEndpoint[obj]
	if !ok {
		bLock := false
		bLock, err = fileLock.TryLock()
		if nil != err || false == bLock {
			err = errors.New("fail to TryLock")
			return
		}
		defer fileLock.Unlock()
		manager = new(tars.EndpointManager)
		manager.Init(obj, comm)
		mapHttpEndpoint[obj] = manager
		common.Infof("new EndpointManager. %s", obj)
	}

	point := manager.GetNextEndpoint()
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

	return
}
