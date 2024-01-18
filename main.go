package main

import (
	"diceGame/config"
	"diceGame/db"
	dcserver "diceGame/dcserver"
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
	err = db.InitMysql()
	if err != nil {
		utils.Logger.Errorf(err.Error())
		return
	}

	if config.CFG.Lobby.Addr != "" {
		server.InitLobbyRoom()
	}

	if config.CFG.DC.Addr != "" {
		dcserver.InitServer()
	}

	server.InitHttpServer()
	server.InitMsgServer()

	select {}
}
