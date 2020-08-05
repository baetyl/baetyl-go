package mqtt

import (
	"sync"
)

type Counter struct {
	next  ID
	mutex sync.Mutex
}

// NewCounter creates a new counter
func NewCounter() *Counter {
	return NewCounterWithNext(1)
}

// NewIDCounterWithNext returns a new counter that will emit the specified if
// id as the next id.
func NewCounterWithNext(next ID) *Counter {
	return &Counter{
		next: next,
	}
}

// NextID will return the next id and increase id
func (c *Counter) NextID() ID {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// cache next id
	id := c.next

	// increment id
	c.next = NextCounterID(id)

	return id
}

// GetNextID will return the next id without increment
func (c *Counter) GetNextID() ID {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.next
}

// Reset will reset the counter.
func (c *Counter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// reset counter
	c.next = 1
}

func NextCounterID(id ID) ID {
	id++
	if id == 0 {
		id++
	}
	return id
}
