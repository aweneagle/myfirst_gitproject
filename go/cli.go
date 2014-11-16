package	main
import	"./netevent"
import	"fmt"
import	"time"

func main() {

	ne := netevent.Init()

	ne.OnConn = func(fd uint32) error {
		fmt.Println("connect connect:", fd)
		go func() {
			response := make([]byte, 128)
			for {
				ne.Request(fd, []byte("123456789000"), response)
				fmt.Println("thread 1:", string(response))
				time.Sleep(time.Second * 1)
			}
		}()
		for {
			response := make([]byte, 128)
			ne.Request(fd, []byte("abcdefghi"), response)
			fmt.Println("thread 2:", string(response))
			time.Sleep(time.Second * 1)
		}
		return nil
	}
	/*
	ne.OnConn = func(fd uint32) error {
		fmt.Println("connect:", fd)
		ne.Send(fd, []byte("abcdef"))
		return nil
	}
	*/

	ne.OnClose = func(fd uint32) error {
		ne.Close(fd)
		fmt.Println("connect closed:", fd)
		return nil
	}

	ne.OnRecv = func(fd uint32, pack []byte) error {
		fmt.Println("OnRecv:", string(pack))
		ne.Send(fd, pack)
		time.Sleep(1 * time.Second)
		return nil
	}

	ne.Dial("127.0.0.1:8888")


}
