package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	lua "github.com/yuin/gopher-lua"

	"diceGame/config"
	"diceGame/utils"
)

var RedisConn *redis.Client

var ctx = context.Background()

func initRedisClient() error {
	fmt.Println("initRedisClient")

	opt := redis.Options{
		Addr:     config.CFG.Redis.Addr,
		Password: config.CFG.Redis.Password,
		Username: config.CFG.Redis.UserName,
		DB:       config.CFG.Redis.DB,
	}
	if config.CFG.Redis.PoolSize != 0 {
		opt.PoolSize = config.CFG.Redis.PoolSize
	}

	rdb := redis.NewClient(&opt)

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

func RBDget(L *lua.LState) int {
	utils.Logger.Debug("RBDget")
	key := L.ToString(1)
	utils.Logger.Debugf("key:%s", key)

	runCtx, cancel := context.WithTimeout(ctx, 2000*time.Millisecond)
	defer cancel()

	_, err := RedisConn.Ping(ctx).Result()
	if err != nil {
		utils.Logger.Warn(err)
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	result, err := RedisConn.Get(runCtx, key).Result()
	if err != nil {
		if err == redis.Nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("redis: nil"))
		} else {
			utils.Logger.Warn(err)
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
		}
	} else {
		L.Push(lua.LString(result))
		L.Push(lua.LNil)
	}
	return 2
}

func RBDset(L *lua.LState) int {
	utils.Logger.Debug("RBDset")
	key := L.ToString(1)
	val := L.ToString(2)
	expire := L.ToInt(3)
	utils.Logger.Debugf("key:%s, val:%s, expire:%d", key, val, expire)

	runCtx, cancel := context.WithTimeout(ctx, 2000*time.Millisecond)
	defer cancel()

	_, err := RedisConn.Ping(ctx).Result()
	if err != nil {
		utils.Logger.Warn(err)
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	result, err := RedisConn.Set(runCtx, key, val, time.Duration(expire)*time.Second).Result()
	if err != nil {
		if err == redis.Nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("redis: nil"))
		} else {
			utils.Logger.Warn(err)
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
		}
	} else {
		L.Push(lua.LString(result))
		L.Push(lua.LNil)
	}
	return 2
}
