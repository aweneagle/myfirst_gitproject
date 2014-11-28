package	main
import	"./ne"

func main() {
	_ = ne.Init()
	/*

	// global functions 
	ne.OnStart = func()error{ return nil }
	ne.OnShutdown = func()error{ return nil }	//catch USR2 signal
	ne.Host = "127.0.0.1"
	ne.Run()
	ne.Listen(9999, ne.SYNC | ne.ASYNC)
	ne.Shutdown(9999)
	ne.Close(fd)


	// port functions 
	ne.Port(9999).OnStart = func()error{return nil}

	ne.Port(9999).OnShutdown = func()error{return nil}

	ne.Port(9999).OnConn = func(fd uint32)error{return nil}

	ne.Port(9999).OnClose = func(fd uint32)error{return nil}

	ne.Port(9999).OnRecv = func(fd uint32)error{return nil}


	// connect functions 
	ne.Conn(fd).OnRecv = func(fd uint32, pack []byte)error{return nil}

	ne.Conn(fd).OnClose = func(fd uint32)error{return nil}

	ne.Conn(fd).OnPackEof = func(stream []byte)(int, error){ return 0, nil }


	ne.Conn(fd).Info()


	ne.Conn(fd).Connect("127.0.0.1", 8888)

	ne.Conn(fd).Dial("127.0.0.1", 8888)

	ne.Conn(fd).Send(data []byte)

	ne.Conn(fd).Close()

	*/

}
