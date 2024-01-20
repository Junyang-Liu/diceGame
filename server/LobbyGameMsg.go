package server

import (
	"diceGame/config"
	"diceGame/utils"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	lua "github.com/yuin/gopher-lua"
)

func InitLobbyRoom() {
	lobbyId := config.CFG.Lobby.LobbyId
	NewGameVm(lobbyId, lobbyId, true)
}

const gameWsWait = 20 * time.Second

func RecvGameServerMsg(conn *websocket.Conn) {
	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			utils.Logger.Errorf(fmt.Sprintf("game server offline %d-%s-%s-%s", msgType, conn.RemoteAddr(), message, err.Error()))
			return
		}
		SetWsPingHandler(conn, gameWsWait)
		DoGameServerMsg(conn, conn.RemoteAddr().String(), message)

	}
}

func DoGameServerMsg(conn *websocket.Conn, addr string, message []byte) {
	lobbyId := config.CFG.Lobby.LobbyId
	utils.Logger.Debugf("DoGameServerMsg message:%s lobbyId:%d", message, lobbyId)

	if lobbyId != 0 {
		op := "GameServerCall"
		var callData any
		data := make(map[string]interface{})
		data["conn"] = conn
		data["addr"] = addr
		data["message"] = string(message)

		callData = data
		runMsg := &RunMsg{isLobby: true, op: &op, data: &callData}

		if err := CacheRunMsg(lobbyId, runMsg); err != nil {
			utils.Logger.Error(err.Error())
		}
	}
}

func GameServerSent(L *lua.LState) int {

	utils.Logger.Debugf("GameServerSent")
	args := L.GetTop()
	if args != 2 {
		utils.Logger.Errorf("need 2 args")
		return 0
	}
	WS := L.ToUserData(1)
	conn := WS.Value.(*websocket.Conn)
	if conn == nil {
		utils.Logger.Errorf("WS nil")
		return 0
	}

	msg := L.ToString(2)
	utils.Logger.Debugf(fmt.Sprintf("msg: %s", msg))

	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		utils.Logger.Errorf(err.Error())
		return 0
	}
	return 0
}
