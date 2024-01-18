package main

import (
	"context"
	"diceGame/config"
	"diceGame/db"
)

type User struct {
	Name  string `redis:"name,omitempty"`
	Id    int    `redis:"id"`
	Photo string `redis:"photo,omitempty"`
}

var ctx = context.Background()

func main2() {
	config.InitCFG("../go.yaml")
	db.InitRedis()
	conn := db.GetGlobalRedisConn()

	user1 := User{Name: "u1", Id: 100001}
	err := conn.HSet(ctx, "user:100001", user1).Err()
	if err != nil {
		panic(err)
	}
	user2 := User{Name: "u2", Id: 100002}
	err = conn.HSet(ctx, "user:100002", user2).Err()
	if err != nil {
		panic(err)
	}
	user3 := User{Name: "u3", Id: 100003}
	err = conn.HSet(ctx, "user:100003", user3).Err()
	if err != nil {
		panic(err)
	}
}
