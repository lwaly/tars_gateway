package proxy_tars

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/lwaly/tars_gateway/common"
	. "github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/util"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/gofrs/flock"
	"github.com/golang/protobuf/proto"
	"golang.org/sync/syncmap"
)

const BITS = 2048
const CMD_LOGOUT = 1999
const CMD_HEART uint32 = 101
const CONNECT_CLOSE int = 1

var userCount int64
var iSign uint64

type StTarsTcpProxy struct {
}

func (outInfo *StTarsTcpProxy) InitProxy() {
	util.InitQueue(util.HandlerQueueFunc(HandleQueue))
}
func (outInfo *StTarsTcpProxy) TcpProxyGet() interface{} {
	temp := new(stTarsTcpProxy)
	temp.iSign = atomic.AddUint64(&iSign, 1)
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
func (outInfo *StTarsTcpProxy) HandleReq(info, body, reqTemp interface{}) (output, reqOut interface{}, err error) {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return nil, nil, errors.New("fail to convert")
	}
	return tempInfo.HandleReq(body, reqTemp)
}
func (outInfo *StTarsTcpProxy) HandleRsp(info, output, reqOut interface{}) (outHeadRsp []byte, err error) {
	tempInfo, ok := info.(*stTarsTcpProxy)
	if !ok {
		common.Errorf("fail to convert")
		return nil, errors.New("fail to convert")
	}
	return tempInfo.HandleRsp(output, reqOut)
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

type stTarsTcpProxy struct {
	privateKey *rsa.PrivateKey
	uid        uint64
	reader     chan []byte
	mapServer  syncmap.Map //
	isExit     int         //是否退出
	iSign      uint64      //结构体对象唯一标识
}

var mapApp map[uint32]string
var mapServer map[string]map[uint32]string
var comm *tars.Communicator
var mapTcpEndpoint map[string]*tars.EndpointManager

var mapUser syncmap.Map

var fileLock *flock.Flock
var secret = "test"

func init() {
	mapTcpEndpoint = make(map[string]*tars.EndpointManager)

	mapApp = make(map[uint32]string)
	mapServer = make(map[string]map[uint32]string)
	mapUser.Store(uint64(0), &stTarsTcpProxy{reader: make(chan []byte)})

	fileLock = flock.New("/var/lock/gateway-lock.lock")

	secret, _ = common.Conf.GetValue("token", "secret")

	mapTemp, err := common.Conf.GetSection("app")
	if nil != err {
		fmt.Printf("fail to get app info")
		return
	}
	for key, value := range mapTemp {
		s := strings.Split(value, " ")
		if 0 == len(s) {
			fmt.Println("fail to Split app info")
			break
		}
		var appid int
		if appid, err = strconv.Atoi(s[0]); err != nil {
			fmt.Printf("fail to get app id")
			return
		}
		_, ok := mapApp[uint32(appid)]
		if ok {
			fmt.Println("repeat app", appid)
			break
		}
		mapApp[uint32(appid)] = key

		mTemp := make(map[uint32]string)

		for index := 1; index < len(s); {
			var serverId int
			if serverId, err = strconv.Atoi(s[index]); err != nil {
				fmt.Printf("fail to get app id")
				return
			}
			_, ok := mTemp[uint32(serverId)]
			if ok {
				fmt.Println("repeat serverId", serverId)
				return
			}
			index++
			mTemp[uint32(serverId)] = s[index]
			index++
		}
		_, ok = mapServer[key]
		if ok {
			fmt.Println("repeat app", key)
			break
		}
		mapServer[key] = mTemp
	}
	ip, err := common.Conf.GetValue("tars", "ip")
	if nil != err {
		fmt.Printf("fail to get log path")
		return
	}

	port, err := common.Conf.GetValue("tars", "port")
	if nil != err {
		fmt.Printf("fail to get log path")
		return
	}

	addr := fmt.Sprintf("tars.tarsregistry.QueryObj@tcp -h %s -p %s -t 10000", ip, port)
	comm = tars.NewCommunicator()
	comm.SetLocator(addr)
}

func (info *stTarsTcpProxy) InitProxy() {
}

func (info *stTarsTcpProxy) HandleReq(body, reqTemp interface{}) (output, reqOut interface{}, err error) {
	var ok bool
	reqOutTemp, ok := reqTemp.(MsgHead)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}
	key := "1"
	common.Infof("begin msg.server=%d.cmd=%d.Encrypt=%d.RouteId=%d.seq=%d.uid=%s",
		reqOutTemp.GetServer(), reqOutTemp.GetServant(), reqOutTemp.GetEncrypt(), reqOutTemp.GetRouteId(), reqOutTemp.GetSeq(), key)
	if 0 == reqOutTemp.GetBodyLen() && CMD_HEART == reqOutTemp.GetServant() {
		common.Infof("heart msg")
		return
	}

	var rspBody MsgBody

	if nil != body {
		rspBody, ok = body.(MsgBody)
		if !ok {
			common.Errorf("fail to convert head")
			err = errors.New("fail to convert head")
			return
		}
	}

	if (1 == reqOutTemp.GetEncrypt()) && (nil != info.privateKey) {
		if rspBody.Body, err = util.Decrypt(rspBody.Body, info.privateKey); nil != err {
			common.Errorf("fail to Decrypt msg")
			err = errors.New(ErrUnknown)
			return
		}
	} else if (2 == reqOutTemp.GetEncrypt()) && (nil != info.privateKey) {
		if rspBody.Body, err = util.DecryptPkcs(rspBody.Body, info.privateKey); nil != err {
			common.Errorf("fail to Decrypt msg")
			err = errors.New(ErrUnknown)
			return
		}
	} else if 3 == reqOutTemp.GetEncrypt() || nil == info.privateKey || 0 == reqOutTemp.GetBodyLen() {
		common.Infof("do not verify or empty body.%d %d", reqOutTemp.GetBodyLen(), reqOutTemp.GetEncrypt())
	} else {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}

	obj, err := info.objFind(&reqOutTemp)

	if nil != err {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}

	var manager *tars.EndpointManager
	manager, ok = mapTcpEndpoint[obj]
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
		mapTcpEndpoint[obj] = manager
		common.Infof("new EndpointManager. %s", obj)
	}

	point := manager.GetNextEndpoint()
	if nil != point {
		var appObj *Server
		addr := fmt.Sprintf("%d%d", IPString2Long(point.Host), point.Port)
		appObjTemp, ok := info.mapServer.Load(addr)
		if !ok {
			common.Infof("first.%s", key)
			appObj = new(Server)
			comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)
			info.mapServer.Store(addr, appObj)
		} else {
			appObj, ok = appObjTemp.(*Server)
			if !ok {
				common.Infof("first.%s", key)
				info.mapServer.Delete(addr)
				appObj = new(Server)
				comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)
				info.mapServer.Store(addr, appObj)
			}
		}

		input := Request{Version: reqOutTemp.GetVersion(), Servant: reqOutTemp.GetServant(), Seq: reqOutTemp.GetSeq(), Uid: reqOutTemp.GetRouteId(), Body: rspBody.GetBody()}
		outputTemp, err := appObj.Handle(input)
		if err != nil {
			common.Errorf("err: %v body=%s extend=%s", err, string(outputTemp.GetBody()[:]), string(outputTemp.GetExtend()[:]))
			//user.mapServer.Delete(addr)
			err = errors.New(common.ErrUnknown)
			return outputTemp, reqOutTemp, err
		}
		info.verify(&outputTemp)
		output = outputTemp
		reqOut = reqOutTemp
	}

	return
}

