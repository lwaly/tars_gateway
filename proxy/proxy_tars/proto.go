package proxy_tars

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"

	"github.com/lwaly/tars_gateway/common"

	proto "github.com/golang/protobuf/proto"
)

const TIMEOUT = "tcp_timeout"

type ProtoProtocol struct {
}

func (outInfo *ProtoProtocol) NewCodec(conn net.Conn, timeOut, heartbeat int) interface{} {
	return &protoProtocol{
		conn:      conn,
		timeOut:   time.Duration(timeOut),
		heartbeat: int64(heartbeat),
		endtime:   time.Now().Unix(),
		IsClose:   false,
	}
}
func (outInfo *ProtoProtocol) Receive(info interface{}) (interface{}, interface{}, error, int) {
	tempInfo, ok := info.(*protoProtocol)
	if !ok {
		common.Errorf("fail to convert")
		return nil, nil, errors.New("fail to convert"), 0
	}
	return tempInfo.Receive()
}
func (outInfo *ProtoProtocol) Send(info interface{}, b interface{}) error {
	tempInfo, ok := info.(*protoProtocol)
	if !ok {
		common.Errorf("fail to convert")
		return errors.New("fail to convert")
	}
	return tempInfo.Send(b)
}
func (outInfo *ProtoProtocol) Close(info interface{}) error {
	tempInfo, ok := info.(*protoProtocol)
	if !ok {
		common.Errorf("fail to convert")
		return errors.New("fail to convert")
	}
	return tempInfo.Close()
}
func (outInfo *ProtoProtocol) NoticeClose(info interface{}) {
	tempInfo, ok := info.(*protoProtocol)
	if !ok {
		common.Errorf("fail to convert")
		return
	}
	tempInfo.NoticeClose()
	return
}

type protoProtocol struct {
	preBuf    bytes.Buffer //加入发生缓存区
	conn      net.Conn
	timeOut   time.Duration
	heartbeat int64
	endtime   int64
	IsClose   bool
}

func (sc *protoProtocol) Receive() (interface{}, interface{}, error, int) {
	var headLength [44]byte
HEAD_LENGTH:
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeOut * time.Millisecond))
	n, err := io.ReadFull(sc.conn, headLength[:])

	if sc.IsClose {
		return nil, nil, nil, 0
	}
	//检测心跳超时是否有消息到来
	if (0 == n) && (io.EOF.Error() != err.Error()) {
		if (sc.endtime + sc.heartbeat) > time.Now().Unix() {
			goto HEAD_LENGTH
		} else {
			common.Errorf("err:%v %v %v", sc.timeOut, sc.heartbeat, err)
			return nil, nil, errors.New(TIMEOUT), 0
		}
	}

	if 0 == n && err != nil {
		if io.EOF.Error() != err.Error() {
			common.Errorf("err:%v %d", err, n)
		} else {
			common.Infof("err:%v %d", err, n)
		}
		return nil, nil, err, 0
	}

	sc.endtime = time.Now().Unix()
	var reqHead MsgHead

	if err = proto.Unmarshal(headLength[:], &reqHead); err != nil {
		common.Errorf("err:%v", err)
		return nil, nil, err, 0
	}

	if 0 == reqHead.GetBodyLen() {
		sc.endtime = time.Now().Unix()
		return reqHead, nil, nil, len(headLength)
	}

	var head [5]byte
HEAD:
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeOut * time.Millisecond))
	n, err = io.ReadFull(sc.conn, head[:])

	if sc.IsClose {
		return nil, nil, nil, 0
	}
	//检测心跳超时是否有消息到来
	if (0 == n) && (io.EOF.Error() != err.Error()) {
		if (sc.endtime + sc.heartbeat) > time.Now().Unix() {
			goto HEAD
		} else {
			common.Errorf("err:%v", err)
			return nil, nil, errors.New(TIMEOUT), 0
		}
	}

	if 0 == n && err != nil {
		common.Errorf("err:%v %d", err, n)
		return nil, nil, err, 0
	}

	sc.endtime = time.Now().Unix()

	if err = proto.Unmarshal(append(headLength[:], head[:]...), &reqHead); err != nil {
		common.Errorf("err:%v", err)
		return nil, nil, err, 0
	}

	buf := make([]byte, reqHead.GetBodyLen())
BODY:
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeOut * time.Millisecond))
	n, err = io.ReadFull(sc.conn, buf)

	if sc.IsClose {
		return nil, nil, nil, 0
	}

	//检测心跳超时是否有消息到来
	if 0 == n {
		if (sc.endtime + sc.heartbeat) > time.Now().Unix() {
			goto BODY
		} else {
			common.Errorf("err:%v", err)
			return nil, nil, errors.New("connect timeout"), 0
		}
	}

	if err != nil {
		common.Errorf("err:%v %d", err, n)
		return nil, nil, err, 0
	}

	var reqBody MsgBody
	if err = proto.Unmarshal(buf[:], &reqBody); err != nil {
		common.Errorf("err:%v", err)
		return nil, nil, err, 0
	}

	sc.endtime = time.Now().Unix()
	return reqHead, reqBody, nil, len(headLength) + len(head) + len(buf)
}

func (sc *protoProtocol) Send(msg interface{}) error {
	//在预处理缓冲区写好东西先
	sc.preBuf.Reset()
	sc.preBuf.Write(msg.([]byte))

	buff := sc.preBuf.Bytes()
	//数据写入
	sc.conn.SetWriteDeadline(time.Now().Add(sc.timeOut * time.Millisecond))
	_, err := sc.conn.Write(buff)
	if err != nil {
		return err
	}
	sc.endtime = time.Now().Unix()
	return nil
}

func (sc *protoProtocol) Close() error {
	return sc.conn.Close()
}

func (sc *protoProtocol) NoticeClose() {
	sc.IsClose = true
}
