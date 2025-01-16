package roundrobin

import (
	"sync"
	"testing"

	roundrobin "round-robin-api/internal/roundrobin"
)

func TestRoundRobin_NextFn(t *testing.T) {
	instances := []string{"instance1", "instance2", "instance3"}
	rr := &roundrobin.RoundRobin{
		Instances: instances,
		Index:     0,
		Mutex:     sync.Mutex{},
	}

	expectedOrder := []string{"instance1", "instance2", "instance3", "instance1"}
	for _, expected := range expectedOrder {
		actual := rr.Next()
		if actual != expected {
			t.Errorf("Expected %s but got %s", expected, actual)
		}
	}
}
