/* a simple chat message server
 *
 * 支持分布式服务器服务
 * by awen, 2014.11.13
 */


package	chat
import	"../netevent/netevent"
import	"strconv"
import	"fmt"
import	"time"
import	"strings"

type	server	struct {

	ne	netevent.NetEvent

	/* 连接在本服务器的用户, userid => fd */
	online	map[uint32]uint32

	/* 连接在其它服务器的用户, userid => server_name */
	neighbors	map[uint32]string

	/* 群用户, group_id => userid */
	groups	map[uint32][]uint32

	/* 登录在其它服务器上的群, group_id => server_name */
	neighbor_groups	map[uint32]string

	/* 与其他服务器通讯的http连接 server_name => fd */
	peers_http map[string]uint32

	open_http_port	bool

	Host	string
	Port	int

	/* 本服务器最大用户数，默认为 65535 */
	MaxUserNum	uint32

	/* 其它服务器 server_name => host_port */
	PeerServers	map[string]string

	/* 服务器名称  */
	ServerName	string

}


func Server() *server {
	s = server{
		MaxUserNum: 65535
	}
	s.online = make(map[uint32]uint32)
	s.groups = make(map[uint32]uint32)
	s.neighbors = make(map[uint32]string)
	s.neighbor_groups = make(map[uint32]string)
	s.open_http_port = false
	s.open_message_port = false

	s.PeerServers = make(map[string]string)
	s.ServerName = ""
	s.Host = ""
	s.Port = 0

	return s
}

func (s *server) log_error(msg ... string) {
	now := time.Now()
	line := "CHAT|" + now.Format("2006-01-02 15:04:05") + "|ERROR|" + strings.Join(msg, "|")
	fmt.Println(line)
}

func (s *server) Start() {
	if s.ServerName == "" {
		s.log_error("no server name found")
		return
	}
	if s.Host == "" {
		s.log_error("no server host found")
		return
	}
	if s.Port == 0 {
		s.log_error("no server port found")
		return
	}

	http := netevent.Init()
	http.MaxConnNum = s.MaxUserNum
	http.OnRecv = s.on_recv
	http.OnConn = s.on_conn
	http.OnClose = s.on_close

	s.ne = http
	s.open_http_port = true

	// 开启http服务，加入服务器群，让peer servers能够把最新的用户信息和群信息发过来
	for name,peer := range s.PeerServers {
		fd, err := http.Connect(peer)
		if err != nil {
			s.log_error(err.Error())
			return
		}
		s.peers_http[name] = fd

		//返回的数据将被传递到回调函数 OnRecv 中处理
		s.http_request(name, "/server/join?server_name=" + s.ServerName + "&host=" + s.Host + "&port=" + strconv.FormatInt(s.Port, 10))
	}

	// 从peer servers中读取 用户群信息 和 用户信息，同步到本服务器
	for name, fd := range s.peers_http {
		var (
			users	string
			err	error
		)

		for i := 0; users != ""; i ++ {
			/* 
				s.ne.Request(fd, data)  // it will block until data is return, OnRecv will not be trigged
			*/
			users, err = s.http_request(name, "/users/get?page=" + strconv.FormatInt(i, 10))
			if err != nil {
				s.log_error(err.Error())
			}
			fmt.Println("new users:" + users)
		}

		for i := 0; groups != ""; i ++ {
			groups, err = s.http_request(name, "/groups/get?page=" + strconv.FormatInt(i, 10))
			if err != nil {
				s.log_error(err.Error())
			}
			fmt.Println("new users:" + groups)
		}

	}

	// 开启message服务, 接受用户的连接请求，进行消息服务
	s.open_message_port = true

	for {
		// do nothing, just keep loop here ...
	}

}

