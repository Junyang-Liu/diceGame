package server

import (
	"bytes"
	"diceGame/utils"
	"fmt"

	"github.com/gorilla/websocket"
)

func RecvGameServerMsg(conn *websocket.Conn) {
	for {
		msgType, raw, err := conn.ReadMessage()
		if err != nil {
			utils.Logger.Errorf(fmt.Sprintf("game server offline %d-%s-%s-%s", msgType, conn.RemoteAddr(), raw, err.Error()))
			return
		}

		utils.Logger.Info(fmt.Sprintf("%d %s %s", msgType, conn.RemoteAddr(), raw))
		buffer := bytes.Buffer{}
		buffer.Write([]byte(`{"fast_label":0, "msg":`))
		buffer.Write(raw)
		buffer.Write([]byte(`}`))

		conn.WriteMessage(msgType, buffer.Bytes())
	}
}
