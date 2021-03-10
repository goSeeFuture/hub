package hub

import (
	"sync/atomic"
)

type IDataProcessor interface {
	// 数据处理器名称
	Name() string
	// 数据处理函数，会加入到数据处理链中
	// 返回data，将会调用后续处理链的函数继续处理该data
	// 返回nil表示data处理结束，不再交给处理链的后续函数处理
	OnData(data interface{}) interface{}
}

// 数据处理器队列，不支持多线程写
type Queue struct {
	handlers atomic.Value
}

func newQueue(processors []IDataProcessor) *Queue {
	c := &Queue{}
	if processors == nil {
		processors = []IDataProcessor{}
	}
	c.handlers.Store(processors)
	return c
}

func indexOfHandle(processors []IDataProcessor, name string) int {
	for index, e := range processors {
		if e.Name() == name {
			return index
		}
	}

	return -1
}

// 在末尾增加handle
func (c *Queue) Append(handle IDataProcessor) bool {
	processors := c.handlers.Load().([]IDataProcessor)
	if indexOfHandle(processors, handle.Name()) != -1 {
		return false // 重名
	}

	t := make([]IDataProcessor, len(processors)+1)
	copy(t, processors)
	t[len(processors)] = handle
	c.handlers.Store(t)

	return true
}

// 把handle插入到名为name的handle前
func (c *Queue) Insert(name string, handle IDataProcessor) bool {
	processors := c.handlers.Load().([]IDataProcessor)
	t := make([]IDataProcessor, len(processors)+1)

	if indexOfHandle(processors, handle.Name()) != -1 {
		return false // 重名
	}

	index := indexOfHandle(processors, name)
	if index == -1 {
		// append
		copy(t, processors)
		t[len(processors)] = handle
	} else {
		for i := 0; i < len(t); i++ {
			if i < index {
				t[i] = processors[i]
			} else if i > index {
				t[i] = processors[i-1]
			} else {
				t[index] = handle
			}
		}
	}

	c.handlers.Store(t)
	return true
}

// 删除名为name的handle
func (c *Queue) Delete(name string) IDataProcessor {
	processors := c.handlers.Load().([]IDataProcessor)
	index := indexOfHandle(processors, name)
	if index == -1 {
		return nil
	}

	t := make([]IDataProcessor, len(processors)-1)
	for i, e := range processors {
		if i < index {
			t[i] = e
		} else if i > index {
			t[i-1] = e
		}
	}

	c.handlers.Store(t)
	return processors[index]
}

// 集合中handle的数量
func (c *Queue) Len() int {
	return len(c.handlers.Load().([]IDataProcessor))
}

// 按顺序拼接handle名称
func (c *Queue) String() string {
	var s string
	for _, handle := range c.handlers.Load().([]IDataProcessor) {
		s += handle.Name() + "/"
	}
	if s != "" {
		s = s[:len(s)-1]
	}
	return s
}

type Cursor struct {
	handlers []IDataProcessor
	index    int
}

func (c *Queue) Cursor() *Cursor {
	v := c.handlers.Load()
	return &Cursor{handlers: v.([]IDataProcessor), index: -1}
}

func (c *Cursor) Next() bool {
	if c.index >= len(c.handlers)-1 {
		return false
	}

	c.index++
	return true
}

func (c *Cursor) Value() IDataProcessor {
	return c.handlers[c.index]
}
