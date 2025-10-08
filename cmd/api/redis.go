package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func (app *application) connectToRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     app.redisURL,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	log.Println("Connected to Redis!")
	return rdb, nil
}
