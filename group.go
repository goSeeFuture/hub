package hub

import (
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const groupChanLen = 12

var (
	unnamegroup int64
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

// Group 监听多个通道，并支持异步串行
type Group struct {
	hub         *Hub
	processChan chan interface{}
	calls       sync.Map // map[string]func(interface{}) Return
	events      sync.Map // map[string]func(interface{})
	config      groupconfig
}

type groupconfig struct {
	Name       string
	Handles    []IDataProcessor
	ChannelLen int
	Recovery   int // -1 总是恢复； 0 不恢复； >0 恢复次数
}

type GroupOption func(gc *groupconfig)

// 组名
func GroupName(name string) func(gc *groupconfig) {
	return func(gc *groupconfig) {
		gc.Name = name
	}
}

// 数据通道长度
func GroupChannelLen(channelLen int) func(gc *groupconfig) {
	return func(gc *groupconfig) {
		gc.ChannelLen = channelLen
	}
}

// 数据处理器
func GroupHandles(processors ...IDataProcessor) func(gc *groupconfig) {
	return func(gc *groupconfig) {
		gc.Handles = processors
	}
}

// recovery
// 默认值0： -1 总是恢复； 0 不恢复； >0 恢复次数
func GroupRecovery(recovery int) func(gc *groupconfig) {
	return func(gc *groupconfig) {
		gc.Recovery = recovery
	}
}

// 构建通道聚合处理组
func NewGroup(options ...GroupOption) *Group {
	config := groupconfig{
		ChannelLen: groupChanLen,
	}
	for _, option := range options {
		option(&config)
	}

	g := &Group{
		config:      config,
		processChan: make(chan interface{}, groupChanLen),
	}
	if g.config.Name == "" {
		number := atomic.AddInt64(&unnamegroup, 1)
		g.config.Name = "Group" + strconv.FormatInt(number, 10)
	}

	if len(g.config.Handles) == 0 {
		g.hub = newHub(groupChanLen, g.config.Recovery, g)
	} else {
		processors := append([]IDataProcessor{g}, g.config.Handles...)
		g.hub = newHub(groupChanLen, g.config.Recovery, processors...)
	}

	g.Attach(g.processChan)
	return g
}

// 数据处理器名称
func (g *Group) Name() string {
	return g.config.Name
}

// 处理聚合的各个通道数据
func (g *Group) OnData(data interface{}) interface{} {
	switch x := data.(type) {
	case asyncCall:
		// 加入到hub，关注异步返回值
		g.AttachCB(x.out, func() {
			x.exec()
		})
	case asyncReturn:
		if x.callback != nil {
			x.callback(Return(x))
		}
		close(x.out)
	case asyncEventCall:
		x.exec()
	case eventCall:
		x.exec(x.arg)
	default:
		return data
	}

	return data
}

// 发送事件，给 group 中的 handler 处理
// 	返回值 registered 描述 event 是否注册了 handler
func (g *Group) Emit(event string, arg interface{}) (registered bool) {
	h, exist := g.events.Load(event)
	if !exist {
		log.Trace().Str("event", event).Msg("not register event handler")
		return false
	}

	g.processChan <- eventCall{exec: h.(func(arg interface{})), arg: arg}
	return exist
}

// 调用事件，跨协程调用group中的函数
// 	event 事件名称
// 	arg 附带参数，无参数传 nil
// 	返回值 waitResult() Return 调用后将阻塞等待事件执行完毕，hotpot.Return 包含事件处理函数的返回值
// 	返回值 registered 描述 event 是否注册了 handler
func (g *Group) Call(event string, arg interface{}) (ret Return, registered bool) {
	if isClosedChan(g.processChan) {
		return
	}

	h, exist := g.calls.Load(event)
	if !exist {
		log.Trace().Str("event", event).Msg("not register event handler")
		return
	}

	out := make(chan interface{}, 1)
	fn := func() (ret Return) {
		handler := func(arg Return) {
			ret = arg
		}

		select {
		case _, ok := <-out:
			if ok {
				close(out)
			}
			panic("handler cannot in Group goroutine")
		default:
			recvAsyncReturn(out, handler)()
		}

		return
	}

	g.processChan <- newEventAsyncCall(out, h.(func(arg interface{}) Return), arg)

	return fn(), true
}

// 绑定事件处理函数
func (g *Group) ListenEvent(event string, handler func(arg interface{})) {
	g.events.Store(event, handler) // 注册自定义事件
	log.Trace().Str("event", event).Msg("register event handler")
}

// 绑定调用处理函数
func (g *Group) ListenCall(event string, handler func(arg interface{}) Return) {
	g.calls.Store(event, handler) // 注册自定义事件
	log.Trace().Str("call", event).Msg("register event call handler")
}

// 慢调用，用协程执行fn，并将结果送回到 group 协程
func (g *Group) SlowCall(fn func(interface{}) Return, arg interface{}, callback func(Return)) {
	if isClosedChan(g.processChan) {
		return
	}
	g.processChan <- newAsyncCall(fn, arg, callback)
}

// 延时执行，超时后通过 group 协程调用 fn
func (g *Group) AfterFunc(dur time.Duration, fn func()) {
	g.SlowCall(func(_ interface{}) Return {
		<-time.After(dur)
		return Return{}
	}, nil, func(_ Return) { fn() })
}

// 每隔dur执行fn，当fn返回false时终止
func (g *Group) Tick(dur time.Duration, fn func() bool) {
	g.SlowCall(func(_ interface{}) Return {
		<-time.After(dur)
		return Return{}
	}, nil, g.tickCallback(dur, fn))
}

func (g *Group) tickCallback(dur time.Duration, fn func() bool) func(Return) {
	return func(Return) {
		if !fn() {
			return
		}

		g.SlowCall(func(_ interface{}) Return {
			<-time.After(dur)
			return Return{}
		}, nil, g.tickCallback(dur, fn))
	}
}

// 停止并释放资源
func (g *Group) Stop() {
	if !g.hub.IsWorking() {
		return
	}

	g.hub.Stop()
	close(g.processChan)
}

// 增加监听通道，同步等待，确保添加成功
func (g *Group) Attach(producer chan interface{}) {
	var wg sync.WaitGroup
	wg.Add(1)
	g.hub.Add(producer, func() {
		wg.Done()
	})
	wg.Wait()
}

// 移除监听通道 producer，同步等待，确保移除成功
func (g *Group) Detach(producer chan interface{}) {
	var wg sync.WaitGroup
	wg.Add(1)
	g.hub.Remove(producer, func() {
		wg.Done()
	})
	wg.Wait()
}

// 增加监听通道
func (g *Group) AttachCB(producer chan interface{}, cb func()) {
	g.hub.Add(producer, cb)
}

// 移除监听通道 producer
func (g *Group) DetachCB(producer chan interface{}, cb func()) {
	g.hub.Remove(producer, cb)
}

// 工作中
func (g *Group) IsWorking() bool {
	return g.hub.IsWorking()
}

// 数据处理链
func (g *Group) Processors() *Queue {
	return g.hub.processors
}
