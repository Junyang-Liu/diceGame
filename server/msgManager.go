package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"

	"diceGame/config"
	"diceGame/utils"
)

var msgQue chan *MsgCell

func init() {
}

type Msg struct {
	MsgType int         `json:"type"`
	Op      string      `json:"op"`
	Data    interface{} `json:"data"`
}

type MsgCell struct {
	Msg  *Msg
	User *OnLineUser
}

func (msgCell *MsgCell) String() string {
	Addr := "nil"
	if msgCell.User.Addr != nil {
		Addr = *msgCell.User.Addr
	}
	return fmt.Sprintf(
		"{Msg:{MsgType: %d, Op: %s, Data: %s} User:{Addr: %s, Uid: %d, RoomId: %d}}",
		msgCell.Msg.MsgType, msgCell.Msg.Op, msgCell.Msg.Data,
		Addr, msgCell.User.Uid, msgCell.User.RoomId,
	)
}

const PING = "PING"

func RecvMsg(conn *websocket.Conn) {
	SetWsCloseHandler(conn)
	SetWsPingHandler(conn, wsWait)

	var user *OnLineUser
	uid := GetOnLineUID(conn.RemoteAddr().String())
	if uid != 0 {
		user = GetOnLineUser(uid)
	}

	for {
		msgType, raw, err := conn.ReadMessage()
		if err != nil {
			UserOffLine(conn.RemoteAddr().String())
			utils.Logger.Errorf(fmt.Sprintf("掉线了 %d-%s-%s-%s", msgType, conn.RemoteAddr(), raw, err.Error()))
			return
		}

		utils.Logger.Debug(fmt.Sprintf("msgType:%d RemoteAddr:%s raw:%s", msgType, conn.RemoteAddr(), raw))
		if msgType != websocket.TextMessage {
			utils.Logger.Errorf("unsuport msgType:%d, RemoteAddr:%s", msgType, conn.RemoteAddr())
			UserOffLine(conn.RemoteAddr().String())
			return
		}

		// https://stackoverflow.com/questions/10585355/sending-websocket-ping-pong-frame-from-browser
		// it looks like js or browser can't sending a ping/pong frame
		if fmt.Sprintf("%s", raw[:]) == PING {
			TextPingHandler(conn, wsWait)
			continue
		}

		newMsg := Msg{}

		json.Unmarshal(raw, &newMsg)

		utils.Logger.Debug(fmt.Sprintf("newMsg: %#v", newMsg))
		if user != nil {
			MsgCache(user, &newMsg)
		} else {
			user = DoLogin(conn, &newMsg)
		}
	}
}

func DoLogin(conn *websocket.Conn, msg *Msg) *OnLineUser {
	if msg != nil && msg.Op == "login" && msg.Data != nil {
		uid := int(msg.Data.(float64))
		ret, err := UserOnLine(uid, conn)
		if err != nil {
			utils.Logger.Error(err.Error())
			err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":1, "op":"login", "data":{"msg":"login err: %s"}}`, err.Error())))
			if err != nil {
				utils.Logger.Error(err.Error())
			}
			return nil
		} else {
			retStr, _ := json.Marshal(ret)
			err := ret.WsWriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":1, "op":"login", "data":{"msg":"login success", "code":0, "user":%s}}`, []byte(retStr))))
			if err != nil {
				utils.Logger.Error(err.Error())
				return nil
			}
			return ret
		}
	}

	err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("haven't login")))
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	return nil
}

func SetWsCloseHandler(conn *websocket.Conn) {
	conn.SetCloseHandler(func(code int, text string) error {
		utils.Logger.Info(fmt.Sprintf("下线了 %d %s", code, text))
		UserOffLine(conn.RemoteAddr().String())
		return nil
	})
}

const wsWait = 10 * time.Second

func SetWsPingHandler(conn *websocket.Conn, wait time.Duration) {
	conn.SetReadDeadline(time.Now().Add(wait))
	conn.SetWriteDeadline(time.Now().Add(wait))

	conn.SetPingHandler(func(appData string) error {
		utils.Logger.Debugf("conn ping RemoteAddr:%s", conn.RemoteAddr())
		conn.SetReadDeadline(time.Now().Add(wait))
		conn.SetWriteDeadline(time.Now().Add(wait))
		if user := GetOnLineUserByAddr(conn.RemoteAddr().String()); user != nil {
			if err := user.WsWriteMessage(websocket.PongMessage, []byte{}); err != nil {
				utils.Logger.Error(err)
			}
		} else if err := conn.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
			utils.Logger.Error(err)
		}
		return nil
	})
}

func TextPingHandler(conn *websocket.Conn, wait time.Duration) {
	utils.Logger.Debugf("conn ping RemoteAddr:%s", conn.RemoteAddr())
	conn.SetReadDeadline(time.Now().Add(wait))
	conn.SetWriteDeadline(time.Now().Add(wait))
	if user := GetOnLineUserByAddr(conn.RemoteAddr().String()); user != nil {
		if err := user.WsWriteMessage(websocket.PongMessage, []byte{}); err != nil {
			utils.Logger.Error(err)
		}
	} else if err := conn.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
		utils.Logger.Error(err)
	}
}

func MsgCache(user *OnLineUser, msg *Msg) {
	msgCell := MsgCell{User: user, Msg: msg}

	utils.Logger.Info(fmt.Sprintf("remote %s, uid %d, new Msg %+v", *user.Addr, user.Uid, *msgCell.Msg))

	msgQue <- &msgCell
}

func InitMsgServer() {
	if config.CFG.Lobby.Addr == "" && config.CFG.Server.Addr == "" {
		utils.Logger.Warn("will not InitMsgServer")
		return
	}
	var max = 100000
	if config.CFG.MsgMaxMain != 0 {
		max = config.CFG.MsgMaxMain
		utils.Logger.Warnf("MaxMain: %d", max)
	}
	msgQue = make(chan *MsgCell, max)

	go DispatchMsg()
}

func DispatchMsg() {
	for {
		msgC, ok := <-msgQue
		if !ok {
			utils.Logger.Warn("msgQue chan close!")
			return
		}
		utils.Logger.Debug(fmt.Sprintf("DispatchMsg msgC: %+v", msgC))
		DoMsg(msgC)

	}
}

const (
	SysType  = 1
	GameType = 2
)

func DoMsg(msgC *MsgCell) {
	if msgC != nil && msgC.Msg != nil {
		if msgC.Msg.MsgType == SysType {
			DoSysMsg(msgC)
		} else if msgC.Msg.MsgType == GameType {
			DoGameMsg(msgC)
		} else {
			utils.Logger.Errorf(fmt.Sprintf("unkown MsgType! msgC: %s", msgC.String()))
		}
		return
	}
	utils.Logger.Errorf(fmt.Sprintf("DoMsg faild! msgC: %s", msgC.String()))

}

func DoSysMsg(msgC *MsgCell) {
	DoSysCommand(msgC.User, msgC.Msg.Op, msgC.Msg.Data, "")
}

func DoGameMsg(msgC *MsgCell) {
	DoSysCommand(msgC.User, "opGame", msgC.Msg.Data, msgC.Msg.Op)
}
