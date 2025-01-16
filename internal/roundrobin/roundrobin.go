package roundrobin

import (
	"log"
	"sync"
)

type RoundRobin struct {
	Instances []string
	Index     int
	Mutex     sync.Mutex
}

func (rr *RoundRobin) Next() string {
	rr.Mutex.Lock()
	defer rr.Mutex.Unlock()

	instance := rr.Instances[rr.Index]
	log.Printf("Routing to instance: %s (Index: %d)", instance, rr.Index)
	rr.Index = (rr.Index + 1) % len(rr.Instances)
	return instance
}
