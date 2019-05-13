package main

import (
"github.com/go-redis/redis"
"github.com/tsocial/vite"
"strings"
)

const (
	redisSentinelServers    = "0.0.0.0:26379|0.0.0.0:26380|0.0.0.0:26381"
	redisSentinelMasterName = "master"
	redisPassword           = "password"
	redisDB                 = 0
	separator               = "|"
)

var Redis *vite.Redis

func loadRedisConfig() {
	redisConfig := &vite.RedisConfig{
		UseSentinel:        true,
		SentinelMasterName: vite.EVString("REDIS_SENTINEL_MASTER_NAME", redisSentinelMasterName),
		SentinelServers:    strings.Split(vite.EVString("REDIS_SENTINEL_SERVERS", redisSentinelServers), separator),
		Password:           vite.EVString("REDIS_PASSWORD", redisPassword),
		DB:                 vite.EVInt("REDIS_DB", redisDB),
		PoolSize:           vite.EVInt("REDIS_POOL_SIZE", 10),
	}
	var err error
	Redis, err = vite.CreateRedisClient(redisConfig)
	if err != nil {
		panic(err)
	}
}

func readKey(key string) string {
	return Redis.Client.Get(key).String()
}

func writeKey(key string, value string) {
	err := Redis.Client.Set(key, value, 10000000000).Err()
	if err != nil {
		panic(err)
	}
}

func readKeyWithContext(client *redis.Client, key string) string {
	return client.Get(key).String()
}

func writeKeyWithContext(client *redis.Client, key string, value string) {
	err := client.Set(key, value, 10000000000).Err()
	if err != nil {
		panic(err)
	}
}

