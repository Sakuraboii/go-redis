package cache

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

var REDIS *redis.Client //

func ConnectRedisCache() {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, redis_err := redisClient.Ping().Result(); redis_err != nil {
		fmt.Println(redis_err.Error())
		panic("Error: Unable to connect to Redis")

	}

	REDIS = redisClient
	fmt.Println("Connected to Redis cache successfully")
}

func SetInCache(c *redis.Client, key int64, value interface{}) bool {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Unable to marshal element to JSON")
		return false
	}

	_, err = c.Set(strconv.FormatInt(key, 10), marshalledValue, 1*time.Hour).Result()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Unable to set element in cache")
		return false

	}

	return true
}

func GetFromCache(c *redis.Client, key int64) interface{} {
	value, err := c.Get(strconv.FormatInt(key, 10)).Result()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Unable to get element from cache")
		return nil
	}

	return value
}

func DeleteFromCache(c *redis.Client, key int64) {
	_, err := c.Del(strconv.FormatInt(key, 10)).Result()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error: Unable to delete element from cache")
	}

	return
}
