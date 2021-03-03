package synced

// RWMutex adds debugging-related functionality to sync.RWMutex
type RWMutex struct {
	*Mutex

	BeforeRLock         func()
	AfterRLock          func()
	BeforeRUnlock       func()
	AfterRUnlock        func()
	AfterRUnlockRecover func(r interface{})
}

// NewRWMutex returns a pointer to a new RWMutex with default callbacks assigned
func NewRWMutex(p MutexParams) *RWMutex {
	const mname = "RWMutex"
	m := &RWMutex{Mutex: NewMutex(p)}
	if p.SetDefaultCallbacks {
		m.BeforeRLock = func() { m.defaultCallback("BeforeRLock", mname, p) }
		m.AfterRLock = func() { m.defaultCallback("AfterRLock", mname, p) }
		m.BeforeRUnlock = func() { m.defaultCallback("BeforeRUnlock", mname, p) }
		m.AfterRUnlock = func() { m.defaultCallback("AfterRUnlock", mname, p) }
		m.AfterRUnlockRecover = func(r interface{}) { m.defaultCallback1("AfterRUnlockRecover", mname, p, r) }
	}
	return m
}

// RLock calls the underlying RWMutex.RLock method. BeforeRLock and AfterRLock callbacks will be executed
// before and after such call respectively. If callback was not specified, it will be ignored.
func (m *RWMutex) RLock() {
	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.BeforeRLock != nil {
			m.BeforeRLock()
		}
	}()

	m.Mutex.mu.RLock()

	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.AfterRLock != nil {
			m.AfterRLock()
		}
	}()
}

// RUnlock calls the underlying RWMutex.RUnlock method. BeforeRUnlock and AfterRUnlock callbacks will be executed
// before and after such call respectively. If a panic will occur at underlying RWMutex unlocking, it will be
// handled by a call to recover() and BeforeRUnlockRecover and AfterRUnlockRecover will be called respectively.
// If callback was not specified, it will be ignored.
func (m *RWMutex) RUnlock() {
	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.BeforeRUnlock != nil {
			m.BeforeRUnlock()
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
			if m.AfterRUnlockRecover != nil {
				m.AfterRUnlockRecover(r)
			}
		}()
	}()
	m.Mutex.mu.RUnlock()

	func() {
		m.callbacksMu.Lock()
		defer m.callbacksMu.Unlock()
		if m.AfterRUnlock != nil {
			m.AfterRUnlock()
		}
	}()
}
