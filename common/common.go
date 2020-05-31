package common

import (
	"encoding/json"
	"net"
)

const (
	OK                 = 0
	ERR_UNKNOWN        = 10000
	ERR_INPUT_DATA     = 10001
	ERR_DATABASE       = 10002
	ERR_DUP_USER       = 10003
	ERR_NO_USER        = 10004
	ERR_PASS           = 10005
	ERR_NO_USE_RPASS   = 10006
	ERR_NO_USER_CHANGE = 10007
	ERR_INVALID_USER   = 10008
	ERR_OPEN_FILE      = 10009
	ERR_WRIT_EFILE     = 10010
	ERR_SYSTEM         = 10011
	ERR_EXPIRED        = 10012
	ERR_PERMISSION     = 10013
)

const (
	ErrInputData    = "数据输入错误"
	ErrDatabase     = "数据库操作错误"
	ErrDupUser      = "用户信息已存在"
	ErrNoUser       = "用户信息不存在"
	ErrPass         = "密码不正确"
	ErrNoUserPass   = "用户信息不存在或密码不正确"
	ErrNoUserChange = "用户信息不存在或数据未改变"
	ErrInvalidUser  = "用户信息不正确"
	ErrOpenFile     = "打开文件出错"
	ErrWriteFile    = "写文件出错"
	ErrSystem       = "操作系统错误"
	ErrUnknown      = "未知错误"
)

type CodeInfo struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

func NewErrorInfo(code int, info string) []byte {
	b, _ := json.Marshal(&CodeInfo{code, info})
	return b
}

func IPString2Long(ip string) uint {
	b := net.ParseIP(ip).To4()
	if b == nil {
		return 0
	}

	return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
}
