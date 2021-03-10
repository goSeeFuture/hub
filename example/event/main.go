package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goSeeFuture/hub"
)

type G1Processor struct {
	g2 *hub.Group
}

func (p G1Processor) Name() string {
	return "MyProcessor"
}

func (p *G1Processor) OnData(data interface{}) interface{} {
	fmt.Println("recv:", data)
	if data.(int) == 2 {
		// 向g2组发送名为`发现目标`的自定义事件，并附带数据data
		p.g2.Emit("发现目标", data)
	}

	return nil // 次处已经处理完data，不再向后传递
}

func main() {
	g2 := hub.NewGroup()
	g1 := hub.NewGroup(hub.GroupHandles(&G1Processor{g2}))
	// g2组监听自定义事件`发现目标`，并打印出事件附带的数据arg
	g2.ListenEvent("发现目标", func(arg interface{}) {
		fmt.Println("目标", arg, "已被处理！")
	})

	// g1等待数据，如果是2，则通知g2
	ch1 := make(chan interface{})
	g1.Attach(ch1)

	go func() {
		for i := 0; i < 3; i++ {
			ch1 <- i + 1
		}
	}()

	// wait finished
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
