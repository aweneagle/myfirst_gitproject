/* server of chat 
 *
 * by awen, 2014.10.30
 */

package chat

import	"fmt"
import	"net"
import	"strconv"
import	"strings"
import	"os"

const IDS_PAGE_SIZE = 1024

const P_ERR_WRONG_RECEIVER_UID = 100
const P_ERR_WRONG_RECEIVER_GROUP = 101
const P_ERR_LISTEN_HTTP_PORT = 102
const P_ERR_LISTEN_PORT = 103
const P_ERR_WRONG_RECEIVER_PROXY = 104


type server struct {

 online	map[uint64]*connect	// clients
 on_other_line	map[uint64]string	// clients on other proxy servers, [userid] => server_name

 peer_servers_http	map[string]*http	// [server_name] => *http
 peer_servers_message	map[string]*connect	// [server_name] => *connect,  proxy server message channel
 peer_servers	map[string]string	//[server_name] => "host:port"

 message_channel_on	bool
 http_channel_on	bool

 groups	map[uint64][uint64]uint4	// map[ group_id ][ userid ] = 1
 on_other_groups	map[uint64][string]uint4	//groups on other proxy servers, map[ group_id ][ server_name ] = 1

 host	string
 port	uint64
 server_name	string	//length of 3, for example "s01"
 pwd	string	//length of 4+


 error_log	string	//file path
 info_log	string	//file path

}

func (s *server) log_error(err interface{}) {
	s.log(s.error_log, err)
}
func (s *server) log_info(err interface{}) {
	s.log(s.info_log, err)
}

func (s *server) log(file_path string, err interface{} ) {
	fd,_ := os.OpenFile(s.error_log, O_CREATE|O_APPEND|O_RDWR, 0666)
	switch err.(type) {
	case string :
		fd.Write([]byte(err))

	case error :
		fd.Write([]byte(err.Error))
	}
	fd.Close()
}


func (s *server) start_up_listener() {
	tcp_addr, err := net.ResolveTCPAddr("tcp", s.host + ":" + strconv.FormatUint(s.port, 10))
	if err != nil {
		panic("listen")
	}

	listener, err := net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		panic("listen addr")
	}

	for {
		sock, err := listener.Accept()
		if err != nil {
			s.log_error(err)
			continue;
		}
		go s.handle_tcp_sock(sock)
	}
}

func (s *server) handle_tcp_sock(c *net.Conn) {
	var err error = nil
	stream := create_stream(c)
	defer func() {
		stream.Close()
		s.log_error(err)
	}()

	var buff [4]byte
	_, err = stream.Read(buff[0:])
	if err != nil {
		return
	}

	stream.push_back(buff[0:])
	switch string(buff) {
	case "CHAT" :
		if ! s.message_channel_on {
			s.log_error("message channel closed")
			return
		}
		conn := new connect(stream)
		cmd := conn.login()
		s.handle_message_command(cmd, conn)
		for {
			cmd = conn.read_cmd()
			res := s.handle_message_command(cmd, conn)
			if res != 0 {
				break;
			}
		}

	case "POST" :
	case "GET " :
		if ! s.http_channel_on {
			s.log_error("http channel closed")
			return
		}
		conn := new http(stream)
		for {
			request := conn.read_request()
			resp := s.handle_http_request(request, conn)
			if resp == nil {
				break;
			}
			conn.write_response(resp)
		}

	default:
		s.log_error("wrong protocol")
		return
	}

}

func (s *server) handle_message_command(cmd *command, c *connect) int {
	succ := 0

	switch cmd.(type) {
	case pk_proxy_login :
		s.peer_server_message[cmd.server_name] = c

	case pk_user_login :
		s.online[c.userid] = c
		s.notify_peer_servers("/user/login?uid=" + strconv.FormatUint(c.userid, 10) + "&server=" + s.server_name)

	case v1_pk_data :
		for _,proxy := range s.peer_server_message {
			proxy.write_cmd(cmd)
		}

	case v1_pk_heartbeat :
		log_debug("heart beat:", cmd.userid)

	case pk_user_logout :
		delete(s.online, c.login.userid)
		log_info("logout:", c.login.userid)
		succ = -1

	case pk_proxy_logout :
		delete(s.peer_server_message, c.login.server_name)
		log_info("logout:", c.login.server_name)
		succ = -1

	default :
		s.log_error("wrong package")
	}

	return succ
}

