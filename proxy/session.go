package proxy

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/lwaly/tars_gateway/protocol"
)

var gSessionId uint64 = 0

//会话信息
type Session struct {
	id           uint64
	codec        protocol.Protocol
	controller   Controller
	conn         net.Conn
	stProxyConf  *StProxyConf
	snapshot     interface{}
	sendCh       chan interface{}
	sendLock     sync.RWMutex
	recvLock     sync.Mutex
	closeFlag    int32
	closeCh      chan int
	closeLock    sync.Mutex
	heartbeat    int
	rateLimitObj string
}
type (
	//会话处理器
	Handler interface {
		HandleSession(session *Session)
	}

	HandlerFunc func(session *Session)
)

func (hf HandlerFunc) HandleSession(session *Session) {
	hf(session)
}

//新建会话
func NewSession(codec protocol.Protocol, controller Controller, conn net.Conn, stProxyConf *StProxyConf) *Session {
	session := &Session{
		id:          atomic.AddUint64(&gSessionId, 1),
		codec:       codec,
		controller:  controller,
		closeCh:     make(chan int),
		conn:        conn,
		stProxyConf: stProxyConf,
	}
	return session
}

func (session *Session) Receive(info interface{}) (interface{}, interface{}, error, int) {
	session.recvLock.Lock()
	defer session.recvLock.Unlock()

	msg, req, err, n := session.codec.Receive(info)
	if err != nil {
		session.Close(info)
	}
	return msg, req, err, n
}

//同步发送
func (session *Session) Send(info, msg interface{}) error {
	if session.sendCh == nil {
		if session.IsSessionClosed() {
			return ErrSessionClosed
		}

		session.sendLock.Lock()
		defer session.sendLock.Unlock()

		err := session.codec.Send(info, msg)
		if err != nil {
			session.Close(info)
		}
		return err
	}

	session.sendLock.RLock()
	if session.IsSessionClosed() {
		session.sendLock.RUnlock()
		return ErrSessionClosed
	}

	select {
	case session.sendCh <- msg:
		session.sendLock.RUnlock()
		return nil
	default:
		session.sendLock.RUnlock()
		session.Close(info)
		return ErrSessionClosed
	}
}

//异步发送
func (session *Session) asyncSend() {
	defer session.Close(nil)
	for {
		select {
		case data, ok := <-session.sendCh:
			if !ok || session.codec.Send(nil, data) != nil { //发送失败
				return
			}
		case <-session.closeCh: //主动关闭
			return
		}
	}
}

func (session *Session) IsSessionClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) Close(info interface{}) error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		close(session.closeCh)

		if session.sendCh != nil {
			session.sendLock.Lock()
			close(session.sendCh)
			session.sendLock.Unlock()
		}

		//关闭协议传输
		err := session.codec.Close(info)
		return err
	}
	return ErrSessionClosed
}

func (session *Session) NoticeClose(info interface{}) {
	session.codec.NoticeClose(info)
}
