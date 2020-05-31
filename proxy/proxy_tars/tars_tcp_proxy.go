package proxy_tars

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/protocol"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/gofrs/flock"
	"github.com/golang/protobuf/proto"
)

const BITS = 2048
const CMD_LOGOUT = 1999
const CMD_HEART uint32 = 101

type StTarsTcpProxy struct {
}

var mapApp map[uint32]string
var mapServer map[string]map[uint32]string
var addr string
var comm *tars.Communicator
var mapTcpEndpoint map[string]*tars.EndpointManager
var mapExceptionMsg map[uint32]uint32
var maxExceptionMsg int
var connCount int32

var fileLock *flock.Flock

func (info *StTarsTcpProxy) InitProxy() {
	mapTcpEndpoint = make(map[string]*tars.EndpointManager)
	mapApp = make(map[uint32]string)
	mapServer = make(map[string]map[uint32]string)
	mapExceptionMsg = make(map[uint32]uint32)

	fileLock = flock.New("/var/lock/gateway-lock.lock")

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

	app, ok := mapApp[reqOutTemp.GetApp()]
	if !ok {
		common.Errorf("fail to get app name ", reqOutTemp.GetApp())
		err = errors.New("fail to get app name")
		return
	}

	mapAppS, ok := mapServer[app]
	if !ok {
		common.Errorf("fail to get app name ", reqOutTemp.GetApp())
		err = errors.New("fail to get app name ")
		return
	}

	mapSt, ok := mapAppS[reqOutTemp.GetServer()]
	if !ok {
		common.Errorf("fail to get app name ", reqOutTemp.GetApp())
		err = errors.New("fail to get app name")
		return
	}

	obj := fmt.Sprintf("%s.%s.%sObj", app, mapSt, mapSt)

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

		common.Infof("first.%d", key)
		appObj := new(protocol.Server)
		comm.StringToProxy(fmt.Sprintf("%s@tcp -h %s -p %d", obj, point.Host, point.Port), appObj)

		input := protocol.Request{Version: reqOutTemp.GetVersion(), Servant: reqOutTemp.GetServant(), Seq: reqOutTemp.GetSeq(), Uid: 1, Body: rspBody.GetBody()}
		outputTemp, err := appObj.Handle(input)
		if err != nil {
			common.Errorf("err: %v body=%s extend=%s", err, string(outputTemp.GetBody()[:]), string(outputTemp.GetExtend()[:]))
			//user.mapServer.Delete(addr)
			err = errors.New(common.ErrUnknown)
			return outputTemp, reqOutTemp, err
		}
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

func (info *StTarsTcpProxy) Verify() error {
	return nil
}
