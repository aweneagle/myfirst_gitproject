package	main
import	"./netevent"
import	"fmt"

func main() {
	ne := &netevent.NetEvent{}
	ne.Port(9999).OnConn = func(fd uint32) {
		fmt.Println("connected:", fd)
	}

	ne.Port(9999).OnRecv = func(fd uint32, data []byte) {
		fmt.Println("received:", string(data))
		fmt.Println("response ...")
		ne.Conn(fd).Send([]byte("data received succ ...!"))
	}

	ne.Port(9999).OnClose = func(fd uint32) {
		fmt.Println("closed:", fd)
	}

	ne.Port(9999).Listen("127.0.0.1")
}
