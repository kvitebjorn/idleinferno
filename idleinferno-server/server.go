package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"sync/atomic"

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
	db db.Database
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
	myRouter.HandleFunc("/", home)
	myRouter.HandleFunc("/ping", pong)
	myRouter.HandleFunc("/user/{name}", s.getUser)
	myRouter.HandleFunc("/ws", s.handleConnection)

	go handleMessages()

	fmt.Println("idleinferno server started on :12315")
	err := http.ListenAndServe(":12315", myRouter)
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
		log.Printf("%v %s\n", msg.Code, err.Error())
		return
	}

	user := msg.User
	maybeUser := s.db.ReadUser(user.Name)
	if maybeUser == nil {
		log.Println("User doesn't exist:", user.Name)
		return
	}
	if !auth.CheckHash(user.Password, maybeUser.Password) {
		log.Println("Invalid user credentials for", user.Name)
		return
	}
	if maybeUser.Online {
		log.Println(user.Name, "is already online.")
		return
	}
	err = s.db.UpdateUserOnline(user.Name)
	if err != nil {
		log.Println("Failed to come online for user", user.Name)
		return
	}

	userId := USER_COUNTER.Add(1)
	if userId == math.MaxUint64-1 {
		fmt.Println("Server full")
		return
	}

	player := requests.Player{Name: user.Name}
	client := Client{&player, conn}
	USERS_MU.Lock()
	USERS[userId] = &client
	USERS_MU.Unlock()

	connMsg := fmt.Sprintf("%s connected!", client.Player.Name)
	log.Println(connMsg)
	BROADCAST <- requests.PlayerMessage{Player: SERVER_PLAYER, Message: connMsg, Code: requests.Chatter}

	// Listen for messages, and add them to the broadcast channel to be fanned out
	for {
		var msg requests.PlayerMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			USERS_MU.Lock()
			delete(USERS, userId)
			USERS_MU.Unlock()

			_ = s.db.UpdateUserOffline(user.Name)
			log.Println(user.Name, "went offline.")

			disconnectMsg := fmt.Sprintf("%s disconnected!", client.Player.Name)
			BROADCAST <- requests.PlayerMessage{
				Player:  *client.Player,
				Message: "bye",
				Code:    requests.Valediction}
			BROADCAST <- requests.PlayerMessage{
				Player:  SERVER_PLAYER,
				Message: disconnectMsg,
				Code:    requests.Chatter}
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
	log.Println("Initializing database...")
	s.db = &sqlite.Sqlite{}
	s.db.Init()
	log.Println("Database initialized successfully!")

	log.Println("Starting idleinferno...")
	log.Println("Initializing world...")
	game := game.Game{World: s.initWorld()}
	log.Println("World initialized successfully!")

	// Start the request listener
	go s.Start()

	// Start the game
	log.Println("Running main game loop...")
	game.Run()
	log.Println("Main game loop exited.")

	log.Println("Saving the world...")
	s.saveWorld(game.World)
	log.Println("World saved!")

	err := s.db.Close()
	if err != nil {
		log.Fatalln("Error closing database:", err.Error())
	}
	log.Println("Exiting idleinferno.")
}

func (s *Server) initWorld() *model.World {
	players := s.db.ReadPlayers()
	items := s.db.ReadItems()

	world := &model.World{}

	for _, p := range players {
		world.Players = append(world.Players, p)
		world.PlayerGrid[p.Location.Y][p.Location.X] = p
	}
	for _, i := range items {
		world.Items = append(world.Items, i)
		world.ItemGrid[i.Location.Y][i.Location.X] = i
	}

	log.Println(world.ToString())
	return world
}

func (s *Server) saveWorld(world *model.World) {
	for _, player := range world.Players {
		_ = s.db.UpdatePlayer(player)
	}

	// TODO: save the rest of the state
	return
}
