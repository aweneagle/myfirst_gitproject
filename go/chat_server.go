package main
import "./chat"
import "time"

func main() {
	serv := chat.NewServer("127.0.0.1", 9999)
	serv.Start()
}

