/* server of chat 
 *
 * by awen, 2014.10.30
 */

package chat

import	"fmt"
import	"net"

const P_ERR_WRONG_RECEIVER_UID = 100
const P_ERR_WRONG_RECEIVER_GROUP = 101
const P_ERR_LISTEN_HTTP_PORT = 102
const P_ERR_LISTEN_PORT = 103


type server struct {

 online	map[uint64]*connect	// clients
 on_other_line map[uint64]string	// clients on other proxy servers, [userid]=> "host:port"

 http	map[string]*http	// [host":"port] => *http


 peer_servers	map[string]*connect	// proxy server message channel

 groups	map[uint64][uint64]uint4	// map[ group_id ][ userid ] = 1
 on_other_groups map[uint64][string]uint4	//groups on other proxy servers, map[ group_id ][ server ] = 1


 host	string
 port	uint64
 http_port	uint64


 pwd	string	//length of 32

 peer_servers_message map[string]string	// ["host:port"] "host:port"
 peer_servers_http	map[string]string	// ["host:port"] "host:port"

 error_log	//log file path

}


/* run a server
*/
func (s *server) Start() {
	s.online = make(map[uint64]*connect)
	s.on_other_line = make(map[uint64]string)
	s.http = make(map[string]*http)
	s.peer_servers = make(map[string]*connect)
	s.groups = make(map[uint64][uint64]uint4)

	// open http port
	s.open_http_port()

	// connect to other peer servers
	for server_str,_ := range s.peer_servers_http {
		conn, err := net.Dial("tcp", server_str)

		if err != nil {
			s.log_error(err.Error())
			continue
		}

		s.http[server_str] = CreateHttp(conn, server_str)
		s.http[server_str].request("/new_peer_server?host=" + s.host + "&port=" + strconv.FormatUint(s.port, 10) + "&http_port=" + strconv.FormatUint(s.http_port))
	}

	// read all groups, users from other peers
	for server_str,http := range s.http {
		groups := http.request("/all_groups")

		for _,gid := range groups.resp {
			s.on_other_groups[ strconv.ParseUint(gid, 10, 64) ][ http.server ] = 1

			users := http.request("/group_members?gid=" + gid)

			for _,uid := range users.resp {
				s.on_other_line[ strconv.ParseUint(uid, 10, 64) ][ http.server ] = 1
			}
		}

	}

	// open message port
	s.open_message_port()

}

/* open a http port
*/
func (s *server) open_http_port() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", s.host +":"+ strconv.FromatUint(s.http_port, 10))
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		panic(P_ERR_LISTEN_HTTP_PORT)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.log_error(err.Error())
			continue
		}

		go s.handle_http_conn(conn)
	}
}

/* open message port
*/
func (s *server) open_message_port() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", s.host +":"+ strconv.FromatUint(s.port, 10))
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		panic(P_ERR_LISTEN_PORT)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.log_error(err.Error())
			continue
		}

		go s.handle_message_conn(conn)
	}
}

func (s *server) handle_message_conn(c *net.Conn) {
	conn = CreateConn(c)

	var login pk_login

	conn.Login(&login)

	switch login_pk.role {
	case ROLE_CLIENT:
		s.online[login.userid] = conn

	case ROLE_PROXY:
		s.peer_servers[login.server] = conn
	}

	switch login.version {
	case "10":
		for {
			cmd := conn.
		}

	case "20":
	default:
		fmt.Println("unsupport protocol version")
	}
}

func (s *server) handle_http_conn(c *net.Conn) {
	http = CreateHttp(c)

	for {
		cmd := http.read_in()
		if cmd == nil {
			http.end()
			break;
		}
		response := cmd.handle(s)
		http.write_out(response)
	}
}


/* send package to user
 */
func (s *server) send_to_user (pk *pk_data) {
	conn, exist := s.online[ pk.receiver ]

	if !exist {
		panic(P_ERR_WRONG_RECEIVER_UID)
	}

	conn.Write(pk.orig_pack)
}

/* send package to group 
 */
func (s *server) send_to_group(pk *pk_data) {
	group, exist := s.groups[ pk.receiver ]
	if !exist {
		panic(P_ERR_WRONG_RECEIVER_GROUP)
	}

	for userid,_ := range group {
		conn, exist := s.online[ userid ]
		if !exist {
			fmt.Println("wrong receiver userid", userid)
			continue
		}

		conn.Write(pk.orig_pack)
	}
}


/* add a client 
*/
func (s *server) add_client(userid uint64, c *connect) {
	s.online[ userid ] = c
}
