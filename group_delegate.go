package hub

import (
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

// 工作组，有委托能力
//
// 可以将 producer 委托给其他 Group 处理
type groupDelegate struct {
	*Group

	// 是否建立委托关系
	delegated atomic.Value
}

// 构建委托工作组
func newGroupDelegate(options ...GroupOption) *groupDelegate {
	gd := &groupDelegate{
		Group: NewGroup(options...),
	}

	gd.delegated.Store(false)

	return gd
}

// 委托其他工作组处理生产数据
func (gd *groupDelegate) DelegateChan(producer chan interface{}, g *Group) {
	if gd.IsDelegated() {
		// 已经建立其他委托关系，必须停止才能再次委托
		panic("already delegate to some group, must stop it first")
	}

	gd.DetachCB(producer, func() {
		gd.delegated.Store(true)
		log.Trace().Bool("delegated", gd.IsDelegated()).Msg("委托生效")
		g.Attach(producer)
	})
}

// 中止委托关系，并自己处理工作
func (gd *groupDelegate) SelfSupport(producer chan interface{}, g *Group) {
	if !gd.IsDelegated() {
		log.Trace().Bool("delegated", gd.IsDelegated()).Msg("没有建立委托")
		return // 没有建立委托
	}

	fn := func() {
		// 标记未委托
		gd.delegated.Store(false)
		log.Trace().Bool("delegated", gd.IsDelegated()).Msg("标记未委托")
	}
	// 重新监听生产通道
	g.DetachCB(producer, func() {
		gd.AttachCB(producer, fn)
	})
}

// 中止委托关系
func (gd *groupDelegate) StopDelegate() bool {
	if !gd.IsDelegated() {
		return false // 没有建立委托
	}

	// 标记未委托
	log.Trace().Bool("delegated", gd.IsDelegated()).Msg("中止委托关系")
	gd.delegated.Store(false)
	return true
}

// 是否已经委托
func (gd *groupDelegate) IsDelegated() bool {
	return gd.delegated.Load().(bool)
}
