package	main
import	"./netevent"
import	"fmt"
import	"time"

func main() {
	ne := netevent.Init("*:8888")

	ne.OnConn = func(fd uint32) error {
		fmt.Println("connect connect:", fd)
		ne.Send(fd, []byte("abcdefgh"))
		time.Sleep(time.Second * 1)
		return nil
	}

	ne.OnClose = func(fd uint32) error {
		ne.Close(fd)
		fmt.Println("connect closed:", fd)
		return nil
	}

	ne.OnRecv = func(fd uint32, pack []byte) error {
		fmt.Println(string(pack))
		ne.Send(fd, pack)
		return nil
	}

	ne.Connect("127.0.0.1:8888")
	ne.Start()


}
