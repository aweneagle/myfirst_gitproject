/* server of chat 
 *
 * by awen, 2014.10.30
 */

package chat

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
 peers	map[string]*connect	// [server_name] => *connect,  proxy server message channel
 peer_servers	map[string]string	//[server_name] => "host:port"

 message_channel_on	bool
 http_channel_on	bool

 groups	map[uint64]map[uint64]uint8	// map[ group_id ][ userid ] = 1
 on_other_groups	map[uint64]map[string]uint8	//groups on other proxy servers, map[ group_id ][ server_name ] = 1

 host	string
 port	uint64
 server_name	string	//length of 3, for example "s01"
 pwd	string	//length of 4+


 error_log	string	//file path
 info_log	string	//file path

}

func (s *server) log_error(err interface{}) {
	if err != nil {
		s.log(s.error_log, err)
	}
}
func (s *server) log_info(err interface{}) {
	if err != nil {
		s.log(s.info_log, err)
	}
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

	var buff [6]byte
	read, err = stream.Read(buff[0:])
	if err != nil {
		s.log_error("read package header:" + err.Error())
		return
	}

	stream.push_back(buff[0:read])

	/* handle message socket */
	chat_reg := regexp.Compile("^.*CHAT$")
	chat_found := chat_reg.Find(buff)
	if chat_found != nil {
		if ! s.message_channel_on {
			s.log_error("message channel closed")
			return
		}
		conn := create_connect(stream)
		cmd := conn.accept_login()
		s.handle_message_command(cmd, conn)
		for {
			cmd = conn.read_cmd()
			res := s.handle_message_command(cmd, conn)
			if res != 0 {
				break;
			}
		}
		return

	}

	/* find out wrong protocol received */
	get_reg := regexp.Compile("^GET\\s.*$")
	get_found := get_reg.Find(buff)
	if get_found == nil {

		post_reg := regexp.Compile("^POST\\s.*$")
		post_found := post_reg.Find(buff)

		if post_found == nil {
			s.log_error("wrong protocol received")
			return
		}
	}

	/* handle http socket */
	if ! s.http_channel_on {
		s.log_error("http channel closed")
		return
	}
	conn := create_http(stream)
	for {
		request := conn.read_request()
		resp := s.handle_http_request(request, conn)
		if resp == nil {
			/* to close this socket */
			break;
		}
		conn.write_response(resp)
	}


}

func (s *server) send_cmd_to_proxy( server_name string, c *cmd ) error {
	proxy, e := s.peers[ server_name ]
	if e {
		proxy.write_cmd(cmd)
		return nil
	}
	return errors.New("no proxy found:" + server_name)
}

func (s *server) send_group_message_to_user(cmd *command) {
	members, exist := s.groups[cmd.receiver]
	if exist {
		for uid, _ := range s.groups[cmd.receiver] {
			user, e := s.online[uid]
			if e {
				user.write_cmd(cmd)
			} else {
				s.log_error("group member not exist:" + strconv.FormatUint(uid))
			}
		}
	}
}

func (s *server) send_group_message_to_proxy(cmd *command) {
	proxy_groups, e := s.on_other_groups[cmd.receiver]
	if e {
		for proxy_server_name, _ := range proxy_groups {
			err := s.send_cmd_to_proxy(proxy_server_name, cmd)
			if err != nil {
				s.log_error(err)
			}
		}
	}
}

func (s *server) send_user_message_to_user(cmd *command) {
	/* send to client */
	user, exist := s.online[cmd.receiver]
	if exist {
		err := s.online[cmd.receiver].write_cmd(cmd)
		if err != nil {
			s.log_error(err)
		}

	} else {
		s.log_error("user not found:" + strconv.FormatUint(cmd.receiver, 10))
	}
}

func (s *server) send_user_message_to_proxy(cmd *command) {
	user_proxy_server_name, proxy_exist := s.on_other_line[cmd.receiver]
	if !proxy_exist {
		s.log_error("user's proxy not found:" + strconv.FormatUint(cmd.receiver, 10))
		return
	}

	/* send message to proxy */
	err := s.send_cmd_to_proxy(user_proxy_server_name, cmd)
	if err != nil {
		s.log_error(err)
	}
}


