package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func (app *application) connectToRedis() (*redis.Client, error) {
	opt, err := redis.ParseURL(app.redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	log.Println("Connected to Redis!")
	return rdb, nil
}
