package main

func main() {
	server := initServer()
	server.Run()
}

func initServer() *Server {
	return &Server{}
}
