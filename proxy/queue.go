package proxy

import (
	"strconv"
	"strings"
	"time"

	"github.com/lwaly/tars_gateway/common"
	"github.com/lwaly/tars_gateway/protocol"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

type stQueue struct {
	addr           string
	cluster        string
	client         string
	group_object   string
	durable        string
	start_way      string
	gateway_object string
}

var gConn stan.Conn
var queue stQueue
var startOpt stan.SubscriptionOption

func InitQueue() {
	queue.addr, _ = common.Conf.GetValue("queue", "addr")
	queue.cluster, _ = common.Conf.GetValue("queue", "cluster")
	queue.client, _ = common.Conf.GetValue("queue", "client")
	queue.group_object, _ = common.Conf.GetValue("queue", "group_object")
	queue.durable, _ = common.Conf.GetValue("queue", "durable")
	queue.start_way, _ = common.Conf.GetValue("queue", "start_way")
	queue.gateway_object, _ = common.Conf.GetValue("queue", "gateway_object")

	if "" == queue.addr || "" == queue.cluster || "" == queue.client || "" == queue.group_object || "" == queue.durable || "" == queue.start_way || "" == queue.gateway_object {
		common.Errorf("fail to get queue config.%v", queue)
		return
	}

	s := strings.Split(queue.start_way, " ")
	startOpt = stan.StartAt(pb.StartPosition_NewOnly)
	if 0 == strings.Compare(s[0], "startSeq") {
		t, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			common.Errorf("queue config err.%v %v", queue, err)
			return
		}
		startOpt = stan.StartAtSequence(uint64(t))
	} else if 0 == strings.Compare(s[0], "deliverLast") {
		startOpt = stan.StartWithLastReceived()
	} else if 0 == strings.Compare(s[0], "deliverAll") {
		startOpt = stan.DeliverAllAvailable()
	} else if 0 == strings.Compare(s[0], "startDelta") {
		t, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			common.Errorf("queue config err.%v %v", queue, err)
			return
		}

		startOpt = stan.StartAtTimeDelta(time.Duration(t))
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
		common.Errorf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, queue.addr)
		time.Sleep(time.Duration(3) * time.Second)
		goto AGAIN
	}

	return
}

func connectQueue() (err error) {
	gConn, err = stan.Connect(queue.cluster, queue.client, stan.NatsURL(queue.addr), stan.SetConnectionLostHandler(rconnect))
	if err != nil {
		common.Errorf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, queue.addr)
		return
	}

	sGroupOb := strings.Split(queue.group_object, " ")
	for i := 0; i < len(sGroupOb); {
		_, err = gConn.QueueSubscribe(sGroupOb[i+1], sGroupOb[i], queueHandle, startOpt, stan.DurableName(queue.durable), stan.SetManualAckMode())
		if err != nil {
			gConn.Close()
			common.Errorf("fail to subscribe queue.%v %v", queue, err)
			return
		}
		i += 2
	}
	return
}

func queueHandle(msg *stan.Msg) {
	common.Infof("cmd.%d", msg.Sequence)
	input := protocol.Request{}
	err := proto.Unmarshal(msg.Data, &input)

	if nil != err {
		common.Errorf("fail Unmarshal msg.%v", err)
	} else {
		common.Infof("cmd.%d %d", input.GetServant(), msg.Sequence)
		// if input.Uid == uint32(machine) {
		// 	common.Infof("own msg")
		// } else {
		// 	switch input.GetServant() {
		// 	case uint32(protocol.ECmd_E_LOGIN_NOTIFY_REQ):
		// 		HandleToken(&input)
		// 	}
		// }
	}
	msg.Ack()
	return
}

func queueSend(subj string, b []byte, cmd uint32) (err error) {
	common.Infof("cmd.%d", cmd)
	req := protocol.Request{Version: 1, Servant: cmd, Seq: GetSeq(), Uid: uint32(machine), Body: b}
	b, err = proto.Marshal(&req)
	if err != nil {
		common.Errorf("faile to Marshal msg.err: %v", err)
		return
	}
	err = gConn.Publish(subj, b)
	if err != nil {
		common.Errorf("Error during publish: %v\n", err)
	}
	return
}
