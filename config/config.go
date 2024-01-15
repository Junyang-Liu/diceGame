package config

import (
	"diceGame/utils"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Congfig struct {
	Server struct {
		Addr      string `yaml:"addr"`
		LobbyAddr string `yaml:"lobby_addr"`
		GameID    int    `yaml:"game_id"`
		Priority  int    `yaml:"priority"`
		LuaStart  string `yaml:"lua_start"`
	} `yaml:"server"`

	Lobby struct {
		Addr     string `yaml:"addr"`
		LobbyId  int    `yaml:"lobby_id"`
		LuaStart string `yaml:"lua_start"`
	} `yaml:"lobby"`

	DC struct {
		Addr   string `yaml:"addr"`
		Secret string `yaml:"secret"`
	} `yaml:"dc"`

	Redis struct {
		Addr     string `yaml:"addr"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		PoolSize int    `yaml:"poolsize"`
	} `yaml:"redis"`

	Mysql struct {
		Addr     string `yaml:"addr"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
		DB       string `yaml:"db"`
		PoolSize int    `yaml:"poolsize"`
	} `yaml:"mysql"`

	Model string `yaml:"model"`

	ServerModel string `yaml:"server_model"`

	LogLevel string `yaml:"log"`

	MsgMaxMain int `yaml:"msg_max_main"`
}

var CFG Congfig

func InitCFG(paras ...string) {
	fileName := "go.yaml"
	if len(paras) > 0 && paras[0] != "" {
		fileName = paras[0]
	}
	fmt.Printf("fileName: %s, paras:%v\n", fileName, paras)
	cfgFileBuffer, err := os.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("config init faild err: %s", err.Error()))
	}

	CFG = Congfig{}
	err = yaml.Unmarshal(cfgFileBuffer, &CFG)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", CFG)
	if CFG.LogLevel != "" {
		utils.SetLogLevel(CFG.LogLevel)
	}
}
