package hub

import (
	"reflect"
	"runtime"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

const stackBufferSize = 2048

// Hub 监听多个通道数据
type Hub struct {
	producer   chan producerOp
	processors *Queue

	keep    atomic.Value
	working atomic.Value
}

type producerOp struct {
	producer chan interface{}
	op       producerOpType
	cb       func()
}

type producerOpType int

const (
	addProducer    producerOpType = 1
	removeProducer producerOpType = 2
)

// NewHub 构建Hub
//
// @param recovery -1 总是恢复； 0 不恢复； >0 恢复次数
func newHub(producerLen int, recovery int, processors ...IDataProcessor) *Hub {
	hub := &Hub{
		processors: newQueue(processors),
		producer:   make(chan producerOp, producerLen),
	}

	hub.working.Store(false)
	go hub.process(recovery)
	return hub
}

// 添加生产者通道
func (h *Hub) Add(producer chan interface{}, cb func()) {
	h.producer <- producerOp{producer, addProducer, cb}
}

// 移除生产者通道
func (h *Hub) Remove(producer chan interface{}, cb func()) {
	h.producer <- producerOp{producer, removeProducer, cb}
}

func (h *Hub) process(recovery int) {
	h.working.Store(true)

	var cases []reflect.SelectCase
	backup := h.keep.Load()
	if backup == nil {
		cases = []reflect.SelectCase{
			{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(h.producer),
			},
		}
		h.keep.Store(cases)
	} else {
		// 存在备份时，直接恢复
		cases = backup.([]reflect.SelectCase)
	}

	defer func() {
		if r := recover(); r != nil {
			if stackBufferSize > 0 {
				buf := make([]byte, stackBufferSize)
				l := runtime.Stack(buf, false)
				log.Printf("%v: %s", r, buf[:l])
			} else {
				log.Printf("%v", r)
			}
		}

		flag := recovery != 0
		if recovery > 0 {
			recovery--
		}

		log.Warn().Bool("recovery", flag).Int("remain", recovery).Msg("panic recovery")

		if flag {
			go h.process(recovery)
		}
	}()

	for {
		chosen, recv, recvOK := reflect.Select(cases)
		if !recvOK {
			// remove close chan
			cases = append(cases[:chosen], cases[chosen+1:]...)
			continue
		}

		switch value := recv.Interface().(type) {
		case producerOp:
			if value.op == addProducer {
				// append case
				cases = append(cases, reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(value.producer),
				})
				if value.cb != nil {
					value.cb()
				}
				log.Trace().Int("len", len(cases)).Msg("append producer")
			} else if value.op == removeProducer {
				// remove case
				var removed bool
				for i, e := range cases {
					t, ok := e.Chan.Interface().(chan interface{})
					if ok && t == value.producer {
						cases = append(cases[:i], cases[i+1:]...)
						removed = true
						break
					}
				}
				log.Trace().Int("len", len(cases)).Bool("removed", removed).Bool("cb", value.cb != nil).Msg("remove producer")
				if removed && value.cb != nil {
					value.cb()
				}
			}
		default:
			// 执行高风险调用前，先备份一次
			h.keep.Store(cases)
			// 调用自定义处理
			data := recv.Interface()
			cursor := h.processors.Cursor()
			for data != nil && cursor.Next() {
				data = cursor.Value().OnData(data)
			}
		}
	}
}

// 停止
func (h *Hub) Stop() {
	// 关闭add通道
	close(h.producer)
	h.working.Store(false)
}

// 工作中
func (h *Hub) IsWorking() bool {
	return h.working.Load().(bool)
}
