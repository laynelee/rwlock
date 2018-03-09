package rwlock

import "sync/atomic"

type Lock struct {
	chR        chan struct{}
	chW        chan struct{}
	rlockCount int32
}

func New() *Lock {
	ret := new(Lock)
	ret.chR = make(chan struct{}, 1)
	ret.chW = make(chan struct{}, 1)
	atomic.StoreInt32(&ret.rlockCount, 0)
	return ret
}

func (l *Lock) RLock() {
	l.chW <- struct{}{}

	if atomic.AddInt32(&l.rlockCount, 1) == 1 {
		l.chR <- struct{}{}
	}

	<-l.chW
}

func (l *Lock) RUnlock() {
	if atomic.AddInt32(&l.rlockCount, -1) == 0 {
		<-l.chR
	}
}

func (l *Lock) TryLock() bool {
	select {
	case l.chW <- struct{}{}:
	default:
		return false
	}

	select {
	case l.chR <- struct{}{}:
		return true
	default:
		<-l.chW
		return false
	}
}

func (l *Lock) TryRLock() bool {
	select {
	case l.chW <- struct{}{}:
	default:
		return false
	}

	if atomic.AddInt32(&l.rlockCount, 1) == 1 {
		l.chR <- struct{}{}
	}

	<-l.chW
	return true
}

func (l *Lock) Lock() {
	l.chW <- struct{}{}
	l.chR <- struct{}{}
}

func (l *Lock) Unlock() {
	<-l.chR
	<-l.chW
}
