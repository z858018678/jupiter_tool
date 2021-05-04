package pbu

import "google.golang.org/grpc/balancer"

type Pbu interface {
	// Next returns next selected item.
	Get(int64) (interface{}, func(balancer.DoneInfo))
	// Add a item.
	Add(int64, interface{})
}
