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

	g.AfterFunc(time.Second, interval(g, tm))

	// wait finished
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}

// 每秒打印一次和开始时间之间的时差
func interval(g *hub.Group, tm time.Time) func() {
	return func() {
		fmt.Println(time.Since(tm))
		g.AfterFunc(time.Second, interval(g, tm))
	}
}
