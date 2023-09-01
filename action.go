package redislock

import "context"

type Action interface {
	Do(ctx context.Context, lost <-chan struct{})
}

type OnlyOne func(ctx context.Context) // 适合执行一次的程序，即时执行完成后，也一直续租持有锁,防止其他程序执行

func (one OnlyOne) Do(ctx context.Context, lost <-chan struct{}) {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		// 将lost和上层ctx整合成subCtx传给函数
		select {
		case <-ctx.Done():
		case <-lost:
		}
		cancel()
	}()

	one(subCtx)

	// 阻塞
	select {
	case <-subCtx.Done():
	}
}

type ActionFunc func(ctx context.Context)

func (action ActionFunc) Do(ctx context.Context, lost <-chan struct{}) {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		// 将lost和上层ctx整合成subCtx传给函数
		select {
		case <-ctx.Done():
		case <-lost:
		}
		cancel()
	}()

	action(subCtx)
}
