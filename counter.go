package synced

import (
	"encoding/json"
	"sync"
)

// Counter that is thread-safe
type Counter struct {
	count int
	sync.Mutex
}

// NewCounter returns a new synced counter initialized by initialValue
func NewCounter(initialValue int) Counter { return Counter{initialValue, sync.Mutex{}} }

func (c *Counter) dec()      { c.count-- }
func (c *Counter) inc()      { c.count++ }
func (c *Counter) add(i int) { c.count += i }
func (c *Counter) sub(i int) { c.count -= i }
func (c *Counter) set(i int) { c.count = i }

// Inc increases counter by 1. Returns original value
func (c *Counter) Inc() int {
	c.Lock()
	defer c.Unlock()
	v := c.count
	c.inc()
	return v
}

// Add i to counter. Returns original value
func (c *Counter) Add(i int) int {
	c.Lock()
	defer c.Unlock()
	v := c.count
	c.add(i)
	return v
}

// Dec decreases counter by 1. Returns original value
func (c *Counter) Dec() int {
	c.Lock()
	defer c.Unlock()
	v := c.count
	c.dec()
	return v
}

// Set counter to i. Returns original value
func (c *Counter) Set(i int) int {
	c.Lock()
	defer c.Unlock()
	v := c.count
	c.set(i)
	return v
}

// Get returns current counter value
func (c *Counter) Get() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}

// MarshalJSON implements json.Marshaler
func (c *Counter) MarshalJSON() ([]byte, error) {
	c.Lock()
	defer c.Unlock()
	return json.Marshal(c.count)
}

// UnmarshalJSON implements json.Unmarshaler
func (c *Counter) UnmarshalJSON(data []byte) error {
	var count int
	if err := json.Unmarshal(data, &count); err != nil {
		return err
	}
	c.Lock()
	c.count = count
	c.Unlock()
	return nil
}
