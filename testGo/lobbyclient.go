package main

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8081", "http service address")

func main() {

	log.Printf("start")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/dice_lobby"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	c.SetCloseHandler(func(code int, text string) error {
		log.Fatal("closing ", code, text)
		return nil
	})

	isfinish := false
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			if isfinish {
				c.Close()
				break
			}
		}
	}()

	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"login","type":1, "data":100001}`))
	if err != nil {
		log.Println("write:", err)
		return
	}

	// err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"newGame","type":1}`))
	// if err != nil {
	// 	log.Println("write:", err)
	// 	return
	// }

	time.Sleep(1 * time.Second)

	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"inGame","type":2, "data":11}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"opGameTest","type":2,"data":{"test":666}}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"getAllGame","type":2,"data":{"test":666}}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}
	// c.Close()
	// time.Sleep(1 * time.Second)

	// err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"opLeave","type":2,"data":{"test":666}}`))
	// if err != nil {
	// 	log.Println("write close:", err)
	// 	return
	// }

	// time.Sleep(1 * time.Second)

	// err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"opReady","type":2,"data":{"test":666}}`))
	// if err != nil {
	// 	log.Println("write close:", err)
	// 	return
	// }

	time.Sleep(5 * time.Second)
	// time.Millisecond
	// isfinish = true
	select {}

}
