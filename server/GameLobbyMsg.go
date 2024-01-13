package server

import (
	"diceGame/config"
	"diceGame/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	lua "github.com/yuin/gopher-lua"
)

var LobbyWS *websocket.Conn

func InitClientToLobby() {
	if config.CFG.Model == "debug" {
		time.Sleep(500 * time.Millisecond)
	}
	addr := config.CFG.Server.LobbyAddr
	u := url.URL{Scheme: "ws", Host: addr, Path: "/lobby"}
	log.Printf("connecting to lobby%s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Println("dial:", err)
		return
	}
	LobbyWS = c
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read lobby:", err)
				return
			}

			DoLobbyMsg(message)
		}
	}()

	msg := &LobbyMsg{FastLabel: EnumNewGameServer,
		Msg: []byte(fmt.Sprintf(`{"game_addr":"%s/dice"}`, config.CFG.Server.Addr))}
	SentToLobby(msg)

}

const (
	EnumNone = iota
	EnumNewGameServer
	EnumNewGameRoom
	EnumDestroyGameRoom
	EnumNewUser
	EnumUserLeave
	EnumGameRoomStarPlay
	EnumGameRoomEndPlay
)

type LobbyMsg struct {
	Msg       json.RawMessage `json:"msg"`
	RoomId    int             `json:"room_id"`
	GameID    int             `json:"game_id"`
	Priority  int             `json:"priority"`
	FastLabel int             `json:"fast_label"`
}

func DoLobbyMsg(message []byte) {
	utils.Logger.Debugf("DoLobbyMsg message: %s", message)
	msg := LobbyMsg{}
	if err := json.Unmarshal(message, &msg); err != nil {
		utils.Logger.Errorf("DoLobbyMsg err:%s", err.Error())
		return
	}

	if msg.FastLabel == EnumNewGameRoom {
		info := map[string]int{}
		if err := json.Unmarshal([]byte(msg.Msg), &info); err != nil {
			utils.Logger.Errorf("DoLobbyMsg err:%s", err.Error())
			return
		}
		uid, ok := info["uid"]
		if !ok {
			utils.Logger.Errorf("DoLobbyMsg NewGameRoom get none uid")
			return
		}

		if msg.RoomId == 0 {
			utils.Logger.Warn("DoLobbyMsg NewGameRoom get 0 RoomId, game server create a RoomId")
		}
		ret := NewGameVm(uid, msg.RoomId, false)
		retMsg := &LobbyMsg{RoomId: ret, FastLabel: EnumNewGameRoom, Msg: []byte(fmt.Sprintf(`{"uid":%d}`, uid))}

		SentToLobby(retMsg)
		return
	}

	if msg.FastLabel == EnumDestroyGameRoom {
		if msg.RoomId == 0 {
			utils.Logger.Warn("DoLobbyMsg DestroyGameRoom faild, RoomId 0")
			return
		}
		luaFucName := `Room:destroy`
		runMsg := RunMsg{isLobby: true, op: &luaFucName}
		CacheRunMsg(msg.RoomId, &runMsg)
		return
	}

	if msg.FastLabel == EnumNewUser {
		user := OnLineUser{}
		if err := json.Unmarshal([]byte(msg.Msg), &user); err != nil {
			utils.Logger.Errorf("DoLobbyMsg new user err:%s", err.Error())
			return
		} else {
			SetOnlineUser(&user)
			userByte, _ := json.Marshal(user)
			retMsg := &LobbyMsg{Msg: userByte, FastLabel: EnumNewUser}
			SentToLobby(retMsg)
		}
		return
	}

	utils.Logger.Warnf("lobby msg not handle message:%s", message)

}

func SentToLobby(msg *LobbyMsg) {
	if LobbyWS == nil {
		utils.Logger.Error("LobbyWS nil")
		return
	}
	msg.GameID = config.CFG.Server.GameID
	msg.Priority = config.CFG.Server.Priority

	utils.Logger.Debugf("SentToLobby msg:%v", *msg)
	if err := LobbyWS.WriteJSON(*msg); err != nil {
		utils.Logger.Error("SentToLobby  err:", err)
	}
}

func UserLeaveToLobby(roomId, uid int) {
	msg := &LobbyMsg{Msg: []byte(fmt.Sprintf(`{"uid":%d}`, uid)), RoomId: roomId, FastLabel: EnumUserLeave}
	SentToLobby(msg)
}

func GameStarPlayToLobby(L *lua.LState) int {
	utils.Logger.Debug("GameStarPlayToLobby")
	ret := L.GetGlobal("__THIS_ROOM_ID")
	L.Push(ret)
	roomId := L.ToInt(1)
	utils.Logger.Debugf("roomId %d", roomId)
	L.Pop(1)
	msg := &LobbyMsg{RoomId: roomId, FastLabel: EnumGameRoomStarPlay}
	SentToLobby(msg)
	return 0
}

func GameEndPlayToLobby(L *lua.LState) int {
	utils.Logger.Debug("GameEndPlayToLobby")
	result := L.ToString(1)
	if result == "" {
		utils.Logger.Warn("GameEndPlayToLobby result len 0")
	}
	L.Pop(1)

	ret := L.GetGlobal("__THIS_ROOM_ID")
	L.Push(ret)
	roomId := L.ToInt(1)
	utils.Logger.Debugf("roomId %d", roomId)
	L.Pop(1)
	msg := &LobbyMsg{Msg: []byte(result), RoomId: roomId, FastLabel: EnumGameRoomEndPlay}
	SentToLobby(msg)
	return 0
}
