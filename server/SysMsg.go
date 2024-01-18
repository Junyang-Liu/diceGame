package server

import (
	"diceGame/config"
	"diceGame/utils"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

const ()

type FuncType func(user *OnLineUser, op string, data any, gameOp string)

var SysFuncMap = map[string]FuncType{}

func init() {
	SysFuncMap["newGame"] = newGame
	SysFuncMap["opGame"] = playerOpGame
}

func DoSysCommand(user *OnLineUser, op string, data any, gameOp string) {
	if f, ok := SysFuncMap[op]; ok {
		f(user, op, data, gameOp)
		return
	}

	conn := user.WS
	if conn != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("not support op:%s", op)))
	} else {
		utils.Logger.Errorf(fmt.Sprintf("not found conn uid: %d", user.Uid))
	}
}

func newGame(user *OnLineUser, op string, data any, gameOp string) {
	utils.Logger.Debug("newGame")
	if config.CFG.Lobby.Addr != "" && config.CFG.Model != "debug" {
		utils.Logger.Error("newGame faild, lobby server and not debug model")
		return
	}

	if user.RoomId != 0 {
		if user.WS != nil {
			if err := user.WS.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("already in room: %d", user.RoomId))); err != nil {
				utils.Logger.Errorf(err.Error())
			}
		} else {
			utils.Logger.Errorf(fmt.Sprintf("not found conn uid: %d", user.Uid))
		}
		return
	}

	roomID := NewGameVm(user.Uid, 0, false)

	utils.Logger.Warnf("newGame by a user, roomID:%d check if you are debugging!", roomID)
	retData := map[string]int{"roomId": roomID, "code": 0}
	respMsg := Msg{MsgType: SysType, Op: op + "Rsp", Data: retData}
	res, _ := json.Marshal(respMsg)
	if user.WS != nil {
		if err := user.WS.WriteMessage(websocket.TextMessage, res); err != nil {
			utils.Logger.Errorf(err.Error())
		}
	} else {
		utils.Logger.Errorf(fmt.Sprintf("not found conn uid: %d", user.Uid))
	}

}

func playerOpGame(user *OnLineUser, op string, data any, gameOp string) {
	PlayerOpGame(user, gameOp, data)
}
