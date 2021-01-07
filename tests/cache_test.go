package tests

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/lwaly/tars_gateway/util"
)

func TestCache(t *testing.T) {
	obj := "test"
	util.InitCache(obj+"."+obj, "*/10 * * * * ?",
		time.Duration(10000)*time.Millisecond, 100)
	go func() {
		for i := 1; i < 30; i++ {
			util.CacheTcpAdd(obj+"."+obj+"."+obj, strconv.FormatInt(int64(i), 10),
				[]byte(strconv.FormatInt(int64(i), 10)))
			// fmt.Println("add", )
			time.Sleep(1 * time.Second)
		}
	}()

	time.Sleep(3 * time.Second)
	for i := 1; i < 30; i++ {
		if err, v := util.CacheTcpGet(obj+"."+obj+"."+obj, strconv.FormatInt(int64(i), 10)); nil == err {
			J, _ := strconv.ParseInt(string(v.([]byte)), 10, 64)
			fmt.Println("get", J)
		} else {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(23 * time.Second)
}
