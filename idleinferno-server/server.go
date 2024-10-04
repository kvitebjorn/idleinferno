package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kvitebjorn/idleinferno/internal/auth"
	"github.com/kvitebjorn/idleinferno/internal/db"
	"github.com/kvitebjorn/idleinferno/internal/db/sqlite"
	"github.com/kvitebjorn/idleinferno/internal/game"
	"github.com/kvitebjorn/idleinferno/internal/game/model"
	"github.com/kvitebjorn/idleinferno/internal/requests"
)

type Server struct {
	db              db.Database
	game            *game.Game
	broadcastBuffer *bytes.Buffer
}

type Client struct {
	Player *requests.Player
	Conn   *websocket.Conn
}

var (
	USER_COUNTER  atomic.Uint64
	SERVER_PLAYER = requests.Player{Name: "DANTE"}
)

func (s *Server) Start() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", home).Methods(http.MethodGet)
	myRouter.HandleFunc("/ping", pong).Methods(http.MethodGet)
	myRouter.HandleFunc("/user/{name}", s.getUser).Methods(http.MethodGet)
	myRouter.HandleFunc("/user/e/{email}", s.getUserByEmail).Methods(http.MethodGet)
	myRouter.HandleFunc("/user/create", s.createUser).Methods(http.MethodPost)
	myRouter.HandleFunc("/player/{name}", s.getPlayer).Methods(http.MethodGet)
	myRouter.HandleFunc("/ws", s.handleConnection)

	go handleMessages()
	go s.sendLogsToWebSocket()

	fmt.Println("idleinferno server started on :33379")
	err := http.ListenAndServe(":33379", myRouter)
	if err != nil {
		panic("Error starting idleinferno server: " + err.Error())
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	USERS_MU  sync.Mutex
	USERS     = make(map[uint64]*Client)
	BROADCAST = make(chan requests.PlayerMessage)
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to idleinferno!")
}

func pong(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Pong!")
}

func (s *Server) getPlayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["name"]
	maybePlayer := s.db.ReadPlayer(key)
	if maybePlayer == nil {
		return
	}
	encodedPlayer := requests.Player{
		Name:      maybePlayer.Name,
		Class:     maybePlayer.Class,
		Xp:        maybePlayer.Stats.Xp,
		Level:     maybePlayer.Stats.Level(),
		ItemLevel: maybePlayer.ItemLevel(),
		X:         maybePlayer.Location.X,
		Y:         maybePlayer.Location.Y,
		Created:   maybePlayer.Stats.Created,
		Online:    maybePlayer.Stats.Online,
	}
	json.NewEncoder(w).Encode(encodedPlayer)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["name"]
	maybeUser := s.db.ReadUser(key)
	if maybeUser == nil {
		return
	}
	encodedUser := requests.User{
		Name:     maybeUser.Name,
		Password: maybeUser.Password,
	}
	json.NewEncoder(w).Encode(encodedUser)
}

