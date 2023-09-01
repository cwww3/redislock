package redislock

import (
	"time"

	"github.com/go-redsync/redsync/v4"
)

// Lock is used to hide low-level details of redsync.Mutex such as an extension of it
type Lock struct {
	name   string
	ttl    time.Duration
	mutex  *redsync.Mutex
	extend chan struct{}
}

// Acquire attempts to acquire the lock and blocks while doing so
// Providing a non-nil stop channel can be used to abort the acquire attempt
// Returns lost channel that is closed if the lock is lost or an error
func (lock *Lock) Acquire(stop <-chan struct{}) (<-chan struct{}, error) {
	for {
		lost, err := lock.tryAcquire()
		if err == nil {
			return lost, nil
		}

		select {
		case <-stop:
			{
				return nil, nil
			}
		case <-time.After(lock.ttl / 3): //nolint
			{
				continue
			}
		}
	}
}

// Release releases the lock
func (lock *Lock) Release() {
	close(lock.extend)
	lock.mutex.Unlock() //nolint
}

func (lock *Lock) tryAcquire() (<-chan struct{}, error) {
	if err := lock.mutex.Lock(); err != nil {
		return nil, err
	}

	lost := make(chan struct{})
	lock.extend = make(chan struct{})
	go extendMutex(lock.mutex, lock.ttl, lost, lock.extend)
	return lost, nil
}

func extendMutex(mutex *redsync.Mutex, ttl time.Duration, done chan struct{}, stop <-chan struct{}) {
	defer close(done)
	extendTicker := time.NewTicker(ttl / 3) //nolint
	defer extendTicker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-extendTicker.C:
			result, _ := mutex.Extend()
			if !result {
				return
			}
		}
	}
}
