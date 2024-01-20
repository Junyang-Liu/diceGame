package main

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	log.Printf("start")
	var addr = flag.String("addr", "localhost:8081", "http service address")
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
	ping(c)
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

	ping(c)
	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"getAllGame","type":2,"data":{"test":666}}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"newGameRoom","type":2,"data":{"game_id":101}}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	time.Sleep(1 * time.Second)

	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"enterGameRoom","type":2,"data":{"room_id":666661}}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	time.Sleep(1 * time.Second)

	game()

	ping(c)
	time.Sleep(5 * time.Second)
	// time.Millisecond
	// isfinish = true
	select {}

}

func ping(c *websocket.Conn) {
	err := c.WriteMessage(websocket.PingMessage, []byte{})
	log.Println("ping it")
	if err != nil {
		log.Println("ping:", err)
		return
	}
}

func game() {
	log.Printf("game")
	var addr = flag.String("game_addr", "localhost:8080", "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/dice"}
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

	time.Sleep(1 * time.Second)

	ping(c)
	err = c.WriteMessage(websocket.TextMessage, []byte(`{"op":"inGame","type":2, "data":666661}`))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	ping(c)
	time.Sleep(5 * time.Second)
	// time.Millisecond
	// isfinish = true
	for {
		ping(c)
		time.Sleep(5 * time.Second)
	}

}
