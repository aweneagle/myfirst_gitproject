package	main
import	"./netevent"
import	"fmt"
import	"time"

func main() {
	ne := &netevent.NetEvent{}

	/*
	buff := make([]byte, 1024)
	ne.Conn(0).Connect("127.0.0.1", 9999)
	ne.Conn(0).Send([]byte("abc"))
	ne.Conn(0).Recv(buff)
	fmt.Println(string(buff))
	*/

	data := "1234567890"
	ne.Conn(0).OnConn = func(fd uint32) {
		fmt.Println(fd, " connected")
		ne.Conn(fd).Send([]byte(data))
		fmt.Println(fd, "send:", data)
	}

	ne.Conn(0).OnRecv = func(fd uint32, pack []byte) {
		fmt.Println(fd, " recv:", string(pack))
		ne.Conn(fd).Send(pack)
		time.Sleep(1 * time.Second)
	}

	ne.Conn(0).OnClose = func(fd uint32) {
		fmt.Println(fd, " closed")
	}

	ne.Conn(0).Connect("127.0.0.1", 9999)

	ne.Conn(0).Watch()
}
