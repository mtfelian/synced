package synced

import (
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// Mutex adds debugging-related functionality to sync.Mutex.
// It is coded on top of sync.RWMutex to minimize code duplication.
type Mutex struct {
	mu          sync.RWMutex
	callbacksMu sync.Mutex
	Name        string

	lockedAt time.Time
	timeout  time.Duration
	ticker   *time.Ticker
	closeC   chan struct{}

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

// MutexParams are mutex parameters
type MutexParams struct {
	Name                string
	SetDefaultCallbacks bool
	AddStackTrace       bool
	Timeout             time.Duration
}

// NewMutex returns a pointer to a new Mutex with default callbacks assigned
func NewMutex(p MutexParams) *Mutex {
	const mname = "Mutex"
	m := &Mutex{Name: p.Name, ticker: time.NewTicker(p.Timeout), closeC: make(chan struct{})}
	if p.SetDefaultCallbacks {
		m.BeforeLock = func() { defaultMutexCallback("BeforeLock", mname, p.Name, p.AddStackTrace) }
		m.AfterLock = func() {
			defaultMutexCallback("AfterLock", mname, p.Name, p.AddStackTrace)
			m.closeC = make(chan struct{})
			m.lockedAt = time.Now()
			go func() {
				for {
					select {
					case <-m.ticker.C:
						log.Printf("%s %s is locked for %s", mname, p.Name, time.Now().Sub(m.lockedAt))
					case <-m.closeC:
						return
					}
				}
			}()
			m.ticker.Reset(m.timeout)
		}
		m.BeforeUnlock = func() {
			m.ticker.Stop()
			m.lockedAt = time.Time{}
			m.closeC <- struct{}{}
			defaultMutexCallback("BeforeUnlock", mname, p.Name, p.AddStackTrace)
		}
		m.AfterUnlock = func() { defaultMutexCallback("AfterUnlock", mname, p.Name, p.AddStackTrace) }
		m.AfterUnlockRecover = func(r interface{}) { defaultMutexCallback1("AfterUnlockRecover", mname, p.Name, p.AddStackTrace, r) }
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
