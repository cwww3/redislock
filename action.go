package redislock

import "context"

type Action interface {
	Do(ctx context.Context, lost <-chan struct{})
}

type OnlyOne func(ctx context.Context, lost <-chan struct{}) // 适合执行一次的程序，即时执行完成后，也一直续租持有锁,防止其他程序执行

func (one OnlyOne) Do(ctx context.Context, lost <-chan struct{}) {
	one(ctx, lost)
	select {
	case <-ctx.Done():
	case <-lost:
	}
}

type ActionFunc func(ctx context.Context, lost <-chan struct{})

func (action ActionFunc) Do(ctx context.Context, lost <-chan struct{}) {
	action(ctx, lost)
}
