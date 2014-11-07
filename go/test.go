package	main
import	"./netevent"
import	"fmt"

func main() {
	ne := netevent.Init("*:8888")
	ne.OnConn = func(fd uint64) error {
		ne.SetPackEof(fd, &netevent.PackEofTail{ Eof:"___" })
		return nil
	}
	ne.OnClose = func(fd uint64) error {
		ne.Close(fd)
		return nil
	}
	ne.OnRecv = func(fd uint64, pack []byte) error {
		fmt.Println(string(pack))
		ne.Send(fd, pack)
		return nil
	}

	ne.OnConn(1)
	ne.OnClose(1)
	ne.OnRecv(1,[]byte("abcde"))

	ne.ErrorLog = "/tmp/abc.log"

	ne.TLogErr("aaa")

}
