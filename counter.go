package synced

import "sync"

// Counter that is thread-safe
type Counter struct {
	count int
	sync.Mutex
}

// NewCounter returns a new synced counter initialized by initialValue
func NewCounter(initialValue int) Counter {
	return Counter{
		initialValue,
		sync.Mutex{},
	}
}

// Inc increases counter value by 1
func (c *Counter) Inc() {
	c.Lock()
	c.count++
	c.Unlock()
}

// Add adds i to a counter's value
func (c *Counter) Add(i int) {
	c.Lock()
	c.count += i
	c.Unlock()
}

// Dec decreases counter value by 1
func (c *Counter) Dec() {
	c.Lock()
	c.count--
	c.Unlock()
}

// Get returns current counter value
func (c *Counter) Get() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}