func (s *server) handle_message_command(cmd *command, c *connect) int {
	succ := 0

	switch cmd.(type) {
	case pk_proxy_login :
		s.peers[cmd.server_name] = c

	case pk_user_login :
		s.online[c.userid] = c
		s.notify_peer_servers("/user/login?uid=" + strconv.FormatUint(c.userid, 10) + "&server=" + s.server_name)


	case v1_pk_data :
		switch cmd.receiver_type {

		case RECV_TYPE_USER :
			switch c.login.(type) {
			case *pk_proxy_login:
				s.send_user_message_to_proxy(cmd)

			case *pk_user_login:
				s.send_user_message_to_user(cmd)
			}

		case RECV_TYPE_GROUP :
			switch c.login.(type) {
			case *pk_proxy_login:
				s.send_group_message_to_proxy(cmd)

			case *pk_user_login:
				s.send_group_message_to_user(cmd)
			}

		}


	case v1_pk_heartbeat :
		log_debug("heart beat:", cmd.userid)

	case pk_user_logout :
		delete(s.online, c.login.userid)
		log_info("logout:", c.login.userid)
		succ = -1

	case pk_proxy_logout :
		delete(s.peers, c.login.server_name)
		log_info("logout:", c.login.server_name)
		succ = -1

	default :
		s.log_error("wrong package")
	}

	return succ
}

func (s *sever) handle_http_request(req *http_request, h *http) string {
	resp := ""

	switch req.uri {
	case "/server/leave":
		server_name := req.params["name"]
		delete(s.peer_servers, server_name)
		delete(s.peer_servers_http, server_name)
		resp = nil

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

	case "/halt" :
	default :
		resp = nil

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
func (s *server) Start() error {
	s.online = make(map[uint64]*connect)
	s.on_other_line = make(map[uint64]string)
	s.peer_server_http = make(map[string]*http)
	s.peers = make(map[string]*http)
	s.groups = make(map[uint64][uint64]uint4)
	s.on_other_groups = make(map[uint64][string]string)


	// test error_log and info_log
	fd,err := os.OpenFile(s.error_log, O_CREATE|O_APPEND|O_RDWR, 0666)
	if err != nil {
		return errors.New("failed to open error_log:" + err.Error())
	}
	fd.Close()

	fd,err = os.OpenFile(s.info_log, O_CREATE|O_APPEND|O_RDWR, 0666)
	if err != nil {
		return errors.New("failed to open info_log:" + err.Error())
	}
	fd.Close()


	if s.peer_servers == nil {
		s.peer_servers = make(map[string]string)
	}

	// before starting up port listener, we must set the two channels (message, http) closed first 
	s.http_channel_on = false
	s.message_channel_on = false

	// start up listener
	go s.start_up_listener()

	s.http_channel_on = true

	// connect to other peer servers on http channels 
	for server_name,host_port := range s.peer_servers {
		conn, err := net.Dial("tcp", host_port)
		if err != nil {
			panic ("failed to join in peer servers:" + err.Error())
		}
		stream := create_stream(conn)
		s.peer_servers_http[server_name] = create_http(stream)
	}
	s.notify_peer_servers("/server/join?name=" + s.server_name + "&host=" + s.host + "&port" + strconv.FormatUint(s.port, 10))

	// connect to other peer servers on message channels
	for server_name, host_port := range s.peer_servers {
		conn, err := net.Dial("tcp", host_port)
		if err != nil {
			panic ("failed to join in peer servers:" + err.Error())
		}
		connect := create_connect(create_stream(conn))

		login := new (pk_proxy_login)
		login.server = s.server_name
		login.pwd = "123456"

		err = stream.write_cmd(login)
		if err != nil {
			s.log_error("failed to send command:" + err.Error())
			continue
		}
		s.peers[server_name] = connect

	}

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
					s.on_other_line[userid] = server_name
				}
			}
		}
	}

	// open message port
	s.message_channel_on = true

	return nil
}