func (info *stTarsTcpProxy) HandleRsp(output, reqOut interface{}) (outHeadRsp []byte, err error) {
	var ok bool
	reqOutTemp, ok := reqOut.(MsgHead)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}

	outputTemp, ok := output.(Respond)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}

	var outBodyRsp []byte
	if CMD_HEART != reqOutTemp.GetServant() {
		outBodyRsp, err = proto.Marshal(&outputTemp)
		if err != nil {
			common.Errorf("faile to Marshal msg.err: %v", err)
			return
		}
		reqOutTemp.BodyLen = uint32(len(outBodyRsp))
	} else {
		reqOutTemp.BodyLen = 1
	}

	reqOutTemp.Servant += 1
	outHeadRsp, err = proto.Marshal(&reqOutTemp)
	if err != nil {
		common.Errorf("faile to Marshal msg.err: %v", err)
		return
	}

	outHeadRsp = append(outHeadRsp, outBodyRsp...)
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
	if 0 != len(output.GetExtend()) {
		if claims, err := util.TokenAuth(string(output.GetExtend()), secret); nil != err {
			common.Errorf("authentication token fail.%v.%v.%v", string(output.GetExtend()), secret, err)
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
							atomic.AddInt64(&userCount, 1)
						}
						common.Infof("same connect.modify user.uid=old %d,new %d,userCount=%d", info.uid, claims.Uid, userCount)
						info.uid = claims.Uid
						mapUser.Store(info.uid, info)
					}
				}
			} else { //没有同进程用户，新用户
				if 0 != info.uid { //新用户，同连接
					common.Infof("same connect.modify user.uid=old %d,new %d,userCount=%d", info.uid, claims.Uid, userCount)
					mapUser.Delete(info.uid)
					info.uid = claims.Uid
					mapUser.Store(info.uid, info)
				} else { //新用户，新连接
					info.uid = claims.Uid
					mapUser.Store(info.uid, info)
					atomic.AddInt64(&userCount, 1)
					common.Infof("add connect.uid=%d,userCount=%d.%v", info.uid, userCount, info)
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
	app, ok := mapApp[reqOut.GetApp()]
	if !ok {
		common.Errorf("fail to get app name ", reqOut.GetApp())
		err = errors.New("fail to get app name")
		return
	}

	mapAppS, ok := mapServer[app]
	if !ok {
		common.Errorf("fail to get app name ", reqOut.GetApp())
		err = errors.New("fail to get app name ")
		return
	}

	mapSt, ok := mapAppS[reqOut.GetServer()]
	if !ok {
		common.Errorf("fail to get app name ", reqOut.GetApp())
		err = errors.New("fail to get app name")
		return
	}

	return fmt.Sprintf("%s.%s.%sObj", app, mapSt, mapSt), nil
}

func (info *stTarsTcpProxy) Close() {
	if tempV, ok := mapUser.Load(info.uid); ok {
		if v, ok := tempV.(*stTarsTcpProxy); ok {
			if CONNECT_CLOSE == v.isExit {
				common.Infof("close.uid=%d,userCount=%d", info.uid, userCount)
				mapUser.Delete(info.uid)
				atomic.AddInt64(&userCount, -1)
			}
		}
	}
	common.Infof("close.uid=%d,userCount=%d,%v", info.uid, userCount, info)
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
