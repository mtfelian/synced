package synced

import (
	"fmt"
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

	lockedAt   time.Time
	lockedAtMu sync.Mutex
	timeout    time.Duration
	ticker     *time.Ticker
	closeC     chan struct{}

	lockTag *string

	BeforeLock         func()
	AfterLock          func()
	BeforeUnlock       func()
	AfterUnlock        func()
	AfterUnlockRecover func(r interface{})
}

func printStackTrace(b []byte) { log.Println("StackTrace: " + string(b)) }

func (m *Mutex) defaultCallback(event, mname string, p MutexParams) {
	var tagInfo string
	if m.lockTag != nil {
		tagInfo = fmt.Sprintf(" (tag=%q)", *m.lockTag)
	}
	log.Printf("%s for %s %s%s", event, mname, m.Name, tagInfo)
	if p.AddStackTrace {
		printStackTrace(debug.Stack())
	}
}

func (m *Mutex) defaultCallback1(event, mname string, p MutexParams, r interface{}) {
	var tagInfo string
	if m.lockTag != nil {
		tagInfo = fmt.Sprintf(" (tag=%q)", *m.lockTag)
	}
	log.Printf("%s for %s %s%s: %v", event, mname, m.Name, tagInfo, r)
	if p.AddStackTrace {
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
	m := &Mutex{Name: p.Name}
	haveWarningTimeout := p.Timeout > 0
	if haveWarningTimeout {
		m.ticker = time.NewTicker(p.Timeout)
		m.closeC = make(chan struct{})
	}
	if p.SetDefaultCallbacks {
		m.BeforeLock = func() { m.defaultCallback("BeforeLock", mname, p) }
		m.AfterLock = func() {
			m.defaultCallback("AfterLock", mname, p)
			if haveWarningTimeout {
				m.closeC = make(chan struct{})
				func() {
					m.lockedAtMu.Lock()
					defer m.lockedAtMu.Unlock()
					m.lockedAt = time.Now()
				}()
				go func() {
					for {
						select {
						case <-m.ticker.C:
							var tagInfo string
							if m.lockTag != nil {
								tagInfo = fmt.Sprintf(" (tag=%q)", *m.lockTag)
							}
							var lockedAtValue time.Time
							func() {
								m.lockedAtMu.Lock()
								defer m.lockedAtMu.Unlock()
								lockedAtValue = m.lockedAt
							}()

							if !lockedAtValue.IsZero() {
								duration := time.Now().Sub(lockedAtValue)
								if duration >= m.timeout {
									log.Printf("%s %s%s is locked for %s", mname, p.Name, tagInfo, duration)
								}
							}
						case <-m.closeC:
							return
						}
					}
				}()
				m.ticker.Reset(m.timeout)
			}
		}
		m.BeforeUnlock = func() {
			if haveWarningTimeout {
				m.ticker.Stop()
				func() {
					m.lockedAtMu.Lock()
					defer m.lockedAtMu.Unlock()
					m.lockedAt = time.Time{}
				}()
				m.closeC <- struct{}{}
			}
			m.defaultCallback("BeforeUnlock", mname, p)
		}
		m.AfterUnlock = func() { m.defaultCallback("AfterUnlock", mname, p) }
		m.AfterUnlockRecover = func(r interface{}) { m.defaultCallback1("AfterUnlockRecover", mname, p, r) }
	}
	return m
}

func (m *Mutex) lock(tag *string) {
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
		if tag != nil {
			m.lockTag = tag
		}
		if m.AfterLock != nil {
			m.AfterLock()
		}
	}()
}

// Lock calls the underlying Mutex.Lock method. BeforeLock and AfterLock callbacks will be executed
// before and after such call respectively. If callback was not specified, it will be ignored.
func (m *Mutex) Lock() { m.lock(nil) }

// LockWithTag works like Lock but adds a specified tag to help in debugging process
func (m *Mutex) LockWithTag(tag string) { m.lock(&tag) }

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
	m.lockTag = nil
	m.mu.Unlock()

	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.AfterUnlock != nil {
			m.AfterUnlock()
		}
	}()
}
