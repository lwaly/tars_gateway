// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocol.proto

package proxy_tars

import (
	context "context"
	fmt "fmt"
	tars "github.com/TarsCloud/TarsGo/tars"
	model "github.com/TarsCloud/TarsGo/tars/model"
	requestf "github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	tools "github.com/TarsCloud/TarsGo/tars/util/tools"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ECmd int32

const (
	ECmd_UNKNOWN            ECmd = 0
	ECmd_E_LOGIN_NOTIFY_REQ ECmd = 601
)

var ECmd_name = map[int32]string{
	0:   "UNKNOWN",
	601: "E_LOGIN_NOTIFY_REQ",
}

var ECmd_value = map[string]int32{
	"UNKNOWN":            0,
	"E_LOGIN_NOTIFY_REQ": 601,
}

func (x ECmd) String() string {
	return proto.EnumName(ECmd_name, int32(x))
}

func (ECmd) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{0}
}

// The request message containing the user's name.
type Request struct {
	Version              uint32   `protobuf:"fixed32,1,opt,name=version,proto3" json:"version,omitempty"`
	Servant              uint32   `protobuf:"fixed32,2,opt,name=servant,proto3" json:"servant,omitempty"`
	Seq                  uint32   `protobuf:"fixed32,3,opt,name=seq,proto3" json:"seq,omitempty"`
	Uid                  uint64   `protobuf:"fixed64,4,opt,name=uid,proto3" json:"uid,omitempty"`
	Body                 []byte   `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Request) GetServant() uint32 {
	if m != nil {
		return m.Servant
	}
	return 0
}

func (m *Request) GetSeq() uint32 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *Request) GetUid() uint64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func (m *Request) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

// The response message containing the greetings
type Respond struct {
	Body                 []byte   `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	Extend               []byte   `protobuf:"bytes,2,opt,name=extend,proto3" json:"extend,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Respond) Reset()         { *m = Respond{} }
func (m *Respond) String() string { return proto.CompactTextString(m) }
func (*Respond) ProtoMessage()    {}
func (*Respond) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{1}
}

func (m *Respond) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Respond.Unmarshal(m, b)
}
func (m *Respond) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Respond.Marshal(b, m, deterministic)
}
func (m *Respond) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Respond.Merge(m, src)
}
func (m *Respond) XXX_Size() int {
	return xxx_messageInfo_Respond.Size(m)
}
func (m *Respond) XXX_DiscardUnknown() {
	xxx_messageInfo_Respond.DiscardUnknown(m)
}

var xxx_messageInfo_Respond proto.InternalMessageInfo

func (m *Respond) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *Respond) GetExtend() []byte {
	if m != nil {
		return m.Extend
	}
	return nil
}

// The request message containing the user's name.
type MsgHead struct {
	BodyLen              uint32   `protobuf:"fixed32,1,opt,name=body_len,json=bodyLen,proto3" json:"body_len,omitempty"`
	Version              uint32   `protobuf:"fixed32,2,opt,name=version,proto3" json:"version,omitempty"`
	App                  uint32   `protobuf:"fixed32,3,opt,name=app,proto3" json:"app,omitempty"`
	Server               uint32   `protobuf:"fixed32,4,opt,name=server,proto3" json:"server,omitempty"`
	Servant              uint32   `protobuf:"fixed32,5,opt,name=servant,proto3" json:"servant,omitempty"`
	Seq                  uint32   `protobuf:"fixed32,6,opt,name=seq,proto3" json:"seq,omitempty"`
	RouteId              uint64   `protobuf:"fixed64,7,opt,name=route_id,json=routeId,proto3" json:"route_id,omitempty"`
	Encrypt              uint32   `protobuf:"fixed32,8,opt,name=encrypt,proto3" json:"encrypt,omitempty"`
	CacheIs              uint32   `protobuf:"fixed32,9,opt,name=cache_is,json=cacheIs,proto3" json:"cache_is,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgHead) Reset()         { *m = MsgHead{} }
func (m *MsgHead) String() string { return proto.CompactTextString(m) }
func (*MsgHead) ProtoMessage()    {}
func (*MsgHead) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{2}
}

func (m *MsgHead) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgHead.Unmarshal(m, b)
}
func (m *MsgHead) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgHead.Marshal(b, m, deterministic)
}
func (m *MsgHead) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgHead.Merge(m, src)
}
func (m *MsgHead) XXX_Size() int {
	return xxx_messageInfo_MsgHead.Size(m)
}
func (m *MsgHead) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgHead.DiscardUnknown(m)
}

var xxx_messageInfo_MsgHead proto.InternalMessageInfo

func (m *MsgHead) GetBodyLen() uint32 {
	if m != nil {
		return m.BodyLen
	}
	return 0
}

