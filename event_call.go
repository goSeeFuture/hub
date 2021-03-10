package hub

type asyncEventCall struct {
	out  chan interface{}
	exec func()
}

func newEventAsyncCall(out chan interface{}, fn func(arg interface{}) Return, arg interface{}) asyncEventCall {

	return asyncEventCall{
		out: out,
		exec: func() {
			out <- asyncReturn(fn(arg))
		},
	}
}

type eventCall struct {
	exec func(arg interface{})
	arg  interface{}
}
