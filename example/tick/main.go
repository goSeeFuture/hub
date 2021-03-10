package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goSeeFuture/hub"
)

func main() {
	g := hub.NewGroup()
	tm := time.Now()
	g.Tick(time.Second, intervalLess10s(tm))

	// wait finished
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}

// 每秒打印一次和开始时间之间的时差，超过10秒，则终止
func intervalLess10s(tm time.Time) func() bool {
	return func() bool {
		since := time.Since(tm)
		if since.Seconds() > 10 {
			return false // 终止
		}

		fmt.Println(since)
		return true // 继续下一次定时调用
	}
}
