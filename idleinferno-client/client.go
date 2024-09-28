package main

import (
	"log"
	"net/url"
	"time"

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

	// TODO: specify username
	// Send the initial hello to server
	log.Println("Performing handshake with idleinferno server...")
	var msg requests.Message
	msg.Message = "hi"
	msg.Code = requests.Salutations
	msg.User = requests.User{Username: "testclient"}
	err = c.ws.WriteJSON(&msg)
	if err != nil {
		log.Fatalln("Failed to handshake with idleinferno server: ", err.Error())
	}
	log.Println("idleinferno server handshake successful!")

	// Start listening for server messages
	c.listen()

	// TODO: set up a buffer to send our own messages to server
	// probalby do `go c.listen()` after this
}

func (c *Client) sendMessage(s string) error {
	var msg requests.Message
	msg.Code = requests.Chatter
	msg.Message = s
	err := c.ws.WriteJSON(&msg)
	if err != nil {
		log.Println("Failed to send message to idleinferno server: ", err.Error())
		return err
	}
	return nil
}

func (c *Client) listen() {
	for {
		if c.ws == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}

	for {
		var msg requests.Message
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Error receiving idleinferno server message: ", err.Error())
			c.disconnect()
			return
		}
		switch msg.Code {
		case requests.Chatter:
			log.Println(msg.Message)
		default:
		}
	}
}

func (c *Client) disconnect() {
	if c.ws == nil {
		return
	}
	c.ws.Close()
	log.Println("Disconnected!")
}
