package hub

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type tGroupDelegateHandle struct{}

func (tGroupDelegateHandle) Name() string {
	return "tGroupDelegateHandle"
}
func (tGroupDelegateHandle) OnData(data interface{}) interface{} {
	if data.(int) == 2 {
		panic("测试异常恢复")
	}
	fmt.Println("[gd]on custom data:", data)
	return data
}

type tGroupHandle struct{}

func (tGroupHandle) Name() string {
	return "tGroupHandle"
}
func (tGroupHandle) OnData(data interface{}) interface{} {
	fmt.Printf("[g]on custom data: %v", data)
	return data
}

func Test_groupDelegate(t *testing.T) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})

	// t.Run("处理自定义数据", func(t *testing.T) {
	// 	producer := make(chan interface{}, 4)
	// 	go func() {
	// 		for n := 1; n <= 10; n++ {
	// 			producer <- n
	// 		}
	// 	}()

	// 	gd := newGroupDelegate(func(data interface{}) {
	// 		t.Logf("on custom data: %v", data)
	// 	})
	// 	gd.Attach(producer)

	// 	time.Sleep(time.Second)
	// })

	t.Run("委托处理", func(t *testing.T) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		producer := make(chan interface{}, 4)
		go func() {
			for n := 1; n <= 15; n++ {
				producer <- n
				t.Log("写入", n)
				time.Sleep(time.Second)
			}
		}()

		gd := newGroupDelegate(GroupName("gd"), GroupRecovery(1), GroupHandles(&tGroupDelegateHandle{}))

		g := NewGroup(GroupName("g"), GroupHandles(&tGroupHandle{}))

		gd.Attach(producer)

		time.Sleep(time.Second * 4)

		gd.DelegateChan(producer, g)

		time.Sleep(time.Second * 3)

		gd.SelfSupport(producer, g)
		time.Sleep(time.Second * 3)

		gd.DelegateChan(producer, g)
		time.Sleep(time.Second * 3)

		gd.SelfSupport(producer, g)
		time.Sleep(time.Second * 5)

	})
}
