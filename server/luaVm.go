package server

import (
	"diceGame/config"
	"diceGame/utils"
	"errors"
	"fmt"
	"sync"
	"time"

	luascript "diceGame/luascript"

	"github.com/gorilla/websocket"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"
)

type LuaVmFuncType func(user *OnLineUser, op string, data any)

var LuaVmFuncMap = map[string]LuaVmFuncType{}

func init() {
	LuaVmFuncMap["opGame"] = opGame
	LuaVmFuncMap["inGame"] = inGame
}

const (
	GO_STACK_NAME    = "go"
	LOBBY_STACK_NAME = "lobby"
)

const StartRoomId = 1000

var (
	GLuaVm     = map[int]*lua.LState{}
	VMlock     sync.Mutex
	nextRoomID = StartRoomId
	lock       sync.Mutex
)

func getNewRoomId() int {
	newRoomID := 0
	lock.Lock()
	nextRoomID++
	newRoomID = nextRoomID
	lock.Unlock()

	return newRoomID
}

func setLuaVmGoStack(L *lua.LState, isLobbyServer bool) {
	goStack := L.NewTypeMetatable(GO_STACK_NAME)
	L.SetField(goStack, "send", L.NewFunction(UserSend))
	L.SetField(goStack, "userOut", L.NewFunction(UserOut))
	L.SetField(goStack, "NewTimer", L.NewFunction(NewTimer))
	L.SetField(goStack, "CancelTimer", L.NewFunction(CancelTimer))
	L.SetField(goStack, "TimerLastTime", L.NewFunction(TimerLastTime))
	L.SetField(goStack, "closeVM", L.NewFunction(closeVM))

	if isLobbyServer {
		L.SetField(goStack, "GameServerSent", L.NewFunction(GameServerSent))
	} else {
		lobbyStack := L.NewTypeMetatable(LOBBY_STACK_NAME)
		L.SetField(lobbyStack, "StartPlay", L.NewFunction(GameStarPlayToLobby))
		L.SetField(lobbyStack, "EndPlay", L.NewFunction(GameEndPlayToLobby))
		L.SetGlobal(LOBBY_STACK_NAME, lobbyStack)
	}

	L.SetGlobal(GO_STACK_NAME, goStack)
}

func setLuaVmGameGlobal(L *lua.LState) {

}

func doGameScriptSuccess(L *lua.LState, luaStartPath string) bool {
	if luaStartPath == "" {
		luaStartPath = config.CFG.Server.LuaStart
	} else {
		utils.Logger.Warnf("luaStartPath:%s", luaStartPath)
	}
	if err := L.DoFile(luaStartPath); err != nil {
		utils.Logger.Errorf(err.Error())
		return false
	}
	return true
}

func CallLua(L *lua.LState, uid int, luaFunc, str1 string, data *lua.LTable) error {
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal(luaFunc),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(uid), lua.LString(str1), data); err != nil {
		utils.Logger.Errorf(err.Error())
		return err
	}
	L.Get(-1) // returned value
	L.Pop(1)  // remove received value
	return nil
}

func LobbyCallLua(L *lua.LState, fName string, data *lua.LTable) error {
	utils.Logger.Debugf("LobbyCallLua!!! fName: %s data: %v", fName, *data)

	room := L.GetGlobal("Room")
	f := L.GetField(room, fName)
	if err := L.CallByParam(lua.P{
		Fn:      f,
		NRet:    1,
		Protect: true,
	}, room, data); err != nil {
		utils.Logger.Errorf(err.Error())
		return err
	}
	// L.Get(-1) // returned value
	// L.Pop(1)  // remove received value
	return nil
}

func TimerCallLua(L *lua.LState, luaFunc *lua.LFunction, params []lua.LValue) error {
	utils.Logger.Debugf("TimerCallLua!!! luaFunc%s", luaFunc.String())
	if err := L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1,
		Protect: true,
	}, params...); err != nil {
		utils.Logger.Errorf(err.Error())
		return err
	}
	// L.Get(-1) // returned value
	// L.Pop(1)  // remove received value
	return nil
}

