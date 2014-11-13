package	main
import	"./netevent"
import	"fmt"

func main() {
	ne := netevent.Init()
	ne.ConnBuffSize = 256
	ne.Debug = true;
	ne.OnConn = func(fd uint32) error {
		fmt.Println("connect connect:", fd)
		ne.SetPackEof(fd, &netevent.PackEofHttpGet{}, &netevent.PackEofHttpPost{})
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

	ne.Listen("127.0.0.1:8888")


}
