package synced

import (
	"log"
	"runtime/debug"
	"sync"
)

// Mutex adds debugging-related functionality to sync.Mutex.
// It is coded on top of sync.RWMutex to minimize code duplication.
type Mutex struct {
	mu                 sync.RWMutex
	callbacksMu        sync.Mutex
	Name               string
	BeforeLock         func()
	AfterLock          func()
	BeforeUnlock       func()
	AfterUnlock        func()
	AfterUnlockRecover func(r interface{})
}

func printStackTrace(b []byte) {
	log.Println("StackTrace: " + string(b))
}
func defaultMutexCallback(event, mname, name string, addStackTrace bool) {
	log.Printf("%s for %s %s", event, mname, name)
	if addStackTrace {
		printStackTrace(debug.Stack())
	}
}
func defaultMutexCallback1(event, mname, name string, addStackTrace bool, r interface{}) {
	log.Printf("%s for %s %s: %v", event, mname, name, r)
	if addStackTrace {
		printStackTrace(debug.Stack())
	}
}

// NewMutex returns a pointer to a new Mutex with default callbacks assigned
func NewMutex(name string, setDefaultCallbacks, addStackTrace bool) *Mutex {
	const mname = "Mutex"
	m := &Mutex{Name: name}
	if setDefaultCallbacks {
		m.BeforeLock = func() { defaultMutexCallback("BeforeLock", mname, name, addStackTrace) }
		m.AfterLock = func() { defaultMutexCallback("AfterLock", mname, name, addStackTrace) }
		m.BeforeUnlock = func() { defaultMutexCallback("BeforeUnlock", mname, name, addStackTrace) }
		m.AfterUnlock = func() { defaultMutexCallback("AfterUnlock", mname, name, addStackTrace) }
		m.AfterUnlockRecover = func(r interface{}) { defaultMutexCallback1("AfterUnlockRecover", mname, name, addStackTrace, r) }
	}
	return m
}

// Lock calls the underlying Mutex.Lock method. BeforeLock and AfterLock callbacks will be executed
// before and after such call respectively. If callback was not specified, it will be ignored.
func (m *Mutex) Lock() {
	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.BeforeLock != nil {
			m.BeforeLock()
		}
	}()

	m.mu.Lock()

	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.AfterLock != nil {
			m.AfterLock()
		}
	}()
}

// Unlock calls the underlying Mutex.Unlock method. BeforeUnlock and AfterUnlock callbacks will be executed
// before and after such call respectively. If a panic will occur at underlying Mutex unlocking, it will be
// handled by a call to recover() and BeforeUnlockRecover and AfterUnlockRecover will be called respectively.
// If callback was not specified, it will be ignored.
func (m *Mutex) Unlock() {
	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.BeforeUnlock != nil {
			m.BeforeUnlock()
		}
	}()

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		func() {
			m.callbacksMu.Lock()
			defer m.callbacksMu.Unlock()
			if m.AfterUnlockRecover != nil {
				m.AfterUnlockRecover(r)
			}
		}()
	}()
	m.mu.Unlock()

	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.AfterUnlock != nil {
			m.AfterUnlock()
		}
	}()
}
