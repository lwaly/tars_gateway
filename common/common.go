package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	SWITCH_ON  = 1
	SWITCH_OFF = 2
)

const (
	OK          = 0
	ERR_UNKNOWN = 10000
	ERR_NO_USER = 10001
	ERR_LIMIT   = 10002
)

const (
	ErrUnknown = "未知错误"
	ErrNoUser  = "用户信息不存在"
	ErrLimit   = "限速"
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

func GetExternal() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)
}

func GetIntranetIp() {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("ip:", ipnet.IP.String())
			}
		}
	}
}

func InetAton(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	if 4 > len(bits) {
		return 0
	}
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func InetNtoa(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func IpIsInlist(ip string, list []string) (isIn bool) {
	ss := strings.Split(ip, ":")
	if 0 < len(ss) {
		ip = ss[0]
		for _, v := range list {
			if len(ip) >= len(v) {
				if ip[0:len(v)] == v {
					return true
				}
			}
		}
	}
	return false
}
