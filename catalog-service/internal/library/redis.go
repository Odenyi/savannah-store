package library

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func GetRedisKey(conn *redis.Client, key string) (string, error) {

	var data string
	data, err := conn.Get(key).Result()
	if err != nil {

		return data, fmt.Errorf("error getting key %s: %v", key, err)
	}

	return data, err

}

func SetRedisKey(conn *redis.Client, key string, value string) error {

	_, err := conn.Set(key, value, time.Second*time.Duration(0)).Result()
	if err != nil {

		v := string(value)

		if len(v) > 15 {

			v = v[0:12] + "..."
		}

		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}
	return err
}

func SetRedisKeyWithExpiry(conn *redis.Client, key string, value string, seconds int) error {

	_, err := conn.Set(key, value, time.Second*time.Duration(seconds)).Result()
	if err != nil {

		v := string(value)

		if len(v) > 15 {

			v = v[0:12] + "..."
		}

		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}

	return err
}

func DeleteRedisKey(conn *redis.Client, key string) error {

	_, err := conn.Del(key).Result()

	if err != nil {

		return fmt.Errorf("error getting key %s: %v", key, err)
	}

	return err
}

func GetAllKeys(redisConn *redis.Client, key string) (error, map[string]string) {

	var data []string

	results := map[string]string{}

	data, err := redisConn.Keys(key).Result()
	if err != nil {

		return err, nil
	}

	for _, k := range data {

		dt, err := redisConn.Get(k).Result()
		if err != nil {

			continue
		}

		results[k] = dt
	}

	return nil, results
}

func IncRedisKey(conn *redis.Client, key string) (int64, error) {

	var data int64
	data, err := conn.Incr(key).Result()

	if err != nil {

		return data, fmt.Errorf("error getting key %s: %v", key, err)
	}

	return data, err
}
