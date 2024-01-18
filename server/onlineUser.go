package server

import (
	"diceGame/config"
	"diceGame/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	lua "github.com/yuin/gopher-lua"
)

type OnLineUser struct {
	Addr   *string `json:"-"`
	Uid    int     `json:"uid"`
	ImgUrl string  `json:"img"`
	Name   string  `json:"name"`
	RoomId int     `json:"room_id"`
	InRoom bool    `json:"-"`
	// cache Conn for close
	WS *websocket.Conn `json:"-"`
}

// TODO some Mutex are unessential
var _OnLineLock sync.Mutex
var _OnLine = make(map[int]*OnLineUser)

var _OnLineUIDLock sync.Mutex
var _OnLineUID = make(map[string]int)

func SetOnlineUser(user *OnLineUser) {
	if user.Uid == 0 {
		utils.Logger.Error("SetOnlineUser user id 0")
		return
	}

	_OnLineLock.Lock()
	defer _OnLineLock.Unlock()
	if u, ok := _OnLine[user.Uid]; ok {
		utils.Logger.Warnf("exist user %d, update it", u.Uid)
		u.ImgUrl = user.ImgUrl
		u.Name = user.Name
		// u.RoomId = user.RoomId
		return
	}
	_OnLine[user.Uid] = user
}

func ClearUser(uid int) {
	utils.Logger.Debugf("uid:%d", uid)
	_OnLineLock.Lock()
	utils.Logger.Debugf("_OnLine:%v", _OnLine)
	if user, ok := _OnLine[uid]; ok {
		utils.Logger.Debugf("user:%v", *user)
		_OnLineUIDLock.Lock()
		utils.Logger.Debugf("_OnLineUID:%v", _OnLineUID)
		if _, ok := _OnLineUID[*user.Addr]; ok {
			delete(_OnLineUID, *user.Addr)
		}
		_OnLineUIDLock.Unlock()

		if user.WS != nil {
			if er := user.WS.Close(); er != nil {
				utils.Logger.Errorf(er.Error())
			}
		}
		_OnLine[uid] = nil
	}
	_OnLineLock.Unlock()
	return
}

func UserOffLine(remoteAddr string) {
	utils.Logger.Debugf("UserOffLine remoteAddr:%s", remoteAddr)
	uid := GetOnLineUID(remoteAddr)
	if uid == 0 {
		utils.Logger.Errorf("GetOnLineUID uid = 0! remoteAddr:%s", remoteAddr)
		return
	}
	_OnLineLock.Lock()
	if user, ok := _OnLine[uid]; ok {
		_OnLineUIDLock.Lock()
		if _, ok := _OnLineUID[*user.Addr]; ok {
			delete(_OnLineUID, *user.Addr)
		}
		_OnLineUIDLock.Unlock()

		if user.WS != nil {
			if er := user.WS.Close(); er != nil {
				utils.Logger.Errorf(er.Error())
			}
			user.WS = nil
		}
		user.Addr = nil
		user.InRoom = false
	}
	_OnLineLock.Unlock()
	return
}

func UserOnLine(uid int, conn *websocket.Conn) (*OnLineUser, error) {
	addr := conn.RemoteAddr().String()
	user := GetOnLineUser(uid)
	if user != nil {
		if user.Addr != nil && *user.Addr != addr {
			return nil, errors.New(fmt.Sprintf("user already online Addr:%s", *user.Addr))
		} else {
			user.Addr = &addr
			user.WS = conn
			CacheOnLineUID(user.Uid, conn)
			return user, nil
		}
	}

	if config.CFG.Lobby.Addr != "" {
		if dcUser := RequestDcUser(uid); dcUser != nil {
			dcUser.Addr = &addr
			dcUser.WS = conn
			CacheOnLineUID(dcUser.Uid, conn)
			_OnLineLock.Lock()
			_OnLine[uid] = dcUser
			_OnLineLock.Unlock()
			return dcUser, nil
		}
	}

	if config.CFG.Model == `debug` {
		utils.Logger.Debug("debug model create a debug user")
		newOnline := &OnLineUser{Uid: uid, Addr: &addr, WS: conn, ImgUrl: "local.png", Name: fmt.Sprint(uid)}
		CacheOnLineUID(uid, conn)
		_OnLineLock.Lock()
		_OnLine[uid] = newOnline
		_OnLineLock.Unlock()
		return newOnline, nil
	}

	return nil, errors.New(fmt.Sprintf("user not cache uid: %d", uid))
}

