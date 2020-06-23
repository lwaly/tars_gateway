package protocol

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

func Proto() *ProtoProtocol {
	return &ProtoProtocol{}
}

func (s *ProtoProtocol) NewCodec(conn net.Conn, timeOut, heartbeat int) (Codec, error) {
	codec := &ProtobufCodec{
		conn:      conn,
		timeOut:   time.Duration(timeOut),
		heartbeat: int64(heartbeat),
		endtime:   time.Now().Unix(),
		IsClose:   false,
	}
	return codec, nil
}

type ProtobufCodec struct {
	preBuf    bytes.Buffer //加入发生缓存区
	conn      net.Conn
	timeOut   time.Duration
	heartbeat int64
	endtime   int64
	IsClose   bool
}

func (sc *ProtobufCodec) Receive() (interface{}, interface{}, error, int) {
	var headLength [39]byte
HEAD_LENGTH:
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeOut * time.Second))
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
	var req MsgHead

	if err = proto.Unmarshal(headLength[:], &req); err != nil {
		common.Errorf("err:%v", err)
		return nil, nil, err, 0
	}

	if 0 == req.GetBodyLen() {
		sc.endtime = time.Now().Unix()
		return nil, req, nil, len(headLength)
	}

	var head [5]byte
HEAD:
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeOut * time.Second))
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

	if err = proto.Unmarshal(append(headLength[:], head[:]...), &req); err != nil {
		common.Errorf("err:%v", err)
		return nil, nil, err, 0
	}

	buf := make([]byte, req.GetBodyLen())
BODY:
	sc.conn.SetReadDeadline(time.Now().Add(sc.timeOut * time.Second))
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

	var body MsgBody
	if err = proto.Unmarshal(buf[:], &body); err != nil {
		common.Errorf("err:%v", err)
		return nil, nil, err, 0
	}

	sc.endtime = time.Now().Unix()
	return body, req, nil, len(headLength) + len(head) + len(buf)
}

func (sc *ProtobufCodec) Send(msg interface{}) error {
	//在预处理缓冲区写好东西先
	sc.preBuf.Reset()
	sc.preBuf.Write(msg.([]byte))

	buff := sc.preBuf.Bytes()
	//数据写入
	sc.conn.SetWriteDeadline(time.Now().Add(sc.timeOut * time.Second))
	_, err := sc.conn.Write(buff)
	if err != nil {
		return err
	}
	sc.endtime = time.Now().Unix()
	return nil
}

func (sc *ProtobufCodec) Close() error {
	return sc.conn.Close()
}

func (sc *ProtobufCodec) NoticeClose() {
	sc.IsClose = true
}
