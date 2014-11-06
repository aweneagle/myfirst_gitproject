package	main
import	"fmt"

type ne struct{
	OnRecv	func([]byte, int64) error
	OnClose	func(int64) error
	OnConn	func(int64) error
}
func (n *net) Connect (string host, int64 port) (fd, error) {
}
func (n *net) Close (fd uint64) error {
}

type server struct {
	http_port	ne
	msg_port	ne

	online	map[uint64] uint64	//[userid]	=>	fd
	remap	map[uint64] uint64	//[fd]		=>	userid

	http_proxy	map[string] uint64	//[server_name]	=>	fd
	msg_proxy	map[string] uint64	//[server_name]	=>	fd

}

func (s *server) on_recv_http (pack []byte, fd int64) error {
	/* handle message here */
}

func (s *server) Start(is_async bool) error {
	/* set listener */
	s.http_port.OnRecv = s.on_recv_http
	s.msg_port.OnRecv = s.on_recv_msg

	/* connect to peer servers */
	s.http_proxy["s0"], _ = s.http_port.Connect("127.0.0.1", 2299)
	s.http_port.AddPackEof(s.http_proxy["s0"],
		ne.HttpPackEof{ head:"POST", tail:"\r\n\r\n" },
		ne.HttpPackEof{ head:"GET",	 tail:"\r\n\r\n" }
	);	//to ensure to receive the right package from socket

	// to communicate to peer servers
	req := HttpRequest { uri:"/server/join", method:"POST", params:"serve_name=s1" }
	s.http_port.Send(req.Bytes(), s.http_proxy["s0"])

	s.msg_proxy["s0"], _ = s.http_port.Connect("127.0.0.1", 2298)
	s.msg_port.AddPackEof(
		s.msg_proxy["s0"],
		ne.MsgPackEof{ head:"CHAT", tail:"\r\n\r\n" }
	)

	// proxy login
	login := MsgProxyLogin { server : "s0", pwd : "1234" }
	s.msg_port.Send(login.Bytes(), s.msg_proxy["s0"])

	err := s.http_port.Start(ne.RUN_SYNC)
	if err != nil {
		return err
	}
	if is_async {
		s.msg_port.Start(ne.RUN_ASYNC)
	} else {
		s.msg_port.Start(ne.RUN_SYNC)
	}
}