func RequestDcUser(uid int) *OnLineUser {
	addr := config.CFG.Lobby.DcAddr
	if addr == "" {
		utils.Logger.Error("RequestDcUser but dcaddr not find")
		return nil
	}
	secret := config.CFG.Lobby.DcSecret
	if secret == "" {
		utils.Logger.Warn(`RequestDcUser but secret not find use ""`)
		return nil
	}
	t := time.Now().Add(time.Hour)
	var timestamp int64 = t.Unix()
	token := utils.GenToken("/dc", fmt.Sprint(timestamp), secret)
	reqUrl := fmt.Sprintf("http://%s/dc/user/%d?sign=%s&time=%d", addr, uid, token, timestamp)
	utils.Logger.Debugf("reqUrl:%s", reqUrl)
	resp, err := http.Get(reqUrl)
	if err != nil {
		utils.Logger.Warn(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Logger.Warn(err)
		return nil
	}
	utils.Logger.Debugf("body:%s", body[:])
	user := OnLineUser{}
	er := json.Unmarshal(body, &user)
	if er != nil {
		utils.Logger.Error(er)
		return nil
	}
	if user.Uid != 0 {
		return &user
	}

	return nil
}

func GetOnLineUser(uid int) *OnLineUser {
	_OnLineLock.Lock()
	if user, ok := _OnLine[uid]; ok {
		_OnLineLock.Unlock()
		return user
	}
	_OnLineLock.Unlock()

	return nil
}

func CacheOnLineUID(uid int, conn *websocket.Conn) bool {
	_OnLineUIDLock.Lock()
	defer _OnLineUIDLock.Unlock()

	remote := conn.RemoteAddr().String()
	utils.Logger.Debugf(fmt.Sprintf("cache uid: %d, remote %s", uid, remote))
	if _, ok := _OnLineUID[remote]; ok {
		return true
	}
	_OnLineUID[remote] = uid
	return true
}

func GetOnLineUID(remoteAddr string) int {
	_OnLineUIDLock.Lock()
	defer _OnLineUIDLock.Unlock()

	if uid, ok := _OnLineUID[remoteAddr]; ok {
		return uid
	}
	return 0
}

type ResMsg struct {
	MsgType int             `json:"type"`
	Op      string          `json:"op"`
	Data    json.RawMessage `json:"data"`
}

func UserSend(L *lua.LState) int {
	utils.Logger.Debugf("UserSend")
	args := L.GetTop()
	if args != 3 {
		utils.Logger.Errorf("need 3 args")
		// L.Push(lua.LBool(false))
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

	data := L.ToString(3)

	utils.Logger.Debugf(fmt.Sprintf("data: %s", json.RawMessage(data)))

	respMsg := ResMsg{MsgType: GameType, Op: msg, Data: json.RawMessage(data)}

	res, _ := json.Marshal(respMsg)

	utils.Logger.Debugf(fmt.Sprintf("res: %s", res))
	err := conn.WriteMessage(websocket.TextMessage, res)
	if err != nil {
		utils.Logger.Errorf(err.Error())
		// L.Push(lua.LBool(false))
		return 0
	}
	// L.Push(lua.LBool(true))
	return 0
}

func UserOut(L *lua.LState) int {
	utils.Logger.Debugf("UserOut")
	uid := L.ToInt(1)
	L.Pop(1)

	ClearUser(uid)

	ret := L.GetGlobal("__THIS_ROOM_ID")
	L.Push(ret)
	roomId := L.ToInt(1)
	utils.Logger.Debugf("roomId %d", roomId)
	L.Pop(1)
	UserLeaveToLobby(roomId, uid)
	return 0
}

func closeVM(L *lua.LState) int {
	utils.Logger.Debugf("closeVM")
	ret := L.GetGlobal("__THIS_ROOM_ID")
	L.Push(ret)
	roomId := L.ToInt(1)
	utils.Logger.Debugf("roomId %d", roomId)
	L.Pop(1)
	CloseOneRunQue(roomId)
	return 0
}

func closeWS(L *lua.LState) int {
	utils.Logger.Debugf("closeWS")
	WS := L.ToUserData(1)
	conn := WS.Value.(*websocket.Conn)
	if conn == nil {
		utils.Logger.Errorf("WS nil")
		return 0
	}
	UserOffLine(conn.RemoteAddr().String())
	return 0
}
