package component

import "github/thep2p/skipgraph-go/modules"

type Manager struct {
}

func (Mang Manager) Start(ctx modules.ThrowableContext) {
	//TODO implement me
	panic("implement me")
}

func (Mang Manager) Ready() <-chan interface{} {
	//TODO implement me
	panic("implement me")
}

func (Mang Manager) Done() <-chan interface{} {
	//TODO implement me
	panic("implement me")
}

func (Mang Manager) Add(c modules.Component) {
	//TODO implement me
	panic("implement me")
}

var _ modules.ComponentManager = (*Manager)(nil)