func NewGameVm(uid, roomID int, isLobbyServer bool) int {
	if roomID != 0 && GetGameVm(roomID) != nil {
		utils.Logger.Errorf("exist roomID: %d", roomID)
		return 0
	}
	L := lua.NewState()
	luajson.Preload(L)
	setLuaVmGoStack(L, isLobbyServer)
	setLuaVmGameGlobal(L)

	if err := L.DoString(luascript.Player); err != nil {
		utils.Logger.Errorf(err.Error())
		return 0
	}

	if err := L.DoString(luascript.Room); err != nil {
		utils.Logger.Errorf(err.Error())
		return 0
	}

	if err := L.DoString(luascript.Dice); err != nil {
		utils.Logger.Errorf(err.Error())
		return 0
	}

	var roomId = 0
	if roomID != 0 {
		roomId = roomID
	} else {
		roomId = getNewRoomId()
	}

	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("newRoom"),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(roomId), lua.LNumber(uid)); err != nil {
		utils.Logger.Errorf(err.Error())
		return 0
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)         // remove received value
	if lua.LVAsNumber(ret) == 0 {
		L.SetGlobal("__THIS_ROOM_ID", lua.LNumber(roomId))
		VMlock.Lock()
		GLuaVm[roomId] = L
		VMlock.Unlock()
		CacheNewRunQue(roomId, L)

		luaStartPath := config.CFG.Server.LuaStart
		if isLobbyServer {
			luaStartPath = config.CFG.Lobby.LuaStart
		}
		if doGameScriptSuccess(L, luaStartPath) {
			return roomId
		}
	}
	utils.Logger.Errorf("call lua newRoom error")
	L.Close()
	return 0
}

func GetGameVm(roomID int) *lua.LState {
	VMlock.Lock()
	defer VMlock.Unlock()
	if L, ok := GLuaVm[roomID]; ok {
		return L
	}
	return nil
}

func ReMoveGameVm(roomID int) {
	VMlock.Lock()
	defer VMlock.Unlock()
	if _, ok := GLuaVm[roomID]; ok {
		GLuaVm[roomID] = nil
	}
}

