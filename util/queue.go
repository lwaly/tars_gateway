package util

import (
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/protocol"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

type stQueue struct {
	Addr          string `json:"addr,omitempty"`
	Cluster       string `json:"cluster,omitempty"`
	Client        string `json:"client,omitempty"`
	GroupObject   string `json:"groupObject,omitempty"`
	Durable       string `json:"durable,omitempty"`
	StartWay      string `json:"startWay,omitempty"`
	GatewayObject string `json:"gatewayObject,omitempty"`
	Machine       uint64 `json:"machine,omitempty"`
	handlerQueue  HandlerQueueFunc
	gConn         stan.Conn
	startOpt      stan.SubscriptionOption
}

var queue stQueue
var seq uint32

type HandlerQueueFunc func(b []byte)

func InitQueue(handlerQueue HandlerQueueFunc) {
	if nil == handlerQueue {
		common.Warnf("handlerQueue nil.")
		return
	}

	err := common.Conf.GetStruct("queue", &queue)
	if err != nil {
		common.Errorf("fail to parse Queue config.%v %v", err, queue)
		return
	}

	if 0 == queue.Machine {
		queue.Machine = uint64(common.InetAton(net.ParseIP(common.GetExternal())))
	}

	queue.handlerQueue = handlerQueue
	if "" == queue.Addr || "" == queue.Cluster || "" == queue.Client || "" == queue.GroupObject || "" == queue.Durable || "" == queue.StartWay || "" == queue.GatewayObject {
		common.Errorf("fail to get queue config.%v", queue)
		return
	}

	s := strings.Split(queue.StartWay, " ")
	queue.startOpt = stan.StartAt(pb.StartPosition_NewOnly)
	if 0 == strings.Compare(s[0], "startSeq") {
		t, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			common.Errorf("queue config err.%v %v", queue, err)
			return
		}
		queue.startOpt = stan.StartAtSequence(uint64(t))
	} else if 0 == strings.Compare(s[0], "deliverLast") {
		queue.startOpt = stan.StartWithLastReceived()
	} else if 0 == strings.Compare(s[0], "deliverAll") {
		queue.startOpt = stan.DeliverAllAvailable()
	} else if 0 == strings.Compare(s[0], "startDelta") {
		t, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			common.Errorf("queue config err.%v %v", queue, err)
			return
		}

		queue.startOpt = stan.StartAtTimeDelta(time.Duration(t))
	} else {
		common.Errorf("queue config err.%v", queue)
		return
	}

	connectQueue()
	return
}

func rconnect(_ stan.Conn, reason error) {
	common.Errorf("Connection lost, reason: %v", reason)
	var err error
AGAIN:
	err = connectQueue()
	if err != nil {
		common.Errorf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, queue.Addr)
		time.Sleep(time.Duration(3) * time.Second)
		goto AGAIN
	}

	return
}

func connectQueue() (err error) {
	queue.gConn, err = stan.Connect(queue.Cluster, queue.Client, stan.NatsURL(queue.Addr), stan.SetConnectionLostHandler(rconnect))
	if err != nil {
		common.Errorf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, queue.Addr)
		return
	}

	common.Infof("GroupObject.%v", queue.GroupObject)
	sGroupOb := strings.Split(queue.GroupObject, " ")
	for i := 0; i < len(sGroupOb); {
		_, err = queue.gConn.QueueSubscribe(sGroupOb[i+1], sGroupOb[i], queueHandle, queue.startOpt, stan.DurableName(queue.Durable), stan.SetManualAckMode())
		if err != nil {
			queue.gConn.Close()
			common.Errorf("fail to subscribe queue.%v %v", queue, err)
			return
		}
		i += 2
	}
	return
}

func queueHandle(msg *stan.Msg) {
	common.Infof("Sequence.%d", msg.Sequence)
	input := protocol.Request{}
	err := proto.Unmarshal(msg.Data, &input)

	if nil != err {
		common.Errorf("fail Unmarshal msg.%v", err)
	} else {
		common.Infof("cmd.%d", input.GetServant())
		if input.Uid == queue.Machine {
			common.Infof("own msg")
		} else {
			queue.handlerQueue(input.GetBody())
		}
	}
	msg.Ack()
	return
}

func QueueSend(subj string, b []byte, version, cmd uint32) (err error) {
	common.Infof("cmd.%d", cmd)
	req := protocol.Request{Version: version, Servant: cmd, Seq: atomic.AddUint32(&seq, 1), Uid: queue.Machine, Body: b}
	b, err = proto.Marshal(&req)
	if err != nil {
		common.Errorf("faile to Marshal msg.err: %v", err)
		return
	}
	err = queue.gConn.Publish(subj, b)
	if err != nil {
		common.Errorf("Error during publish: %v\n", err)
	}
	return
}
