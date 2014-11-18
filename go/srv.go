package	main
import	"./netevent"
import	"fmt"

func main() {
	n := netevent.Init()
	n.ConnBuffSize = 256
	n.Debug = true;
	n.OnConn = func(fd uint32) error {
		fmt.Println("connect connect:", fd)
		n.SetPackEof(fd, &netevent.PackEofHttpGet{}, &netevent.PackEofHttpPost{})
		return nil
	}
	n.OnClose = func(fd uint32) error {
		fmt.Println("connect closed:", fd)
		return nil
	}
	n.OnRecv = func(fd uint32, pack []byte) error {
		fmt.Println(string(pack))
		n.Send(fd, pack)
		return nil
	}

	n.Listen("127.0.0.1:8888")


}
