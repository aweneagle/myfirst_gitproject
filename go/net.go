package	main
import	"./netevent"

func main() {
	ne := netevent.Init()

	// global functions 
	ne.OnStart = func(){}
	ne.OnShutdown = func(){}	//catch USR2 signal
	ne.Host = "127.0.0.1"
	ne.Run()
	ne.Listen(9999, ne.SYNC | ne.ASYNC)


	// port functions 
	ne.Port(9999).OnStart = func(){}

	ne.Port(9999).OnShutdown = func(){}

	ne.Port(9999).OnConn = func(fd uint32) {}

	ne.Port(9999).OnClose = func(fd uint32) {}

	ne.Port(9999).OnRecv = func(fd uint32) {}


	// connect functions 
	ne.Conn(fd).OnRecv = func(fd uint32, pack []byte){}

	ne.Conn(fd).OnClose = func(fd uint32){}

	ne.Conn(fd).OnPackEof = func(stream []byte)(int, error){}


	ne.Conn(fd).Info()

	conn_info {
		SysFd
		SysProcId
		RemoteHost
		RemotePort
		LocalHost
		LocalPort
		SentBytes
		RecvBytes
		SentPacks
		RecvPacks
	}


	ne.Conn(fd).Connect("127.0.0.1", 8888)

	ne.Conn(fd).Dial("127.0.0.1", 8888)

	ne.Conn(fd).Send(data []byte)

	ne.Conn(fd).Close()


}