func (m *MsgHead) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *MsgHead) GetApp() uint32 {
	if m != nil {
		return m.App
	}
	return 0
}

func (m *MsgHead) GetServer() uint32 {
	if m != nil {
		return m.Server
	}
	return 0
}

func (m *MsgHead) GetServant() uint32 {
	if m != nil {
		return m.Servant
	}
	return 0
}

func (m *MsgHead) GetSeq() uint32 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *MsgHead) GetRouteId() uint64 {
	if m != nil {
		return m.RouteId
	}
	return 0
}

func (m *MsgHead) GetEncrypt() uint32 {
	if m != nil {
		return m.Encrypt
	}
	return 0
}

func (m *MsgHead) GetCacheIs() uint32 {
	if m != nil {
		return m.CacheIs
	}
	return 0
}

type MsgBody struct {
	Body                 []byte   `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	Extend               []byte   `protobuf:"bytes,2,opt,name=extend,proto3" json:"extend,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgBody) Reset()         { *m = MsgBody{} }
func (m *MsgBody) String() string { return proto.CompactTextString(m) }
func (*MsgBody) ProtoMessage()    {}
func (*MsgBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{3}
}

func (m *MsgBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgBody.Unmarshal(m, b)
}
func (m *MsgBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgBody.Marshal(b, m, deterministic)
}
func (m *MsgBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgBody.Merge(m, src)
}
func (m *MsgBody) XXX_Size() int {
	return xxx_messageInfo_MsgBody.Size(m)
}
func (m *MsgBody) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgBody.DiscardUnknown(m)
}

var xxx_messageInfo_MsgBody proto.InternalMessageInfo

func (m *MsgBody) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *MsgBody) GetExtend() []byte {
	if m != nil {
		return m.Extend
	}
	return nil
}

