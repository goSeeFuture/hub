package hub

import (
	"fmt"
	"testing"
	"time"
)

func data1(ch chan interface{}) {
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
			time.Sleep(time.Second)
		}
		close(ch)
	}()
}

func data2(ch chan interface{}) {
	go func() {
		for i := 'a'; i < 'z'; i++ {
			ch <- string([]byte{byte(i)})
			time.Sleep(time.Second)
		}
		close(ch)
	}()
}

type tHandleData struct{}

func (tHandleData) Name() string {
	return "tHandleData"
}
func (tHandleData) OnData(data interface{}) interface{} {
	fmt.Println("tHandleData:", data)
	return data
}

func TestHub(t *testing.T) {
	ch1 := make(chan interface{}, 1)
	ch2 := make(chan interface{}, 1)

	h := newHub(12, 0, &tHandleData{})

	h.Add(ch1, nil)
	h.Add(ch2, nil)

	data1(ch1)
	data2(ch2)

	time.Sleep(time.Second * 30)
}

func TestRemoveProducer(t *testing.T) {
	ch1 := make(chan interface{}, 1)
	h := newHub(12, 0)
	h.Add(ch1, func() {
		t.Log("producer removed")
	})
	h.Remove(ch1, func() {
		t.Log("producer add")
	})

	time.Sleep(time.Second)
}
