package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/kvitebjorn/idleinferno/internal/auth"
	"github.com/kvitebjorn/idleinferno/internal/requests"
)

type Client struct {
	ws            *websocket.Conn
	serverAddress string
	name          string
	userInput     string
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

	// Listen for input in a separate goroutine
	go c.handleInput()

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
	_, err = c.receiveMessage()
	if err != nil {
		fmt.Println("Failed to log in.")
		c.menu()
	}
	log.Println("idleinferno server handshake successful!")
}

func (c *Client) handleSignUp() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("")
	fmt.Println("Sign up")

	fmt.Print("username: ")
	rawUsername, _ := reader.ReadString('\n')
	trimmedUsername := strings.TrimSpace(rawUsername)
	isValid := checkInput(trimmedUsername)
	if !isValid {
		fmt.Println("Invalid username, alphanumeric chars only allowed.")
		c.handleSignUp()
	}
	maybeUser, _ := c.getUser(trimmedUsername)
	if maybeUser != nil {
		fmt.Println("User name already exists.")
		c.handleSignUp()
	}

	fmt.Print("email: ")
	rawEmail, _ := reader.ReadString('\n')
	trimmedEmail := strings.TrimSpace(rawEmail)
	_, err := mail.ParseAddress(trimmedEmail)
	if err != nil {
		fmt.Println("Invalid email.")
		c.handleSignUp()
	}
	maybeUser, _ = c.getUserByEmail(trimmedEmail)
	if maybeUser != nil {
		fmt.Println("There is already a user for this email.")
		c.handleSignUp()
	}

	fmt.Print("password: ")
	rawPassword1, _ := reader.ReadString('\n')
	trimmedPassword1 := strings.TrimSpace(rawPassword1)
	if len(trimmedPassword1) < 6 {
		fmt.Println("Password length requirement (6) not met.")
		c.handleSignUp()
	}

	fmt.Print("password again: ")
	rawPassword2, _ := reader.ReadString('\n')
	trimmedPassword2 := strings.TrimSpace(rawPassword2)

	if trimmedPassword1 != trimmedPassword2 {
		fmt.Println("Passwords do not match.")
		c.handleSignUp()
	}

	fmt.Print("class: ")
	rawClass, _ := reader.ReadString('\n')
	trimmedClass := strings.TrimSpace(rawClass)
	isValid = checkInput(trimmedClass)
	if !isValid {
		fmt.Println("Invalid class, alphanumeric chars only allowed.")
		c.handleSignUp()
	}

	hashedPassword, err := auth.Hash(trimmedPassword1)
	if err != nil {
		fmt.Println("Error hashing password:", err.Error())
		c.handleSignUp()
	}

	user := requests.User{
		Name:     trimmedUsername,
		Email:    trimmedEmail,
		Password: hashedPassword,
		Class:    trimmedClass,
	}

	err = c.createUser(&user)
	if err != nil {
		fmt.Println("Failed to create user:", err.Error())
		c.handleSignUp()
	}

	fmt.Println("Sign up successful, please log in!")
	c.handleLogin()
}

func (c *Client) handleInfo() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("")
	fmt.Println("Player info")

	fmt.Print("name: ")
	rawName, _ := reader.ReadString('\n')
	trimmedName := strings.TrimSpace(rawName)
	isValid := checkInput(trimmedName)
	if !isValid {
		fmt.Println("Invalid name, alphanumeric chars only allowed.")
		return
	}
	maybePlayer, err := c.getPlayer(trimmedName)
	if maybePlayer == nil || err != nil {
		fmt.Println("Can't find that player.")
		return
	}

	fmt.Println("class:", maybePlayer.Class)
	fmt.Println("xp:", maybePlayer.Xp)
	fmt.Println("level:", maybePlayer.Level)
	fmt.Println("item level:", maybePlayer.ItemLevel)
	fmt.Println("coordinates:", "(", maybePlayer.X, ",", maybePlayer.Y, ")")
	fmt.Println("created:", maybePlayer.Created)
	fmt.Println("online:", maybePlayer.Online)
}

func (c *Client) handleQuit() {
	c.disconnect()
}

func (c *Client) sendMessage(code requests.StatusCode, content string) error {
	var msg requests.PlayerMessage
	msg.Player = requests.Player{Name: c.name}
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

func (c *Client) getPlayer(name string) (*requests.Player, error) {
	requestURL := fmt.Sprintf("http://%s/player/%s", c.serverAddress, name)
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

	user := &requests.Player{}
	err = json.Unmarshal(resBody, user)

	if err != nil {
		return nil, err
	}

	return user, nil
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

func (c *Client) getUserByEmail(email string) (*requests.User, error) {
	requestURL := fmt.Sprintf("http://%s/user/e/%s", c.serverAddress, email)
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

func (c *Client) createUser(user *requests.User) error {
	requestURL := fmt.Sprintf("http://%s/user/create", c.serverAddress)
	jsonBody, err := json.Marshal(user)
	bodyReader := bytes.NewReader(jsonBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return err
	}

	return nil
}

func (c *Client) disconnect() {
	if c.ws != nil {
		c.ws.Close()
	}
	log.Println("Disconnected!")
	os.Exit(1)
}

func (c *Client) handleInput() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print the input prompt
		fmt.Print("[map|info] → ")
		input, _ := reader.ReadString('\n')
		c.userInput = strings.TrimSpace(input) // Store input in the struct field

		if c.userInput == "" {
			continue // Skip empty input
		}

		// Send the input as a message to the server
		err := c.sendMessage(requests.Chatter, c.userInput)
		if err != nil {
			log.Println("Error sending message:", err)
		}
		c.userInput = "" // Clear input after sending
	}
}

func (c *Client) listen() {
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
			// Save the current cursor position
			fmt.Print("\0337")

			fmt.Print("\033[2K\r")   // Clear the line entirely and reset cursor to start
			fmt.Println(msg.Message) // Print the server message

			// Restore the cursor to the original input line
			fmt.Print("\0338")

			// Clear the input line again
			fmt.Print("\033[2K\r")

			// Reprint the input prompt and the current user input
			fmt.Print("[map|info] → ")
			fmt.Print(c.userInput) // Make sure we're printing the current input buffer
			os.Stdout.Sync()
		default:
			// Handle other message types here
		}
	}
}
