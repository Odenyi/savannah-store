package repository

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	// "strconv"
	"time"
)

func RedisClient() *redis.Client {

	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	// db := os.Getenv("redis_database_number")
	auth := os.Getenv("REDIS_AUTH")

	uri := fmt.Sprintf("redis://%s:%s", host, port)
	uri = fmt.Sprintf("%s:%s", host, port)

	opts := redis.Options{
		MinIdleConns: 10,
		IdleTimeout:  60 * time.Second,
		PoolSize:     1000,
		Addr:         uri,
		// DB:           dbNumber, // use default DB
	}

	if len(auth) > 0 {

		opts.Password = auth
	}

	client := redis.NewClient(&opts)

	return client
}
