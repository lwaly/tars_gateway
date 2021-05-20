package proxy

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"
)

type HttpController interface {
	ReloadConf() (err error)
	InitProxyHTTP(key string, p interface{}, ResponseFunc func(p interface{}, rsp *http.Response) error,
		RequestFunc func(p interface{}, w http.ResponseWriter, r *http.Request) (int, error)) (err error)
	ServeHTTP(w http.ResponseWriter, r *http.Request) (err error)
}

type StHttpController struct {
	controller HttpController
	proxyConf  *StProxyConf
}

func InitHttpProxy(key string) (proxyConf *StProxyConf, err error) {
	proxyConf = new(StProxyConf)
	proxyConf.key = key
	return proxyConf, proxyConf.reloadConf()
}

func StartHttpProxy(proxyConf *StProxyConf, h HttpController) (err error) {
	//监听端口
	controller := &StHttpController{proxyConf: proxyConf, controller: h}
	controller.controller.InitProxyHTTP(proxyConf.key, proxyConf, ModifyResponse, ModifyRequest)
	err = http.ListenAndServe(proxyConf.Addr, controller)

	if err != nil {
		common.Errorf("ListenAndServe:%s ", err.Error())
		return
	}

	ticker := time.NewTicker(time.Second * 5)
	confLastUpdateTime := time.Now().UnixNano()
	go func() {
		for {
			select {
			case <-ticker.C:
				if confLastUpdateTime < common.Conf.LastUpdateTimeGet() {
					confLastUpdateTime = common.Conf.LastUpdateTimeGet()
					proxyConf.reloadConf()
					controller.controller.ReloadConf()
				}
			}
		}
	}()

	return err
}

//实现Handler的接口
func (h *StHttpController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if 0 != len(h.proxyConf.WhiteList) {
		if !common.IpIsInlist(r.RemoteAddr, h.proxyConf.WhiteList) {
			common.Errorf("addr not in WhiteList.%v", r.RemoteAddr)
			return
		}
	} else if 0 != len(h.proxyConf.BlackList) && common.IpIsInlist(r.RemoteAddr, h.proxyConf.BlackList) {
		common.Errorf("addr in BlackList.%v", r.RemoteAddr)
		return
	}

	if err := util.HttpConnLimitAdd(h.proxyConf.LimitObj+r.URL.Path, 1); nil != err {
		common.Errorf("over connect limit.%v", err)
		return
	}
	defer util.HttpConnLimitAdd(h.proxyConf.LimitObj+r.URL.Path, -1)

	h.controller.ServeHTTP(w, r)
}

func ModifyResponse(p interface{}, rsp *http.Response) (err error) {
	proxyConf := p.(*StProxyConf)
	if nil != proxyConf {
		//计算是否可缓存，然后在最终得到限速的结果返回，限速不代表不能缓存
		cacheObj := rsp.Request.Header.Get("TARS_CACHE_OBJ")
		cacheKey := rsp.Request.Header.Get("TARS_CACHE_KEY")
		if "" != cacheObj && "" != cacheKey && util.CacheHttpObjExist(cacheObj) {
			defer rsp.Body.Close()
			if body, err := ioutil.ReadAll(rsp.Body); nil == err {
				cacheKeyHead := cacheKey + "head"
				cacheKeyBody := cacheKey + "body"
				rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

				if err = util.CacheHttpHeadAdd(cacheObj, cacheKeyHead, &rsp.Header); nil != err {
					common.Errorf("%v.%v", err, string(body))
				} else {
					if err = util.CacheHttpBodyAdd(cacheObj, cacheKeyBody, body); nil != err {
						common.Errorf("%v.%v", err, string(body))
					}
				}
			} else {
				common.Errorf("fail to read body.%v", err)
			}
		}

		temp := fmt.Sprintf("%s%s", proxyConf.LimitObj, rsp.Request.URL.Path)

		if err = util.RateHttpAdd(temp, rsp, 0); nil != err {
			common.Warnf("More than the size of the max rate limit.%s", temp)
			return
		}
	}

	return nil
}

func ModifyRequest(p interface{}, w http.ResponseWriter, r *http.Request) (code int, err error) {
	proxyConf := p.(*StProxyConf)

	if nil != proxyConf {
		cacheObj := fmt.Sprintf("%s%s", proxyConf.LimitObj, r.URL.Path)

		//查询cache
		if "1" == r.Header.Get("TARS_CACHE") && util.CacheHttpObjExist(cacheObj) {
			defer r.Body.Close()
			if body, err := ioutil.ReadAll(r.Body); nil == err {
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				ha := md5.New()
				ha.Write([]byte(fmt.Sprintf("%s%s%v", r.Method, r.URL.Path, body)))
				cacheKey := base64.StdEncoding.EncodeToString(ha.Sum(nil))
				r.Header.Add("TARS_CACHE_KEY", cacheKey)
				r.Header.Add("TARS_CACHE_OBJ", cacheObj)
				cacheKeyHead := cacheKey + "head"
				cacheKeyBody := cacheKey + "body"
				if err, head, n := util.CacheHttpHeadGet(cacheObj, cacheKeyHead); nil == err {
					if err, v := util.CacheHttpBodyGet(cacheObj, cacheKeyBody); nil == err {
						body, okbody := v.([]byte)
						if okbody {
							temp := fmt.Sprintf("%s%s", proxyConf.LimitObj, r.URL.Path)

							if err = util.RateHttpAdd(temp, nil, n+int64(len(body))); nil != err {
								common.Warnf("More than the size of the max rate limit.%s", temp)
								return common.ERR_LIMIT, err
							}
							if _, err = w.Write(body); nil == err {
								copyHeader(w.Header(), head)
								return common.OK, errors.New("CACHE")
							} else {
								common.Errorf("fail to write body.%v", err)
								return common.OK, nil
							}
						}
					}
				}
			}
		}
	}

	return common.OK, nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
