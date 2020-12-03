package synced

import (
	"encoding/json"
	"sync"
)

// Flag that is thread-safe
type Flag struct {
	state bool
	sync.Mutex
}

// NewFlag returns a new synced flag initialized by initialValue
func NewFlag(initialState bool) Flag { return Flag{initialState, sync.Mutex{}} }

// Set the flag
func (f *Flag) Set() {
	f.Lock()
	f.state = true
	f.Unlock()
}

// Unset the flag
func (f *Flag) Unset() {
	f.Lock()
	f.state = false
	f.Unlock()
}

// Get returns current flag state
func (f *Flag) Get() bool {
	f.Lock()
	defer f.Unlock()
	return f.state
}

// MarshalJSON implements json.Marshaler
func (f *Flag) MarshalJSON() ([]byte, error) {
	f.Lock()
	defer f.Unlock()
	return json.Marshal(f.state)
}

// UnmarshalJSON implements json.Unmarshaler
func (f *Flag) UnmarshalJSON(data []byte) error {
	var state bool
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	f.Lock()
	f.state = state
	f.Unlock()
	return nil
}
