package synced

import "sync"

// Flag that is thread-safe
type Flag struct {
	state bool
	sync.Mutex
}

// NewFlag returns a new synced flag initialized by initialValue
func NewFlag(initialState bool) Flag { return Flag{initialState, sync.Mutex{}} }

// Set the flag
func (c *Flag) Set() {
	c.Lock()
	c.state = true
	c.Unlock()
}

// Unset the flag
func (c *Flag) Unset() {
	c.Lock()
	c.state = false
	c.Unlock()
}

// Get returns current flag state
func (c *Flag) Get() bool {
	c.Lock()
	defer c.Unlock()
	return c.state
}
