package main

import (
	"diceGame/config"
	"diceGame/db"
	"diceGame/server"
	"diceGame/utils"
	"flag"
	_ "net/http/pprof"
)

func main() {

	var confFilePath string
	flag.StringVar(&confFilePath, "c", "./go.yaml", "path to yaml config")
	flag.Parse()
	config.InitCFG(confFilePath)

	utils.Logger.Info("starting ")
	err := db.InitRedis()
	if err != nil {
		utils.Logger.Errorf(err.Error())
		return
	}

	server.InitHttpServer()
	server.InitMsgServer()

	select {}
}
