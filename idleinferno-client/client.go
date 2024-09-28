package main

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/kvitebjorn/idleinferno/internal/requests"
)

type Client struct {
	ws *websocket.Conn
}

func (c *Client) Run() {
	log.Println("Starting idleinferno client...")
	u := url.URL{Scheme: "ws", Host: "10.0.0.33:12315", Path: "/ws"}
	log.Println("Dialing idleinferno server...")
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	c.ws = ws
	if err != nil {
		log.Fatalln("Failed to connect to idleinferno server: ", err.Error())
	}
	log.Println("Connected to idleinferno server!")

	// Send the initial hello to server
	log.Println("Performing handshake with idleinferno server...")
	var msg requests.Message
	msg.Message = "hi"
	msg.Code = requests.Salutations
	err = c.ws.WriteJSON(&msg)
	if err != nil {
		log.Fatalln("Failed to handshake with idleinferno server: ", err.Error())
	}
	log.Println("idleinferno server handshake successful!")
}
