package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kvitebjorn/idleinferno/internal/requests"
)

type Client struct {
	ws            *websocket.Conn
	serverAddress string
	username      string
}

func (c *Client) Run() {
	log.Println("Starting idleinferno client...")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Server ip: ")
	rawHostname, _ := reader.ReadString('\n')
	parsedHostname := net.ParseIP(strings.TrimSpace(rawHostname))
	if parsedHostname == nil {
		log.Fatalln("Invalid server ip: ", parsedHostname)
	}

	fmt.Print("Server port #: ")
	rawPortNumber, _ := reader.ReadString('\n')
	parsedPortNumber, err := strconv.Atoi(strings.TrimSpace(rawPortNumber))
	if err != nil {
		log.Fatalln("Invalid port number: ", parsedPortNumber)
	}
	c.serverAddress = parsedHostname.String() + ":" + strconv.Itoa(parsedPortNumber)
	u := url.URL{
		Scheme: "ws",
		Host:   c.serverAddress,
		Path:   "/ws",
	}

	log.Println("Dialing idleinferno server...")
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	c.ws = ws
	if err != nil {
		log.Fatalln(
			"Failed to connect to idleinferno server ",
			parsedHostname.String(),
			": ",
			err.Error(),
		)
	}
	log.Println("Connected to idleinferno server!")

	c.menu()

	// Send the initial hello to server
	log.Println("Performing handshake with idleinferno server...")
	var msg requests.Message
	msg.Message = "hi"
	msg.Code = requests.Salutations
	msg.User = requests.User{Username: c.username}
	err = c.ws.WriteJSON(&msg)
	if err != nil {
		log.Fatalln("Failed to handshake with idleinferno server: ", err.Error())
	}
	log.Println("idleinferno server handshake successful!")

	log.Println("Listening for messages...")
	c.listen()
}

func (c *Client) menu() {
	// TODO: add option to get info about char
	// TODO: text walkthrough:
	//   login   [l]
	//   info    [i]
	//   sign up [s]
	//   if l:
	//     ask for username
	//     ask for password
	//     auth
	//     if auth'ed, set c.username to username!
	//   if i:
	//     ask for username to lookup
	//       should be > 0 alphanumeric only chars
	//     return a text dump of user, if it exists
	//       will need a new request for this - ReadPlayer (by 'name')
	//     return to main menu
	//   if s:
	//     ask for username
	//       search for existing usernames, repeat question if it exists
	//       should be > 0 alphanumeric only chars
	//     ask for class
	//       should be > 0 alphanumeric only chars
	//     ask for email
	//       search for existing email, repeat question if it exists
	//       use an email validator, or a regex of form *@*.*
	//     ask for password
	//       should be > 0 non-white-space chars
	//     hash email and password
	//       will need to implement auth stuff for this
	//     store in db
	//       will need a new request for this - CreatePlayer
	//     tell them we successfully created the user
	//     return to main menu
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
