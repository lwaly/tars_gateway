package proxy_tars

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lwaly/tars_gateway/common"
	. "github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/protocol"
	"github.com/lwaly/tars_gateway/util"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/gofrs/flock"
	"github.com/golang/protobuf/proto"
	"golang.org/sync/syncmap"
)

const BITS = 2048
const CMD_LOGOUT = 1999
const CMD_HEART uint32 = 101

type StTarsTcpProxy struct {
	privateKey *rsa.PrivateKey
	uid        uint64
	reader     chan []byte
	mapServer  syncmap.Map
	isExit     int
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
	mapUser.Store(uint64(0), &StTarsTcpProxy{reader: make(chan []byte)})

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

func (info *StTarsTcpProxy) InitProxy() {
	util.InitQueue(util.HandlerQueueFunc(HandleQueue))
}

func (info *StTarsTcpProxy) HandleReq(body, reqTemp interface{}) (output, reqOut interface{}, err error) {
	var ok bool
	reqOutTemp, ok := reqTemp.(protocol.MsgHead)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}
	key := "1"
	common.Infof("begin msg.server=%d.cmd=%d.Encrypt=%d.RouteId=%d.seq=%d.uid=%d",
		reqOutTemp.GetServer(), reqOutTemp.GetServant(), reqOutTemp.GetEncrypt(), reqOutTemp.GetRouteId(), reqOutTemp.GetSeq(), key)
	if 0 == reqOutTemp.GetBodyLen() && CMD_HEART == reqOutTemp.GetServant() {
		common.Infof("heart msg")
		return
	}

	var rspBody protocol.MsgBody

	if nil != body {
		rspBody, ok = body.(protocol.MsgBody)
		if !ok {
			common.Errorf("fail to convert head")
			err = errors.New("fail to convert head")
			return
		}
	}

	if (1 == reqOutTemp.GetEncrypt()) && (nil != info.privateKey) && (0 < reqOutTemp.GetBodyLen()) {
		if rspBody.Body, err = util.Decrypt(rspBody.Body, info.privateKey); nil != err {
			common.Errorf("fail to Decrypt msg")
			err = errors.New(ErrUnknown)
			return
		}
	} else if (2 == reqOutTemp.GetEncrypt()) && (nil != info.privateKey) && (0 < reqOutTemp.GetBodyLen()) {
		if rspBody.Body, err = util.DecryptPkcs(rspBody.Body, info.privateKey); nil != err {
			common.Errorf("fail to Decrypt msg")
			err = errors.New(ErrUnknown)
			return
		}
	} else if (3 == reqOutTemp.GetEncrypt()) && (reqOutTemp.GetBodyLen() > 190) && (nil != info.privateKey) {
		rspBody.Body = rspBody.GetExtend()
	} else if nil == info.privateKey || 0 == reqOutTemp.GetBodyLen() {
		common.Infof("do not verify or empty body.%d", reqOutTemp.GetBodyLen())
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
		var appObj *protocol.Server
		addr := fmt.Sprintf("%d%d", IPString2Long(point.Host), point.Port)
		appObjTemp, ok := info.mapServer.Load(addr)
		if !ok {
			common.Infof("first.%d", key)
			appObj = new(protocol.Server)
			comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)
			info.mapServer.Store(addr, appObj)
		} else {
			appObj, ok = appObjTemp.(*protocol.Server)
			if !ok {
				common.Infof("first.%d", key)
				info.mapServer.Delete(addr)
				appObj = new(protocol.Server)
				comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)
				info.mapServer.Store(addr, appObj)
			}
		}

		input := protocol.Request{Version: reqOutTemp.GetVersion(), Servant: reqOutTemp.GetServant(), Seq: reqOutTemp.GetSeq(), Uid: 1, Body: rspBody.GetBody()}
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

func (info *StTarsTcpProxy) HandleRsp(output, reqOut interface{}) (outHeadRsp []byte, err error) {
	var ok bool
	reqOutTemp, ok := reqOut.(protocol.MsgHead)
	if !ok {
		common.Errorf("fail to convert head")
		err = errors.New("fail to convert head")
		return
	}

	outputTemp, ok := output.(protocol.Respond)
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

func (info *StTarsTcpProxy) Verify() (err error) {
	return
}

func (info *StTarsTcpProxy) IsExit() int {
	return info.isExit
}

func (info *StTarsTcpProxy) verify(output *protocol.Respond) (err error) {
	//扩展字段有值，认为是客户端重新认证，更换秘钥
	if 0 != len(output.GetExtend()) {
		if claims, err := util.TokenAuth(string(output.GetExtend()), secret); nil != err {
			common.Errorf("authentication token fail.%v.%v.%v", string(output.GetExtend()), secret, err)
			return err
		} else {
			if tempV, ok := mapUser.Load(claims.Userid); ok {
				if v, ok := tempV.(*StTarsTcpProxy); ok {
					if v.uid != info.uid && 0 != info.uid {
						mapUser.Delete(info.uid)
					}
					info.uid = v.uid
					mapUser.Store(info.uid, info)
				}
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

func (info *StTarsTcpProxy) objFind(reqOut *protocol.MsgHead) (strObj string, err error) {
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

func (info *StTarsTcpProxy) Close() {
	_, ok := mapUser.Load(info.uid)
	if ok {
		mapUser.Delete(info.uid)
	}
}

func HandleQueue(b []byte) {
	req := protocol.LoginNotifyReq{}
	err := proto.Unmarshal(b, &req)

	if nil != err {
		common.Errorf("fail Unmarshal msg.%v", err)
		return
	} else {
		tempV, ok := mapUser.Load(req.GetUid())
		if ok {
			v, ok := tempV.(*StTarsTcpProxy)
			if ok {
				common.Infof("user exit.uid=%d", req.GetUid())
				v.isExit = 1
				return
			}
		}
	}
	return
}
