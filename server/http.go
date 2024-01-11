package server

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"

	"diceGame/config"
	"diceGame/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitHttpServer() error {
	LobbyServerAddr := config.CFG.Lobby.Addr
	if LobbyServerAddr != "" {
		lobbyMux := http.NewServeMux()
		lobbyMux.HandleFunc("/lobby", func(w http.ResponseWriter, r *http.Request) {

			utils.Logger.Debugf("RemoteAddr:%s", r.RemoteAddr)

			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				utils.Logger.Errorf(err.Error())
				conn.Close()
				return
			}

			conn.SetCloseHandler(func(code int, text string) error {
				utils.Logger.Infof("game socket offline code:%d, text:%s", code, text)
				return nil
			})

			RecvGameServerMsg(conn)

		})

		go http.ListenAndServe(LobbyServerAddr, handlers.LoggingHandler(os.Stdout, lobbyMux))
	}

	serverAddr := config.CFG.Server.Addr
	if serverAddr != "" {
		serverMux := http.NewServeMux()
		serverMux.HandleFunc("/dice", func(w http.ResponseWriter, r *http.Request) {

			utils.Logger.Debugf("RemoteAddr:%s", r.RemoteAddr)

			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				utils.Logger.Errorf(err.Error())
				conn.Close()
				return
			}

			RecvMsg(conn)

		})

		go http.ListenAndServe(serverAddr, handlers.LoggingHandler(os.Stdout, serverMux))

	}

	if config.CFG.Server.LobbyAddr != "" {
		InitClientToLobby()
	}
	return nil
}