func (s *sever) handle_http_request(req *http_request, h *http) resp string {
	resp := ""

	switch req.uri {
	case "/server/leave":
		server_name := req.params["name"]
		delete(s.peer_servers, server_name)
		delete(s.peer_servers_http, server_name)

	case "/server/join":
		server_name := req.params["name"]
		server_host := req.params["host"]
		server_port := req.params["port"]
		s.peer_servers[server_name] = server_host + ":" + server_port
		s.peer_servers_http[server_name] = h

	case "/user/login":
		uid, _ := strconv.ParseUint(req.params["uid"], 10, 64)
		s.on_other_line[uid] = req.params["server"]

	case "/user/logout":
		uid, _ := strconv.ParseUint(req.params["uid"], 10, 64)
		delete(s.on_other_line, uid)

	case "/group/join":
		uid, _ := strconv.ParseUint(req.params["uid"], 10, 64)
		gid, _ := strconv.ParseUint(req.params["gid"], 10, 64)
		s.groups[gid][uid] = 1

	case "/group/leave":
		uid, _ := strconv.ParseUint(req.params["uid"], 10, 64)
		gid, _ := strconv.ParseUint(req.params["gid"], 10, 64)
		delete(s.group[gid], uid)
		if len(s.group[gid]) == 0 {
			delete (s.group, gid)
			s.notify_peer_servers("/group/off_server?gid=" + strconv.FormatUint(gid, 10) + "&server=" + s.server_name)
		}

	case "/group/on_server":
		gid, _ := strconv.ParseUint(req.params["uid"], 10, 64)
		s.on_other_groups[gid] = req.params["server"]

	case "/group/off_server":
		gid, _ := strconv.ParseUint(req.params["uid"], 10, 64)
		delete(s.on_other_groups[gid], req.params["server_name"])

	case "/user/get_ids":
		page, _ := strconv.ParseUint(req.params["page"], 10, 64)
		begin := page * IDS_PAGE_SIZE
		end := begin + IDS_PAGE_SIZE
		i := 0
		for uid,_ := range s.online {
			if (i < begin ) {
				continue
			}
			if (i >= end) {
				break;
			}
			resp += strconv.FormatUint(uid, 10) + ","
			i ++
		}

	case "/group/get_ids":
		page, _ := strconv.ParseUint(req.params["page"], 10, 64)
		begin := page * IDS_PAGE_SIZE
		end := begin + IDS_PAGE_SIZE
		i := 0
		for uid,_ := range s.groups {
			if (i < begin ) {
				continue
			}
			if (i >= end) {
				break;
			}
			resp += strconv.FormatUint(uid, 10) + ","
			i ++
		}

	}

	return resp
}

func (s *server) AddPeerServer(name string, host_port string) {
	if s.peer_servers == nil {
		s.peer_servers = make(map[string]string)
	}
	s.peer_servers[name] = host_port
}

func (s *server) notify_peer_servers(uri string) {
	for _, http := range s.peer_servers_http {
		http.request(uri)
	}
}

/* run a server
*/
func (s *server) Start() {
	s.online = make(map[uint64]*connect)
	s.on_other_line = make(map[uint64]string)
	s.peer_server_http = make(map[string]*http)
	s.peer_server_message = make(map[string]*http)
	s.groups = make(map[uint64][uint64]uint4)
	s.on_other_groups = make(map[uint64][string]string)
	s.http_channel_on = false
	s.message_channel_on = false

	if s.peer_servers == nil {
		s.peer_servers = make(map[string]string)
	}

	// start up listener
	go s.start_up_listener()
	s.http_channel_on = true

	// connect to other peer servers
	for server_name,host_port := range s.peer_servers {
		conn, err := net.Dial("tcp", host_port)
		if err != nil {
			panic ("failed to join in peer servers:" + err.Error())
		}
		stream := create_stream(conn)
		s.peer_servers_http[server_name] = create_http(stream)
	}
	s.notify_peer_servers("/server/join?name=" + s.server_name + "&host=" + s.host + "&port" + strconv.FormatUint(s.port, 10))

	// read all groups from other peers
	for server_name,http := range s.peer_servers_http {
		for i := 0 ; true ; i ++ {
			groups := http.request("/group/get_ids?page=" + strconv.FormatUint(i, 10))
			if groups == "" {
				break;
			}
			for _,gid := range strings.Split(groups, ",") {
				if gid != "" {
					groupid, _ :=  strconv.ParseUint(gid, 10, 64)
					s.on_other_groups[groupid][ server_name ] = 1
				}
			}
		}

		for i := 0 ; true ; i ++ {
			users := http.request("/user/get_ids?page=" + strconv.FormatUint(i, 10))
			if users == "" {
				break;
			}
			for _,uid := range strings.Split(users, ",") {
				if uid != "" {
					userid, _ :=  strconv.ParseUint(uid, 10, 64) 
					s.on_other_line[userid][ server_name ] = 1
				}
			}
		}
	}

	// open message port
	s.message_channel_on = true

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
			cmd := conn.ReadIn_V1()
			conn.Handle(cmd, s)
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

/* send package to proxy 
 */
func (s *server) send_to_proxy(pk *pk_data, proxy string) {
	conn, exist := s.peer_server[ proxy ]
	if !exist {
		panic(P_ERR_WRONG_RECEIVER_PROXY]
	}
	conn.Write(pk.orig_pack)
}

/* send package to user
 */
func (s *server) send_to_user (pk *pk_data) {
	conn, exist := s.online[ pk.receiver ]

	if !exist {
		proxy, p_exist := s.on_other_line[ pk.receiver ]
		if !p_exist {
			panic(P_ERR_WRONG_RECEIVER_UID)
		}
		s.send_to_proxy(pk, proxy)
		return
	}
	conn.Write(pk.orig_pack)
}

/* send package to group 
 */
 func (s *server) send_to_group(pk *pk_data) {
	 group, exist := s.groups[ pk.receiver ]
	 if exist {
		 for userid,_ := range group {
			 conn, exist := s.online[ userid ]
			 if !exist {
				 fmt.Println("wrong receiver userid", userid)
				 continue
			 }

			 conn.Write(pk.orig_pack)
		 }
	 }
	 proxy, p_exist := s.on_other_group[ pk.receiver ]
	 if p_exist {
		for proxy,_ := range s.on_other_groups[ pk.receiver ] {
			s.send_to_proxy(pk, proxy)
		}
	 }
	 return
 }


/* add a client 
*/
func (s *server) add_client(userid uint64, c *connect) {
	s.online[ userid ] = c
}
