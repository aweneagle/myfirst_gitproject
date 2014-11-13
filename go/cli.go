package	main
import	"./netevent"
import	"fmt"
import	"time"

func main() {

	ne := netevent.Init()

	ne.OnConn = func(fd uint32) error {
		fmt.Println("connect connect:", fd)
		ne.Send(fd, []byte("1234567890"))
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
		time.Sleep(time.Second * 1)
		return nil
	}

	ne.Dial("127.0.0.1:8888")


}
