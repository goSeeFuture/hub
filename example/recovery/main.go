package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goSeeFuture/hub"
)

// 打印数据处理器
type PanicProcessor struct{}

func (PanicProcessor) Name() string {
	return "Panic"
}

func (PanicProcessor) OnData(data interface{}) interface{} {
	fmt.Println("data:", data)

	panic("异常了，糟糕~")
}

func recoveryN(n int) {
	g := hub.NewGroup(hub.GroupRecovery(n), hub.GroupHandles(&PanicProcessor{}))

	ch := make(chan interface{})
	g.Attach(ch)

	go func() {
		for i := 0; i < 3; i++ {
			ch <- i + 1
		}
	}()
}

func main() {
	// 没有恢复异常，故只会打印1次异常日志
	recoveryN(0)

	// 第2次异常没有恢复，故不会再出现第3次异常
	recoveryN(1)

	// 第3次异常都有恢复，故打印出3次panic
	recoveryN(-1)

	// wait finished
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
