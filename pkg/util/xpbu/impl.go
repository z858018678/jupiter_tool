package pbu

import (
	"sync"

	"google.golang.org/grpc/balancer"
)

type userLoadedNode struct {
	item interface{}
}

type userLoaded struct {
	items   map[int64]*userLoadedNode
	userCap int64
	mu      sync.Mutex
}

func New(userCap int64) Pbu {
	return &userLoaded{
		items:   make(map[int64]*userLoadedNode),
		userCap: userCap,
	}
}

// userGroupID = uid / userCap
func (p *userLoaded) Add(userGroupID int64, item interface{}) {
	p.items[userGroupID] = &userLoadedNode{item: item}
}

func (p *userLoaded) Get(uid int64) (interface{}, func(balancer.DoneInfo)) {
	var userGroupID = uid / p.userCap
	p.mu.Lock()
	var sc, ok = p.items[userGroupID]
	p.mu.Unlock()

	if !ok {
		return nil, func(balancer.DoneInfo) {}
	}

	return sc.item, func(balancer.DoneInfo) {}
}
