package redislock

type Action interface {
	Do(stop, lost <-chan struct{})
}

type OnlyOne func(stop, lost <-chan struct{}) // 适合执行一次的程序，即时执行完成后，也一直续租持有锁,防止其他程序执行

func (one OnlyOne) Do(stop, lost <-chan struct{}) {
	one(stop, lost)
	select {
	case <-stop:
	case <-lost:
	}
}

type ActionFunc func(stop, lost <-chan struct{})

func (action ActionFunc) Do(stop, lost <-chan struct{}) {
	action(stop, lost)
}
