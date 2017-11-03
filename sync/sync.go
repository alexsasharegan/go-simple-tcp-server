// Package sync is modeled after https://github.com/remeh/sizedwaitgroup.
// It provides a way to limit go routines to a given size
// and adheres to the std lib sync.WaitGroup API.
package sync

import (
	"sync"
)

type empty struct{}

// SizedWaitGroup is wrapper around sync.WaitGroup
// with logic for blocked adding when the group limit is reached.
type SizedWaitGroup struct {
	wg *sync.WaitGroup
	c  chan empty
}

// NewSizedWaitGroup returns a pointer to a SizedWaitGroup.
func NewSizedWaitGroup(limit int) *SizedWaitGroup {
	return &SizedWaitGroup{
		c:  make(chan empty, limit),
		wg: &sync.WaitGroup{},
	}
}

// Add adds delta, which may be negative, to the SizedWaitGroup counter.
// It will block until a member is released if the SizedWaitGroup limit has been reached.
func (swg *SizedWaitGroup) Add(delta int) {
	for delta != 0 {
		// Handle negative delta case.
		if delta < 0 {
			<-swg.c
			swg.wg.Done()
			delta++
			continue
		}
		swg.c <- empty{}
		swg.wg.Add(1)
		delta--
	}
}

// Done decrements the SizedWaitGroup counter by one.
func (swg *SizedWaitGroup) Done() {
	swg.Add(-1)
}

// Wait blocks until the SizedWaitGroup counter is zero.
func (swg *SizedWaitGroup) Wait() {
	swg.wg.Wait()
}