func (s *Server) getUserByEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["email"]
	maybeUser := s.db.ReadUserByEmail(key)
	if maybeUser == nil {
		return
	}
	encodedUser := requests.User{
		Name:     maybeUser.Name,
		Password: maybeUser.Password,
	}
	json.NewEncoder(w).Encode(encodedUser)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user requests.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	modelUser := model.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
		Class:    user.Class,
	}
	dbUser := s.db.CreatePlayer(&modelUser)
	if dbUser == nil {
		fmt.Println("Failed to create user", modelUser.Name)
		return
	}
	fmt.Println("Created user", modelUser.Name)
	json.NewEncoder(w).Encode(&user)
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Wait for initial hello message
	var msg requests.UserMessage
	err = conn.ReadJSON(&msg)
	if err != nil {
		fmt.Println("%v %s", msg.Code, err.Error())
		return
	}

	user := msg.User
	maybeUser := s.db.ReadUser(user.Name)
	if maybeUser == nil {
		fmt.Println("User doesn't exist:", user.Name)
		return
	}
	if !auth.CheckHash(user.Password, maybeUser.Password) {
		fmt.Println("Invalid user credentials for", user.Name)
		return
	}
	if maybeUser.Online {
		fmt.Println(user.Name, "is already online.")
		return
	}
	err = s.db.UpdateUserOnline(user.Name)
	if err != nil {
		fmt.Println(user.Name, "failed to come online")
		return
	}

	userId := USER_COUNTER.Add(1)
	if userId == math.MaxUint64-1 {
		log.Println("Server full")
		return
	}

	player := requests.Player{Name: user.Name}
	client := Client{&player, conn}
	USERS_MU.Lock()
	USERS[userId] = &client
	USERS_MU.Unlock()

	gamePlayer := s.db.ReadPlayer(player.Name)
	updatedGamePlayer, err := s.game.World.Login(gamePlayer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_ = s.db.UpdatePlayer(updatedGamePlayer)

	connMsg := fmt.Sprintf("%s connected!", client.Player.Name)
	log.Println(connMsg)

	// We send this because they will miss their own login broadcast message
	conn.WriteJSON(&requests.PlayerMessage{Player: SERVER_PLAYER, Message: connMsg, Code: requests.Chatter})

	// Listen for messages, and add them to the broadcast channel to potentially be fanned out
	for {
		var msg requests.PlayerMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			USERS_MU.Lock()
			delete(USERS, userId)
			USERS_MU.Unlock()

			s.game.World.Logout(gamePlayer)
			_ = s.db.UpdateUserOffline(user.Name)
			log.Println(user.Name, "went offline.")
			return
		}

		BROADCAST <- msg
	}
}

func handleMessages() {
	for {
		msg := <-BROADCAST

		USERS_MU.Lock()
		for _, user := range USERS {
			err := user.Conn.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
			}
		}
		USERS_MU.Unlock()
	}
}

func (s *Server) Run() {
	// Create a buffer to store logs
	s.broadcastBuffer = new(bytes.Buffer)

	// Tee the log output to both the console and the buffer
	log.SetOutput(io.MultiWriter(os.Stdout, s.broadcastBuffer))

	fmt.Println("Initializing database...")
	s.db = &sqlite.Sqlite{}
	s.db.Init()
	fmt.Println("Database initialized successfully!")

	fmt.Println("Starting idleinferno...")
	fmt.Println("Initializing world...")
	s.game = &game.Game{World: s.initWorld()}
	fmt.Println("World initialized successfully!")

	// Start the request listener
	go s.Start()

	// Start the signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			s.saveWorld(s.game.World)
			for _, p := range s.game.World.Players {
				s.db.UpdateUserOffline(p.Name)
			}
			log.Fatalln("Server interrupted.")
			os.Exit(1)
		}
	}()

	// Safety net log out all users on crash
	defer func() {
		s.saveWorld(s.game.World)
		for _, p := range s.game.World.Players {
			s.db.UpdateUserOffline(p.Name)
		}
		log.Println("Server crashed.")
	}()

	// Start the game
	fmt.Println("Running main game loop...")
	s.game.Run(s.saveWorld)
	fmt.Println("Main game loop exited.")

	fmt.Println("Saving the world...")
	s.saveWorld(s.game.World)
	fmt.Println("World saved!")

	err := s.db.Close()
	if err != nil {
		log.Fatalln("Error closing database:", err.Error())
	}
	fmt.Println("Exiting idleinferno.")
}

func (s *Server) initWorld() *model.World {
	world := &model.World{}

	log.Println(world.ToString())
	return world
}

func (s *Server) saveWorld(world *model.World) {
	for _, player := range world.Players {
		_ = s.db.UpdatePlayer(player)
	}

	return
}

func (s *Server) sendLogsToWebSocket() {
	for {
		// Periodically send logs from the buffer
		logData := strings.TrimSpace(s.broadcastBuffer.String())
		if logData != "" {
			msg := requests.PlayerMessage{
				Player:  SERVER_PLAYER,
				Message: logData,
				Code:    requests.Chatter,
			}
			BROADCAST <- msg
			s.broadcastBuffer.Reset()
		}

		time.Sleep(2 * time.Second)
	}
}
