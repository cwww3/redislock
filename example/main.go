package main

import (
	"fmt"
	"time"

	"github.com/cwww3/redislock"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	redislib "github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

func main() {
	c := redislib.NewClient(&redislib.Options{
		Addr: ":6379",
	})
	pool := goredis.NewPool(c)
	rs := redsync.New(pool)

	w := redislock.NewWork(rs, "lock", time.Minute, redislock.OnlyOne(func(ctx context.Context, lost <-chan struct{}) {
		fmt.Println("执行")
	}))

	ctx, cancel := context.WithCancel(context.Background())

	time.AfterFunc(time.Second*25, func() {
		cancel()
	})
	w.Run(ctx)
}
