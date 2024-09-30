package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/kvitebjorn/idleinferno/internal/auth"
	"github.com/kvitebjorn/idleinferno/internal/requests"
)

type Client struct {
	ws            *websocket.Conn
	serverAddress string
	name          string
}

func (c *Client) Run() {
	log.Println("Starting idleinferno client...")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Server ip: ")
	rawHostname, _ := reader.ReadString('\n')
	parsedHostname := net.ParseIP(strings.TrimSpace(rawHostname))
	if parsedHostname == nil {
		log.Fatalln("Invalid server ip:", parsedHostname)
	}

	fmt.Print("Server port #: ")
	rawPortNumber, _ := reader.ReadString('\n')
	parsedPortNumber, err := strconv.Atoi(strings.TrimSpace(rawPortNumber))
	if err != nil {
		log.Fatalln("Invalid port number:", parsedPortNumber)
	}

	c.serverAddress = parsedHostname.String() + ":" + strconv.Itoa(parsedPortNumber)
	err = c.getPong()
	if err != nil {
		log.Fatalln("Unable to reach server.")
		return
	}
	log.Println("Connected to idleinferno server!")

	c.menu()

	log.Println("Listening for messages...")
	c.listen()
}

func (c *Client) menu() {
	fmt.Println("")
	fmt.Println("[l] login")
	fmt.Println("[i] info")
	fmt.Println("[u] sign up")
	fmt.Println("[q] quit")
	fmt.Println("")
	fmt.Print("→ ")
	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		log.Println(err)
		c.menu()
	}

	char = unicode.ToLower(char)
	switch char {
	case 'u':
		c.handleSignUp()
		c.menu()
		break
	case 'l':
		c.handleLogin()
		break
	case 'i':
		c.handleInfo()
		c.menu()
		break
	case 'q':
		c.handleQuit()
		break
	default:
		fmt.Println("Invalid menu selection.")
		c.menu()
	}
}

func isAlphanumeric(word string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(word)
}

func checkInput(s string) bool {
	if !isAlphanumeric(s) || len(s) == 0 {
		fmt.Println("Invalid input.")
		return false
	}
	return true
}

func (c *Client) handleLogin() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("")
	fmt.Println("Login")

	fmt.Print("username: ")
	rawUsername, _ := reader.ReadString('\n')
	trimmedUsername := strings.TrimSpace(rawUsername)
	isValid := checkInput(trimmedUsername)
	if !isValid {
		c.handleLogin()
	}

	fmt.Print("password: ")
	rawPassword, _ := reader.ReadString('\n')
	trimmedPassword := strings.TrimSpace(rawPassword)
	isValid = checkInput(trimmedPassword)
	if !isValid {
		c.handleLogin()
	}

	maybeUser, err := c.getUser(trimmedUsername)
	if err != nil {
		fmt.Println("Failed to get user.")
		c.menu()
	}

	hashedPassword := maybeUser.Password
	isMatch := auth.CheckHash(trimmedPassword, hashedPassword)
	if !isMatch {
		fmt.Println("Username or password does not match our records.")
		c.menu()
	}
	c.name = trimmedUsername

	// Set up the websocket
	u := url.URL{
		Scheme: "ws",
		Host:   c.serverAddress,
		Path:   "/ws",
	}
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	c.ws = ws
	if err != nil {
		log.Fatalln(
			"Failed to connect to idleinferno server:",
			err.Error(),
		)
	}

	fmt.Println("Performing handshake with idleinferno server...")
	var msg requests.UserMessage
	msg.Message = "hi"
	msg.Code = requests.Salutations
	msg.User = requests.User{Name: c.name, Password: trimmedPassword}
	err = c.ws.WriteJSON(&msg)
	if err != nil {
		fmt.Println("Failed to handshake with idleinferno server:", err.Error())
		c.menu()
	}

	// Listen for our online message just to make sure...
	playerMsg, err := c.receiveMessage()
	if err != nil {
		fmt.Println("Failed to log in.")
		c.menu()
	}
	log.Println(playerMsg)
	log.Println("idleinferno server handshake successful!")
}

func (c *Client) handleSignUp() {
	fmt.Println("")
	fmt.Println("Create a player")
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

func (c *Client) handleInfo() {
	fmt.Println("")
	fmt.Println("Get player info")
	//     ask for username to lookup
	//       should be > 0 alphanumeric only chars
	//     return a text dump of user, if it exists
	//       will need a new request for this - ReadPlayer (by 'name')
	//     return to main menu
}

func (c *Client) handleQuit() {
	c.disconnect()
}

func (c *Client) sendMessage(code requests.StatusCode, content string) error {
	var msg requests.PlayerMessage
	msg.Code = code
	msg.Message = content
	err := c.ws.WriteJSON(&msg)
	if err != nil {
		log.Println("Failed to send message to idleinferno server:", err.Error())
		return err
	}
	return nil
}

func (c *Client) receiveMessage() (requests.PlayerMessage, error) {
	var msg requests.PlayerMessage
	err := c.ws.ReadJSON(&msg)
	if err != nil {
		log.Println("Failed to read message from idleinferno server:", err.Error())
		return msg, err
	}
	return msg, nil
}

func (c *Client) getPong() error {
	requestURL := fmt.Sprintf("http://%s/ping", c.serverAddress)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New("not pong :(")
	}

	return nil
}

func (c *Client) getUser(name string) (*requests.User, error) {
	requestURL := fmt.Sprintf("http://%s/user/%s", c.serverAddress, name)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	user := &requests.User{}
	err = json.Unmarshal(resBody, user)

	if err != nil {
		return nil, err
	}

	return user, nil
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
		var msg requests.PlayerMessage
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Error receiving idleinferno server message:", err.Error())
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
	if c.ws != nil {
		c.ws.Close()
	}
	log.Println("Disconnected!")
	os.Exit(1)
}
