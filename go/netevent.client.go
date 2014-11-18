package	main
import	"./netevent"
import	"fmt"
import	"time"
//import	"reflect"
//import	"net"

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
		buf := make([]byte, 128)
		sock := ne.GetSock(fd)
		sock.SetReadDeadline(time.Now().Add(1 * time.Second))
		num, err := sock.Read(buf)
		_, fe := reflect.TypeOf(err).MethodByName("Timeout")
		fmt.Println("error name is:", fe)
		if err != nil {
			fmt.Println("err:", err.Error(), err.(net.Error).Timeout())
		}
		fmt.Println("num:", num)
		return err
	}
	*/

	ne.OnClose = func(fd uint32) error {
		ne.Close(fd)
		fmt.Println("connect closed:", fd)
		return nil
	}

	ne.OnRecv = func(fd uint32, pack []byte) error {
		fmt.Println("OnRecv:", string(pack))
		resp := make([]byte, 128)
		for {
			ne.Request(fd, pack, resp)
			fmt.Println("response:", string(resp))
			time.Sleep(1 * time.Second)
		}
		return nil
	}

	ne.Dial("127.0.0.1:8888")


}
