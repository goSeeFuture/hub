package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		tm := time.Now()                                                   // 等待g2调用`发现目标`的返回
		ret, _ := p.g2.Call("发现目标", data)                                  // 调用g2组绑定的`发现目标`
		fmt.Println("call spend:", time.Since(tm), ", return:", ret.Value) // 计算本次跨协程调用耗费的事件
	}

	return nil // 次处已经处理完data，不再向后传递
}

func main() {
	g2 := hub.NewGroup()
	g1 := hub.NewGroup(hub.GroupHandles(&G1Processor{g2}))
	// 绑定`发现目标`调用的实现函数
	g2.ListenCall("发现目标", func(arg interface{}) hub.Return {
		fmt.Println("目标", arg, "已被处理！")
		time.Sleep(time.Second) // 延时1秒，模拟耗时操作
		return hub.Return{Value: "ok"}
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
	fmt.Println("按 ctrl+c 退出程序")
	<-ch
}
