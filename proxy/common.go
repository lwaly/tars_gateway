package proxy

import (
	"sync"

	"github.com/TarsCloud/TarsGo/tars"
)

var secret = ""
var machine int
var mapKey24 map[string]int64
var mutex sync.Mutex
var g_seq uint32

func init() {
}

func GetSeq() uint32 {
	mutex.Lock()
	defer mutex.Unlock()
	g_seq++
	return g_seq
}

func InitEndpoint(mapEndpoint map[string]*tars.EndpointManager) {

}
