package proxy_tars

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/syncmap"
)

const (
	BITS                 = 2048
	CMD_LOGOUT           = 1999
	CMD_HEART     uint32 = 101
	CONNECT_CLOSE int    = 1
)

var (
	mapUser syncmap.Map
	comm    *tars.Communicator
)

type StTcpServer struct {
	Id        uint32   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	RouteType int      `json:"routeType,omitempty"`
	Switch    uint32   `json:"switch,omitempty"`    //1开启服务
	BlackList []string `json:"blackList,omitempty"` //
	WhiteList []string `json:"whiteList,omitempty"` //
}

type StTcpApp struct {
	Id        uint32        `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	Server    []StTcpServer `json:"server,omitempty"`
	Secret    string        `json:"secret,omitempty"`
	RouteType int           `json:"routeType,omitempty"`
	Switch    uint32        `json:"switch,omitempty"`    //1开启服务
	BlackList []string      `json:"blackList,omitempty"` //
	WhiteList []string      `json:"whiteList,omitempty"` //
}

type StTarsTcpProxy struct {
	userCount          int64
	iSign              uint64
	mapApp             map[uint32]string
	mapServer          map[string]map[uint32]string
	mapAppBlackList    map[uint32][]string
	mapServerBlackList map[uint32]map[uint32][]string
	mapAppWhiteList    map[uint32][]string
	mapServerWhiteList map[uint32]map[uint32][]string
	mapSecret          map[uint32]string
	mapTcpEndpoint     map[string]tars.EndpointManager
	rwMutex            sync.RWMutex
	Secret             string     `json:"secret,omitempty"`
	RouteType          int        `json:"routeType,omitempty"`
	App                []StTcpApp `json:"app,omitempty"`
	LimitObj           string     `json:"limitObj,omitempty"`
	RouteId            uint64
}

type stTarsTcpProxy struct {
	privateKey *rsa.PrivateKey
	uid        uint64
	reader     chan []byte
	mapServer  syncmap.Map //
	isExit     int         //是否退出
	iSign      uint64      //结构体对象唯一标识
	Secret     string
	Addr       string
	outInfo    *StTarsTcpProxy
}

func (outInfo *StTarsTcpProxy) ReloadConf() (err error) {
	if err = common.Conf.GetStruct("tcp", outInfo); nil != err {
		common.Errorf("fail to get app info")
		return
	}
	outInfo.rwMutex.Lock()
	defer outInfo.rwMutex.Unlock()
	mapTemp := make(map[uint32]string)
	for _, value := range outInfo.App {
		_, ok := mapTemp[value.Id]
		if ok {
			common.Errorf("repeat app.%v", value)
			continue
		} else {
			mapTemp[value.Id] = value.Name
		}

		_, ok = outInfo.mapApp[value.Id]
		if !ok {
			outInfo.mapApp[value.Id] = value.Name
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

		outInfo.mapServer[value.Name] = mTemp
	}

	for _, value := range outInfo.App {
		if "" != value.Secret && "empty" != value.Secret {
			outInfo.mapSecret[value.Id] = value.Secret
		} else if "" != outInfo.Secret && "empty" != outInfo.Secret {
			outInfo.mapSecret[value.Id] = outInfo.Secret
		} else {
			outInfo.mapSecret[value.Id] = ""
		}
	}

	for _, value := range outInfo.App {
		_, ok := outInfo.mapAppWhiteList[value.Id]
		if !ok {
			outInfo.mapAppWhiteList[value.Id] = value.WhiteList
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

		outInfo.mapServerWhiteList[value.Id] = mTemp
	}

	for _, value := range outInfo.App {
		_, ok := outInfo.mapAppBlackList[value.Id]
		if !ok {
			outInfo.mapAppBlackList[value.Id] = value.BlackList
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

		outInfo.mapServerBlackList[value.Id] = mTemp
	}

	return
}

func (outInfo *StTarsTcpProxy) InitProxy() (err error) {
	outInfo.mapTcpEndpoint = make(map[string]tars.EndpointManager)

	outInfo.mapApp = make(map[uint32]string)
	outInfo.mapServer = make(map[string]map[uint32]string)
	outInfo.mapSecret = make(map[uint32]string)
	outInfo.mapAppBlackList = make(map[uint32][]string)
	outInfo.mapServerBlackList = make(map[uint32]map[uint32][]string)
	outInfo.mapAppWhiteList = make(map[uint32][]string)
	outInfo.mapServerWhiteList = make(map[uint32]map[uint32][]string)

	mapUser.Store(uint64(0), &stTarsTcpProxy{reader: make(chan []byte)})

	err = common.Conf.GetStruct("tars", outInfo)
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

	util.InitQueue(util.HandlerQueueFunc(HandleQueue))

	outInfo.ReloadConf()
	return
}

func (outInfo *StTarsTcpProxy) TcpProxyGet(ip string) interface{} {
	temp := new(stTarsTcpProxy)
	temp.outInfo = outInfo
	temp.Addr = ip
	temp.iSign = atomic.AddUint64(&outInfo.iSign, 1)
	return temp
}

func (outInfo *StTarsTcpProxy) Verify(info interface{}) error {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return errors.New("fail to convert")
	}
	return tempInfo.Verify()
}

func (outInfo *StTarsTcpProxy) HandlePre(info, reqHead, reqBody interface{}) (limitObj string, reqHeadOut, reqBodyOut interface{}, err error) {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return "", nil, nil, errors.New("fail to convert")
	}
	return tempInfo.HandlePre(reqHead, reqBody)
}

func (outInfo *StTarsTcpProxy) Handle(info, reqHead, reqBody interface{}) (outHeadRsp []byte, err error) {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return nil, errors.New("fail to convert")
	}
	return tempInfo.Handle(reqHead, reqBody)
}

func (outInfo *StTarsTcpProxy) IsExit(info interface{}) int {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return 0
	}
	return tempInfo.IsExit()
}

func (outInfo *StTarsTcpProxy) Close(info interface{}) {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return
	}
	tempInfo.isExit = CONNECT_CLOSE
	tempInfo.Close()
	return
}

func (info *stTarsTcpProxy) InitProxy() {
}

func (info *stTarsTcpProxy) HandlePre(reqHead, reqBody interface{}) (limitObj string, reqHeadOut, reqBodyOut interface{}, err error) {
	var ok bool
	reqHeadTemp, ok := reqHead.(MsgHead)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}
	info.outInfo.rwMutex.RLock()

	if v, ok := info.outInfo.mapSecret[reqHeadTemp.App]; ok {
		info.Secret = v
	}

	if err = info.whiteBlackQuery(&reqHeadTemp); nil != err {
		info.outInfo.rwMutex.RUnlock()
		return
	}

	if limitObj, err = info.objFind(&reqHeadTemp); nil != err {
		info.outInfo.rwMutex.RUnlock()
		return
	}
	info.outInfo.rwMutex.RUnlock()

	common.Infof("begin msg.server=%d.cmd=%d.Encrypt=%d.RouteId=%d.seq=%d",
		reqHeadTemp.GetServer(), reqHeadTemp.GetServant(), reqHeadTemp.GetEncrypt(), reqHeadTemp.GetRouteId(), reqHeadTemp.GetSeq())
	if 0 == reqHeadTemp.GetBodyLen() && CMD_HEART == reqHeadTemp.GetServant() {
		common.Infof("heart msg")
		return limitObj, reqHeadTemp, nil, nil
	} else if nil != reqBody {

		var reqBodyTemp MsgBody

		reqBodyTemp, ok = reqBody.(MsgBody)
		if !ok {
			common.Errorf("fail to convert head")
			err = errors.New("fail to convert head")
			return
		}

		if (1 == reqHeadTemp.GetEncrypt()) && (nil != info.privateKey) {
			if reqBodyTemp.Body, err = util.Decrypt(reqBodyTemp.Body, info.privateKey); nil != err {
				common.Errorf("fail to Decrypt msg")
				err = errors.New(common.ErrUnknown)
				return
			}
		} else if (2 == reqHeadTemp.GetEncrypt()) && (nil != info.privateKey) {
			if reqBodyTemp.Body, err = util.DecryptPkcs(reqBodyTemp.Body, info.privateKey); nil != err {
				common.Errorf("fail to Decrypt msg")
				err = errors.New(common.ErrUnknown)
				return
			}
		} else if 3 == reqHeadTemp.GetEncrypt() || nil == info.privateKey || 0 == reqHeadTemp.GetBodyLen() {
			common.Infof("do not verify or empty reqBody.%d %d", reqHeadTemp.GetBodyLen(), reqHeadTemp.GetEncrypt())
		} else {
			common.Errorf("fail to convert head")
			err = errors.New("fail to convert head")
			return
		}
		return limitObj, reqHeadTemp, reqBodyTemp, nil
	} else {
		return limitObj, reqHeadTemp, nil, nil
	}
}

func (info *stTarsTcpProxy) Handle(reqHead, reqBody interface{}) (msg []byte, err error) {
	var ok bool
	reqHeadOut, ok := reqHead.(MsgHead)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}
	var output Respond
	var obj string
	var cacheObj string
	var reqBodyTemp MsgBody
	var cacheKey string
	var outBodyRsp []byte

	if 0 == reqHeadOut.GetBodyLen() && CMD_HEART == reqHeadOut.GetServant() {
		common.Infof("heart msg")
		return info.handleResponse(outBodyRsp, cacheObj, cacheKey, &reqHeadOut, &output)
	}

	obj, err = info.objFind(&reqHeadOut)
	if nil != err {
		return
	}

	if nil != reqBody {
		reqBodyTemp, ok = reqBody.(MsgBody)
		if !ok {
			common.Errorf("fail to convert head")
			err = errors.New("fail to convert head")
			return
		}
	}

	if 1 == reqHeadOut.GetCacheIs() {
		cacheObj = info.outInfo.LimitObj + "." + obj
		if util.CacheTcpObjExist(cacheObj) {
			if cacheKey, outBodyRsp, err = cacheFind(cacheObj, &reqHeadOut, &reqBodyTemp); nil == err {
				common.Infof("cache")
				return info.handleResponse(outBodyRsp, cacheObj, cacheKey, &reqHeadOut, &output)
			}
		} else {
			cacheObj = ""
		}
	}

	if err = info.handle(obj, &reqHeadOut, &reqBodyTemp, &output); nil != err {
		common.Errorf("error handle:%v", err)
	}
	return info.handleResponse(outBodyRsp, cacheObj, cacheKey, &reqHeadOut, &output)
}

func (info *stTarsTcpProxy) handleResponse(outBodyRsp []byte, cacheObj string, cacheKey string, reqHeadOut *MsgHead, output *Respond) (msg []byte, err error) {
	if nil == outBodyRsp {
		if CMD_HEART != reqHeadOut.GetServant() {
			outBodyRsp, err = proto.Marshal(output)
			if err != nil {
				common.Errorf("faile to Marshal msg.err: %v", err)
				return
			}
			reqHeadOut.BodyLen = uint32(len(outBodyRsp))
			if "" != cacheObj {
				go cacheAdd(cacheObj, cacheKey, outBodyRsp)
			}
		} else {
			reqHeadOut.BodyLen = 1 //实际无body，保证客户端消息头部44字节
		}
	} else {
		reqHeadOut.BodyLen = uint32(len(outBodyRsp))
	}

	reqHeadOut.Servant++

	msg, err = proto.Marshal(reqHeadOut)
	if err != nil {
		common.Errorf("faile to Marshal msg.err: %v", err)
		return
	}

	msg = append(msg, outBodyRsp...)
	return
}

func (info *stTarsTcpProxy) handle(obj string, reqHeadOut *MsgHead, reqBodyTemp *MsgBody, output *Respond) (err error) {
	var ok bool
	var points []*endpoint.Endpoint
	var manager tars.EndpointManager

	obj += "Obj"
	manager, ok = info.outInfo.mapTcpEndpoint[obj]
	if !ok {
		manager = tars.GetManager(comm, obj)
		info.outInfo.mapTcpEndpoint[obj] = manager
		common.Infof("new EndpointManager.")
	}

	points = manager.GetAllEndpoint()
	if 0 != len(points) {
		var point *endpoint.Endpoint
		if 1 == info.outInfo.RouteType {
			tempId := atomic.AddUint64(&info.outInfo.RouteId, 1)
			point = points[tempId%uint64(len(points))]
		} else if 2 == info.outInfo.RouteType {
			point = points[reqHeadOut.GetRouteId()%uint64(len(points))]
		} else {
			tempId := atomic.AddUint64(&info.outInfo.RouteId, 1)
			point = points[tempId%uint64(len(points))]
		}
		if nil != point {
			var appObj *Server
			addr := fmt.Sprintf("%d%d", common.IPString2Long(point.Host), point.Port)
			appObjTemp, ok := info.mapServer.Load(addr)
			if !ok {
				appObj = new(Server)
				comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)
				info.mapServer.Store(addr, appObj)
			} else {
				appObj, ok = appObjTemp.(*Server)
				if !ok {
					info.mapServer.Delete(addr)
					appObj = new(Server)
					comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)
					info.mapServer.Store(addr, appObj)
				}
			}
			input := Request{Version: reqHeadOut.GetVersion(), Servant: reqHeadOut.GetServant(),
				Seq: reqHeadOut.GetSeq(), Uid: reqHeadOut.GetRouteId()}
			if nil != reqBodyTemp {
				input.Body = reqBodyTemp.GetBody()
			}
			*output, err = appObj.Handle(input)
			if err != nil {
				common.Errorf("err: %v reqBody=%s extend=%s", err, string(output.GetBody()[:]), string(output.GetExtend()[:]))
				var b []byte
				b, err = proto.Marshal(&ErrorRsp{Errinfo: &Errorinfo{ErrorCode: common.ERR_UNKNOWN, ErrorInfo: []byte(common.ErrUnknown)}})
				if err != nil {
					common.Errorf("faile to Marshal msg.err: %v", err)
					return
				}
				output.Body = b
				return
			}
			info.verify(output)
		}
	}

	return
}

func (info *stTarsTcpProxy) Verify() (err error) {
	return
}

func (info *stTarsTcpProxy) IsExit() int {
	return info.isExit
}

func (info *stTarsTcpProxy) verify(output *Respond) (err error) {
	//扩展字段有值，认为是客户端重新认证，更换秘钥
	if 0 != len(output.GetExtend()) && "" != info.Secret {
		if claims, err := util.TokenAuth(string(output.GetExtend()), info.Secret); nil != err {
			common.Errorf("authentication token fail.%v.%v.%v", string(output.GetExtend()), info.Secret, err)
			return err
		} else {
			if tempV, ok := mapUser.Load(claims.Uid); ok { //有同进程用户
				if v, ok := tempV.(*stTarsTcpProxy); ok {
					common.Infof("%v \n%v \n%v \n%v \n%v \n%v\n%v \n%v\n", v.iSign, info.iSign, v, info, &v, &info, *v, *info)
					if v.iSign != info.iSign { //不同连接
						//先关闭另一个连接
						v.isExit = CONNECT_CLOSE
						v.Close()
						if 0 != info.uid {
							//连接已校验成功，变更用户，用户数不变
							mapUser.Delete(info.uid)
						} else {
							//新连接，第一次校验成功，用户数加1
							atomic.AddInt64(&info.outInfo.userCount, 1)
						}
						common.Infof("same connect.modify user.uid=old %d,new %d,userCount=%d", info.uid, claims.Uid, info.outInfo.userCount)
						info.uid = claims.Uid
						mapUser.Store(info.uid, info)
					}
				}
			} else { //没有同进程用户，新用户
				if 0 != info.uid { //新用户，同连接
					common.Infof("same connect.modify user.uid=old %d,new %d,userCount=%d", info.uid, claims.Uid, info.outInfo.userCount)
					mapUser.Delete(info.uid)
					info.uid = claims.Uid
					mapUser.Store(info.uid, info)
				} else { //新用户，新连接
					info.uid = claims.Uid
					mapUser.Store(info.uid, info)
					atomic.AddInt64(&info.outInfo.userCount, 1)
					common.Infof("add connect.uid=%d,userCount=%d.%v", info.uid, info.outInfo.userCount, info)
				}
				userLoginNotify(info.uid)
			}
		}

		if info.privateKey, err = util.GenerateKey(BITS); err != nil {
			common.Errorf("Cannot generate RSA key.%v", err)
			return
		}

		if err, output.Extend = util.DumpPublicKeyBytes(&info.privateKey.PublicKey); err != nil {
			common.Errorf("Cannot generate RSA key.%v", err)
			return
		}
	}

	return
}

func (info *stTarsTcpProxy) objFind(reqOut *MsgHead) (strObj string, err error) {
	app, ok := info.outInfo.mapApp[reqOut.GetApp()]
	if !ok {
		common.Errorf("fail to get app name.%d", reqOut.GetApp())
		err = errors.New("fail to get app name")
		return
	}

	if mapAppS, ok := info.outInfo.mapServer[app]; !ok {
		common.Errorf("fail to get app name.%d", reqOut.GetApp())
		err = errors.New("fail to get app name ")
		return
	} else {
		if mapSt, ok := mapAppS[reqOut.GetServer()]; !ok {
			common.Errorf("fail to get app name.%d", reqOut.GetServer())
			err = errors.New("fail to get app name")
			return
		} else {
			return fmt.Sprintf("%s.%s.%s", app, mapSt, mapSt), nil
		}
	}
}

func (info *stTarsTcpProxy) whiteBlackQuery(reqHeadTemp *MsgHead) (err error) {
	isHaswhiteList := 0
	whiteList, _ := info.outInfo.mapAppWhiteList[reqHeadTemp.App]
	if 0 != len(whiteList) {
		isHaswhiteList = 1
		if !common.IpIsInlist(info.Addr, whiteList) {
			common.Errorf("addr not in WhiteList.%v", info.Addr)
			info.Close()
			info.isExit = CONNECT_CLOSE
			err = errors.New("it's not in whiteList")
			return
		}
	} else {
		whiteListServer, _ := info.outInfo.mapServerWhiteList[reqHeadTemp.App]
		if 0 != len(whiteListServer) {
			temp, _ := whiteListServer[reqHeadTemp.Server]
			if 0 != len(temp) {
				isHaswhiteList = 1
				if !common.IpIsInlist(info.Addr, temp) {
					common.Errorf("addr not in WhiteList.%v", info.Addr)
					info.Close()
					info.isExit = CONNECT_CLOSE
					err = errors.New("it's not in whiteList")
					return
				}
			}
		}
	}

	if 0 == isHaswhiteList {
		blackList, _ := info.outInfo.mapAppBlackList[reqHeadTemp.App]
		if 0 != len(blackList) {
			if common.IpIsInlist(info.Addr, blackList) {
				common.Errorf("it's in blackList.%v", info.Addr)
				info.Close()
				info.isExit = CONNECT_CLOSE
				err = errors.New("it's in blackList")
				return
			}
		} else {
			blackListServer, _ := info.outInfo.mapServerBlackList[reqHeadTemp.App]
			if 0 != len(blackListServer) {
				temp, _ := blackListServer[reqHeadTemp.Server]
				if 0 != len(temp) {
					if common.IpIsInlist(info.Addr, temp) {
						common.Errorf("it's in BlackList.%v", info.Addr)
						info.Close()
						info.isExit = CONNECT_CLOSE
						err = errors.New("it's in blackList")
						return
					}
				}
			}
		}
	}

	return
}

func (info *stTarsTcpProxy) Close() {
	if tempV, ok := mapUser.Load(info.uid); ok {
		if v, ok := tempV.(*stTarsTcpProxy); ok {
			if CONNECT_CLOSE == v.isExit {
				common.Infof("close.uid=%d,userCount=%d", info.uid, info.outInfo.userCount)
				mapUser.Delete(info.uid)
				atomic.AddInt64(&info.outInfo.userCount, -1)
			}
		}
	}
	common.Infof("close.uid=%d,userCount=%d,%v", info.uid, info.outInfo.userCount, info)
}

func HandleQueue(b []byte) {
	req := LoginNotifyReq{}
	err := proto.Unmarshal(b, &req)

	if nil != err {
		common.Errorf("fail Unmarshal msg.%v", err)
		return
	} else {
		tempV, ok := mapUser.Load(req.GetUid())
		if ok {
			v, ok := tempV.(*stTarsTcpProxy)
			if ok {
				common.Infof("user exit.uid=%d", req.GetUid())
				v.isExit = CONNECT_CLOSE
				return
			}
		}
	}
	return
}

func userLoginNotify(uid uint64) {
	info := LoginNotifyReq{Uid: uid}
	b, err := proto.Marshal(&info)
	if err != nil {
		common.Errorf("faile to Marshal msg.err: %v", err)
		return
	}
	util.QueueSend("tars_gateway", b, 1, uint32(ECmd_E_LOGIN_NOTIFY_REQ))
}

func cacheFind(obj string, head *MsgHead, reqBody *MsgBody) (cacheKey string, rspBody []byte, err error) {
	ha := md5.New()
	ha.Write([]byte(fmt.Sprintf("%d%d%d%d%d%v%v", head.Version, head.BodyLen, head.App, head.Server, head.Servant, reqBody.Body, reqBody.Extend)))
	cacheKey = base64.StdEncoding.EncodeToString(ha.Sum(nil))

	if err, v := util.CacheTcpGet(obj, cacheKey); nil == err {
		reqBody, okbody := v.([]byte)
		if okbody {
			return cacheKey, reqBody, nil
		}
	} else {
		return cacheKey, nil, err
	}

	return cacheKey, nil, errors.New("not find")
}

func cacheAdd(obj, cacheKey string, msg []byte) (err error) {
	if err := util.CacheTcpAdd(obj, cacheKey, msg); nil != err {
		common.Errorf("%v.%v", err, cacheKey)
		return err
	}
	return nil
}
