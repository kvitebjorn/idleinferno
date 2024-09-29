package requests

type StatusCode uint8

const (
	Salutations StatusCode = iota
	Valediction
	Chatter
	Signup
)

type Player struct {
	Name string
}

type User struct {
	Name     string
	Email    string
	Password string
}

type Message struct {
	Player  Player     `json:"player"`
	Message string     `json:"message"`
	Code    StatusCode `json:"code"`
}

type Ping struct{}