type Errorinfo struct {
	ErrorCode            uint32   `protobuf:"varint,1,opt,name=error_code,json=errorCode,proto3" json:"error_code,omitempty"`
	ErrorInfo            []byte   `protobuf:"bytes,2,opt,name=error_info,json=errorInfo,proto3" json:"error_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Errorinfo) Reset()         { *m = Errorinfo{} }
func (m *Errorinfo) String() string { return proto.CompactTextString(m) }
func (*Errorinfo) ProtoMessage()    {}
func (*Errorinfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{4}
}

func (m *Errorinfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Errorinfo.Unmarshal(m, b)
}
func (m *Errorinfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Errorinfo.Marshal(b, m, deterministic)
}
func (m *Errorinfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Errorinfo.Merge(m, src)
}
func (m *Errorinfo) XXX_Size() int {
	return xxx_messageInfo_Errorinfo.Size(m)
}
func (m *Errorinfo) XXX_DiscardUnknown() {
	xxx_messageInfo_Errorinfo.DiscardUnknown(m)
}

var xxx_messageInfo_Errorinfo proto.InternalMessageInfo

func (m *Errorinfo) GetErrorCode() uint32 {
	if m != nil {
		return m.ErrorCode
	}
	return 0
}

func (m *Errorinfo) GetErrorInfo() []byte {
	if m != nil {
		return m.ErrorInfo
	}
	return nil
}

type ErrorRsp struct {
	Errinfo              *Errorinfo `protobuf:"bytes,1,opt,name=errinfo,proto3" json:"errinfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ErrorRsp) Reset()         { *m = ErrorRsp{} }
func (m *ErrorRsp) String() string { return proto.CompactTextString(m) }
func (*ErrorRsp) ProtoMessage()    {}
func (*ErrorRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{5}
}

func (m *ErrorRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ErrorRsp.Unmarshal(m, b)
}
func (m *ErrorRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ErrorRsp.Marshal(b, m, deterministic)
}
func (m *ErrorRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ErrorRsp.Merge(m, src)
}
func (m *ErrorRsp) XXX_Size() int {
	return xxx_messageInfo_ErrorRsp.Size(m)
}
func (m *ErrorRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_ErrorRsp.DiscardUnknown(m)
}

var xxx_messageInfo_ErrorRsp proto.InternalMessageInfo

func (m *ErrorRsp) GetErrinfo() *Errorinfo {
	if m != nil {
		return m.Errinfo
	}
	return nil
}

type LoginNotifyReq struct {
	Uid                  uint64   `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginNotifyReq) Reset()         { *m = LoginNotifyReq{} }
func (m *LoginNotifyReq) String() string { return proto.CompactTextString(m) }
func (*LoginNotifyReq) ProtoMessage()    {}
func (*LoginNotifyReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2bc2336598a3f7e0, []int{6}
}

func (m *LoginNotifyReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginNotifyReq.Unmarshal(m, b)
}
func (m *LoginNotifyReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginNotifyReq.Marshal(b, m, deterministic)
}
func (m *LoginNotifyReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginNotifyReq.Merge(m, src)
}
func (m *LoginNotifyReq) XXX_Size() int {
	return xxx_messageInfo_LoginNotifyReq.Size(m)
}
func (m *LoginNotifyReq) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginNotifyReq.DiscardUnknown(m)
}

var xxx_messageInfo_LoginNotifyReq proto.InternalMessageInfo

func (m *LoginNotifyReq) GetUid() uint64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func init() {
	proto.RegisterEnum("proxy_tars.ECmd", ECmd_name, ECmd_value)
	proto.RegisterType((*Request)(nil), "proxy_tars.Request")
	proto.RegisterType((*Respond)(nil), "proxy_tars.Respond")
	proto.RegisterType((*MsgHead)(nil), "proxy_tars.MsgHead")
	proto.RegisterType((*MsgBody)(nil), "proxy_tars.MsgBody")
	proto.RegisterType((*Errorinfo)(nil), "proxy_tars.errorinfo")
	proto.RegisterType((*ErrorRsp)(nil), "proxy_tars.error_rsp")
	proto.RegisterType((*LoginNotifyReq)(nil), "proxy_tars.login_notify_req")
}

func init() { proto.RegisterFile("protocol.proto", fileDescriptor_2bc2336598a3f7e0) }

var fileDescriptor_2bc2336598a3f7e0 = []byte{
	// 442 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x41, 0x6f, 0xd3, 0x30,
	0x14, 0x5e, 0xd6, 0x2c, 0xee, 0xde, 0x06, 0x8a, 0x8c, 0x00, 0x0f, 0x09, 0xa9, 0x8a, 0x38, 0x54,
	0x08, 0x15, 0x69, 0xc0, 0x0d, 0x71, 0x60, 0x2a, 0x2c, 0x62, 0x4b, 0x85, 0x05, 0x42, 0x9c, 0xac,
	0x2c, 0x7e, 0x2d, 0x91, 0x8a, 0xed, 0xd9, 0xe9, 0xb4, 0xfe, 0x54, 0xee, 0xfc, 0x10, 0x64, 0x27,
	0x6d, 0x0a, 0xe2, 0xb2, 0xdb, 0xfb, 0xfc, 0xbd, 0x7c, 0xef, 0x7d, 0x2f, 0x1f, 0xdc, 0x37, 0x56,
	0x37, 0xba, 0xd2, 0xcb, 0x49, 0x28, 0x28, 0x18, 0xab, 0x6f, 0xd7, 0xa2, 0x29, 0xad, 0xcb, 0x56,
	0x40, 0x38, 0x5e, 0xaf, 0xd0, 0x35, 0x94, 0x01, 0xb9, 0x41, 0xeb, 0x6a, 0xad, 0x58, 0x34, 0x8a,
	0xc6, 0x84, 0x6f, 0xa0, 0x67, 0x1c, 0xda, 0x9b, 0x52, 0x35, 0x6c, 0xbf, 0x65, 0x3a, 0x48, 0x53,
	0x18, 0x38, 0xbc, 0x66, 0x83, 0xf0, 0xea, 0x4b, 0xff, 0xb2, 0xaa, 0x25, 0x8b, 0x47, 0xd1, 0x38,
	0xe1, 0xbe, 0xa4, 0x14, 0xe2, 0x2b, 0x2d, 0xd7, 0xec, 0x60, 0x14, 0x8d, 0x8f, 0x79, 0xa8, 0xb3,
	0x37, 0x7e, 0xac, 0x33, 0x5a, 0xf5, 0x74, 0xd4, 0xd3, 0xf4, 0x11, 0x24, 0x78, 0xdb, 0xa0, 0x92,
	0x61, 0xde, 0x31, 0xef, 0x50, 0xf6, 0x3b, 0x02, 0x72, 0xe9, 0x16, 0xe7, 0x58, 0x4a, 0x7a, 0x02,
	0x43, 0xdf, 0x2b, 0x96, 0xb8, 0xdd, 0xd7, 0xe3, 0x0b, 0x54, 0xbb, 0x4e, 0xf6, 0xff, 0x76, 0x92,
	0xc2, 0xa0, 0x34, 0x66, 0xb3, 0x6f, 0x69, 0x8c, 0x1f, 0xe5, 0xcd, 0xa0, 0x0d, 0x2b, 0x13, 0xde,
	0xa1, 0x5d, 0xcf, 0x07, 0xff, 0xf5, 0x9c, 0xf4, 0x9e, 0x4f, 0x60, 0x68, 0xf5, 0xaa, 0x41, 0x51,
	0x4b, 0x46, 0x82, 0x71, 0x12, 0x70, 0x2e, 0xbd, 0x0c, 0xaa, 0xca, 0xae, 0x4d, 0xc3, 0x86, 0xad,
	0x4c, 0x07, 0xfd, 0x47, 0x55, 0x59, 0xfd, 0x40, 0x51, 0x3b, 0x76, 0xd8, 0x52, 0x01, 0xe7, 0xce,
	0x5f, 0xe7, 0xd2, 0x2d, 0xde, 0xfb, 0x4b, 0xdc, 0xe5, 0x3a, 0x39, 0x1c, 0xa2, 0xb5, 0xda, 0xd6,
	0x6a, 0xae, 0xe9, 0x53, 0x80, 0x00, 0x44, 0xa5, 0x25, 0x86, 0xcf, 0xef, 0xf1, 0x96, 0x3e, 0xd3,
	0x12, 0x7b, 0xda, 0x37, 0x77, 0x3a, 0x2d, 0x9d, 0xab, 0xb9, 0xce, 0xde, 0x76, 0x52, 0xc2, 0x3a,
	0x43, 0x5f, 0x02, 0x41, 0x1b, 0x54, 0x83, 0xce, 0xd1, 0xe9, 0xc3, 0x49, 0x9f, 0xa0, 0xc9, 0x76,
	0x24, 0xdf, 0x74, 0x65, 0xcf, 0x20, 0x5d, 0xea, 0x45, 0xad, 0x84, 0xd2, 0x4d, 0x3d, 0x5f, 0x0b,
	0xdb, 0xe7, 0xc2, 0x0b, 0xc4, 0x21, 0x17, 0xcf, 0x5f, 0x40, 0x3c, 0x3d, 0xfb, 0x29, 0xe9, 0x11,
	0x90, 0xaf, 0xc5, 0xa7, 0x62, 0xf6, 0xad, 0x48, 0xf7, 0xe8, 0x63, 0xa0, 0x53, 0x71, 0x31, 0xfb,
	0x98, 0x17, 0xa2, 0x98, 0x7d, 0xc9, 0x3f, 0x7c, 0x17, 0x7c, 0xfa, 0x39, 0xfd, 0x15, 0x9f, 0xbe,
	0xdb, 0xfc, 0x27, 0xfa, 0x1a, 0x92, 0xf3, 0x52, 0xc9, 0x25, 0xd2, 0x07, 0xbb, 0x7b, 0x74, 0x31,
	0x7e, 0xf2, 0xcf, 0x63, 0x08, 0x59, 0xb6, 0x77, 0x95, 0x84, 0xec, 0xbf, 0xfa, 0x13, 0x00, 0x00,
	0xff, 0xff, 0xf5, 0x00, 0x5d, 0xca, 0x0d, 0x03, 0x00, 0x00,
}

// This following code was generated by tarsrpc
// Gernerated from protocol.proto
type Server struct {
	s model.Servant
}

//SetServant is required by the servant interface.
func (obj *Server) SetServant(s model.Servant) {
	obj.s = s
}

//AddServant is required by the servant interface
func (obj *Server) AddServant(imp impServer, objStr string) {
	tars.AddServant(obj, imp, objStr)
}

//TarsSetTimeout is required by the servant interface. t is the timeout in ms.
func (obj *Server) TarsSetTimeout(t int) {
	obj.s.TarsSetTimeout(t)
}

type impServer interface {
	Handle(input Request) (output Respond, err error)
}

//Dispatch is used to call the user implement of the defined method.
func (obj *Server) Dispatch(ctx context.Context, val interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
	input := tools.Int8ToByte(req.SBuffer)
	var output []byte
	imp := val.(impServer)
	funcName := req.SFuncName
	switch funcName {

	case "Handle":
		inputDefine := Request{}
		if err = proto.Unmarshal(input, &inputDefine); err != nil {
			return err
		}
		res, err := imp.Handle(inputDefine)
		if err != nil {
			return err
		}
		output, err = proto.Marshal(&res)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("func mismatch")
	}
	var status map[string]string
	*resp = requestf.ResponsePacket{
		IVersion:     1,
		CPacketType:  0,
		IRequestId:   req.IRequestId,
		IMessageType: 0,
		IRet:         0,
		SBuffer:      tools.ByteToInt8(output),
		Status:       status,
		SResultDesc:  "",
		Context:      req.Context,
	}
	return nil
}

// Handle is client rpc method as defined
func (obj *Server) Handle(input Request) (output Respond, err error) {
	var _status map[string]string
	var _context map[string]string
	var inputMarshal []byte
	inputMarshal, err = proto.Marshal(&input)
	if err != nil {
		return output, err
	}
	resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = obj.s.Tars_invoke(ctx, 0, "Handle", inputMarshal, _status, _context, resp)
	if err != nil {
		return output, err
	}
	if err = proto.Unmarshal(tools.Int8ToByte(resp.SBuffer), &output); err != nil {
		return output, err
	}
	return output, nil
}
