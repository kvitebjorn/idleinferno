package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
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
	db   db.Database
	game *game.Game
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
		log.Println("Failed to create user", modelUser.Name)
		return
	}
	log.Println("Created user", modelUser.Name)
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

	gamePlayer := s.db.ReadPlayer(player.Name)
	updatedGamePlayer, err := s.game.World.Login(gamePlayer)
	if err != nil {
		log.Println(err.Error())
		return
	}
	_ = s.db.UpdatePlayer(updatedGamePlayer)

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

			s.game.World.Logout(gamePlayer)
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
	s.game = &game.Game{World: s.initWorld()}
	log.Println("World initialized successfully!")

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

	// Start the game
	log.Println("Running main game loop...")
	s.game.Run(s.saveWorld)
	log.Println("Main game loop exited.")

	log.Println("Saving the world...")
	s.saveWorld(s.game.World)
	log.Println("World saved!")

	err := s.db.Close()
	if err != nil {
		log.Fatalln("Error closing database:", err.Error())
	}
	log.Println("Exiting idleinferno.")
}

func (s *Server) initWorld() *model.World {
	items := s.db.ReadItems()

	world := &model.World{}

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