func UserGetGameVm(uid int, roomId int) (*lua.LState, error) {
	onLineUser := GetOnLineUser(uid)
	if onLineUser == nil {
		return nil, errors.New("onLineUser nil")
	}

	if onLineUser.RoomId != 0 && onLineUser.RoomId != roomId {
		conn := onLineUser.WS
		if conn == nil {
			return nil, errors.New("conn nil")
		}
		err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"data":{"msg":in room %d}}`, onLineUser.RoomId)))
		if err != nil {
			return nil, err
		}
	}
	if onLineUser.RoomId == 0 {
		onLineUser.RoomId = roomId
	}
	L := GetGameVm(onLineUser.RoomId)
	if L == nil {
		return nil, errors.New("GetGameVm nil")
	}
	return L, nil
}

func PlayerOpGame(user *OnLineUser, op string, data any) {
	if f, ok := LuaVmFuncMap[op]; ok {
		f(user, op, data)
		return
	}
	opGame(user, op, data)
}

func inGame(user *OnLineUser, op string, data any) {
	roomId := int(data.(float64))
	if roomId == 0 {
		utils.Logger.Error("roomId = 0")
		return
	}
	L, err := UserGetGameVm(user.Uid, roomId)
	if err != nil {
		utils.Logger.Error(err.Error())
		conn := user.WS
		if conn == nil {
			utils.Logger.Error("conn nil")
			return
		}
		err := conn.WriteMessage(websocket.TextMessage,
			[]byte(fmt.Sprintf(`{"type":1, "op":"%s", "data":{"msg":"not in any room, uid: %d, must inGame before opGame, faild, data: %v", "code":-1}}`, op, user.Uid, data)))
		if err != nil {
			utils.Logger.Errorf(err.Error())
			return
		}
		return
	}

	callData := L.NewTable()
	callData.RawSet(lua.LString("roomId"), lua.LNumber(roomId))
	callData.RawSet(lua.LString("uId"), lua.LNumber(user.Uid))
	WS := L.NewUserData()
	WS.Value = user.WS
	callData.RawSet(lua.LString("__WS"), WS)
	if err := CallLua(L, user.Uid, "userIn", op, callData); err != nil {
		utils.Logger.Errorf("inGame faild uid: %d", user.Uid)
	} else {
		user.RoomId = roomId
		user.InRoom = true
	}
	utils.Logger.Debugf("user: %v", user)
}

func opGame(user *OnLineUser, op string, data any) {

	if user.RoomId == 0 || user.InRoom == false {
		conn := user.WS
		if conn == nil {
			utils.Logger.Error("conn nil")
			return
		}
		err := conn.WriteMessage(websocket.TextMessage,
			[]byte(fmt.Sprintf(`{"type":2, "op":"%s", "data":{"msg":"not in any room, uid: %d, must inGame before opGame, faild, data: %v", "code":-1}}`, op, user.Uid, data)))
		if err != nil {
			utils.Logger.Errorf(err.Error())
			return
		}
	}

	newRunMsg := &RunMsg{uid: &user.Uid, op: &op, data: &data}
	err := CacheRunMsg(user.RoomId, newRunMsg)
	if err != nil {
		utils.Logger.Errorf(err.Error())
		return
	}
}

func NewTimer(L *lua.LState) int {
	utils.Logger.Debugf("NewTimer Run!!")
	args := L.GetTop()
	if args < 2 {
		utils.Logger.Error("NewTimer faild args not enough")
		// L.Push(lua.LBool(false))
		return 0
	}

	mills := L.ToInt(1)
	utils.Logger.Debugf(fmt.Sprintf("mills: %d", mills))

	lf := L.ToFunction(2)
	utils.Logger.Debugf(fmt.Sprintf("lf: %v", *lf))

	utils.Logger.Debugf(fmt.Sprintf("args: %d", args))

	params := []lua.LValue{}
	for i := 2; i < args; i++ {
		param := L.CheckAny(i + 1)
		params = append(params, param)
	}
	err := SetOneTimer(L, mills, lf, params)
	if err != nil {
		utils.Logger.Error(err.Error())
		return 0
	}

	// SetOneTimer push one userdata, return 1 here
	return 1
}

func CancelTimer(L *lua.LState) int {
	utils.Logger.Debugf("CancelTimer Run!!")
	args := L.GetTop()
	if args < 1 {
		utils.Logger.Error("CancelTimer faild args not enough")
		// L.Push(lua.LBool(false))
		return 0
	}

	userdata := L.ToUserData(1)
	utils.Logger.Debugf(fmt.Sprintf("userdata: %v", userdata))
	timer := userdata.Value.(*TimerInfo)
	timer.cancel = true

	utils.Logger.Debugf(fmt.Sprintf("timer: %v", timer))
	return 0
}

func TimerLastTime(L *lua.LState) int {
	utils.Logger.Debugf("TimerLastTime Run!!")
	args := L.GetTop()
	if args < 1 {
		utils.Logger.Error("TimerLastTime faild args not enough")
		// L.Push(lua.LBool(false))
		return 0
	}

	userdata := L.ToUserData(1)
	utils.Logger.Debugf(fmt.Sprintf("userdata: %v", userdata))
	timer := userdata.Value.(*TimerInfo)
	// timer.cancel = true

	utils.Logger.Debugf(fmt.Sprintf("timer: %v", timer))
	now := time.Now()
	last := timer.at.Sub(now).Milliseconds()
	if last < 0 {
		last = 0
	}
	L.Push(lua.LNumber(last))
	return 1
}
