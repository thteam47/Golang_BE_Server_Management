package drive

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	MyRediscache *cache.Cache
}

var Redis = &RedisCache{}

func ConnectRedis(dbredis string) *RedisCache {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": dbredis,
			//"server2": ":6380",
		},
	})

	err := ring.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("Error connect the redis: %s", err)
	}
	memcache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	Redis.MyRediscache = memcache
	return Redis
}
