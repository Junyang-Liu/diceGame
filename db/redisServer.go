package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"diceGame/config"
	"diceGame/utils"
)

var RedisConn *redis.Client

var ctx = context.Background()

func initRedisClient() error {
	fmt.Println("initRedisClient")

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.CFG.Redis.Addr,
		Password: config.CFG.Redis.Password,
		Username: config.CFG.Redis.UserName,
		DB:       config.CFG.Redis.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	RedisConn = rdb
	return nil
}

func GetGlobalRedisConn() *redis.Client {
	return RedisConn
}

func InitRedis() error {
	utils.Logger.Info("InitRedis()")

	if config.CFG.Redis.Addr == "" {
		utils.Logger.Warn("not find redis config")
		return nil
	}

	err := initRedisClient()
	if err != nil {
		return err
	}

	// err = RedisConn.Set(ctx, "key", "value", 0).Err()
	// if err != nil {
	// 	return err
	// }

	// val, err := RedisConn.Get(ctx, "key").Result()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("key", val)

	// val2, err := RedisConn.Get(ctx, "key2").Result()
	// if err == redis.Nil {
	// 	fmt.Println("key2 does not exist")
	// } else if err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("key2", val2)
	// }
	// Output: key value
	// key2 does not exist
	return nil
}
