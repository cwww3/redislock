package redislock

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
)

// From https://github.com/moira-alert/moira/blob/master/database/redis/locks.go

type Worker struct {
	ttl    time.Duration
	name   string
	sync   *redsync.Redsync
	action Action
}

func NewWork(rs *redsync.Redsync, name string, ttl time.Duration, action Action) *Worker {
	return &Worker{
		name:   name,
		ttl:    ttl,
		sync:   rs,
		action: action,
	}
}

func (w *Worker) Run(ctx context.Context) {
	m := w.newLock(w.name, w.ttl)
	lost, err := m.Acquire(ctx)
	if err != nil {
		return
	}

	defer m.Release()
	w.action.Do(ctx, lost)
}

// LoopRun 相较与Run,LoopRun能在因为续租失败后重新获取锁,然后重新执行Do方法
func (w *Worker) LoopRun(ctx context.Context) {
	for {
		func() {
			m := w.newLock(w.name, w.ttl)
			lost, err := m.Acquire(ctx)
			if err != nil {
				return
			}

			defer m.Release()
			w.action.Do(ctx, lost)
		}()
		time.Sleep(w.ttl / 3)
	}
}

// newLock returns the Lock which can be used to Acquire or Release the lock
func (w *Worker) newLock(name string, ttl time.Duration) *Lock {
	mutex := w.sync.NewMutex(name, redsync.WithExpiry(ttl), redsync.WithTries(1))
	return &Lock{name: name, ttl: ttl, mutex: mutex}
}
