package hub

import (
	"errors"
	"runtime"

	"github.com/rs/zerolog/log"
)

type asyncReturn Return

// 异步返回值
type Return struct {
	Value    interface{}
	Error    error
	callback func(Return)
	out      chan interface{}
}

type asyncCall struct {
	out  chan interface{} // 接收返回值
	exec func()           // 执行异步方法
}

// 构建异步调用
func newAsyncCall(
	fn func(arg interface{}) Return,
	arg interface{},
	callback func(arg Return),
) asyncCall {

	var out chan interface{}
	if callback == nil {
		out = nil
	} else {
		out = make(chan interface{}, 1)
	}

	ac := asyncCall{
		out: out,
		exec: func() {
			go asyncExec(out, fn, arg, callback)
		},
	}

	return ac
}

// 包裹回调，接收放回值，并释放out通道
func recvAsyncReturn(
	out chan interface{},
	callback func(arg Return),
) func() {

	return func() {
		v, ok := <-out
		if !ok {
			if callback != nil {
				callback(Return{Error: errors.New("return chan is closed before")})
			}
			return
		}
		if callback != nil {
			callback(Return(v.(asyncReturn)))
		}
		close(out)
	}
}

// 异步执行，panic保护释放掉out
func asyncExec(
	out chan interface{},
	fn func(arg interface{}) Return,
	arg interface{},
	recv func(Return),
) {
	if r := recover(); r != nil {
		if stackBufferSize > 0 {
			buf := make([]byte, stackBufferSize)
			l := runtime.Stack(buf, false)
			log.Printf("%v: %s", r, buf[:l])
		} else {
			log.Printf("%v", r)
		}

		close(out)
	}

	ar := fn(arg)
	ar.out = out
	ar.callback = recv
	if out != nil {
		out <- asyncReturn(ar)
	}
}
