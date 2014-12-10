/* a mini web event-driven frame work
*
*	by awen, 2014.12.08
*/

package "netevent"
import	"net"
import	"errors"
import	"time"

///////////////////////////////////////
// NetEvent 
//////////////////////////////////////
type NetEvent struct {

	/* 
	* 连接池(包含所有连接)
	*/
	conns	map[uint32] *NetConn

	/*
	* 端口池(包含所有监听的端口)
	*/
	ports	map[uint] *NetPort



	/*
	* OnShutdown	回调函数，当 NetEvent收到 USR2 信号退出时，会触发该函数
	*/
	OnShutdown	func()error


	/*
	* 最大连接数, 默认是 65535
	*/
	MaxConnNum	uint32

	/* 
	* 最新fd ( 0 ~ MaxConnNum )
	*/
	new_fd	uint32

	/* 
	* 请求一个新的NetConn
	*/
	new_conn	chan bool

	/* 
	* 请求fd对应的NetConn
	*/
	fd_conn	chan	uint32

	/*
	* 返回 *NetConn
	*/
	return_conn	chan	*NetConn

	/* 
	* fd 已被用完
	*/
	full_fd	chan bool

	/* 
	* 请求一个新的NetPort
	*/
	new_port	chan	uint

	/* 
	* 返回port
	*/
	return_port	chan	uint
}

/* 初始化一个NetEvent 实例
*	
* 		func Init() (*NetEvent)
*	
*/

func Init() *NetEvent {
	ne := &NetEvent{}
	ne.MaxConnNum = 65535
	ne.new_fd = 0
	ne.new_conn = make(chan bool)
	ne.fd_conn = make(chan uint32)
	ne.return_conn = make(chan *NetConn)
	ne.full_fd = make(chan uint32)
	ne.new_port = make(chan uint)
	ne.return_port = make(chan uint)
	/* 启动 创建新连接 服务 */
	go func() {
		var	(
			found	bool	=	false
			fd	uint32	=	0
			new_conn	*NetConn = nil
		)
		for {
			select {
			case	_ = <-ne.new_conn :
				//找到可用的NetConn 或 生成新的 NetConn
				last_fd := ne.new_fd
				found = false
				for !found {
					// 生成新的 NetConn
					if c, exists := ne.conns[ne.new_fd]; !exists {
						ne.conns[ne.new_fd] = &NetConn {
							state :	NC_STATE_INIT,
							is_watched : false,
							sent_bytes : 0,
							recv_bytes : 0,
							sent_pack_num : 0,
							recv_pack_num : 0,
						}
						new_conn = ne.conns[ne.new_fd]
						found = true
					// 重用closed状态下的NetConn
					} else if c.to_state( NC_STATE_INIT ) {
						new_conn = c
						found = true
					}
					ne.new_fd ++

					//rewind
					if ne.new_fd >= ne.MaxConnNum {
						ne.new_fd = 0
					}

					//conns pool is full
					if ne.new_fd == last_fd {
						break;
					}
				}
				if found {
					ne.return_conn <- new_conn
				} else {
					ne.full_fd <- true
				}


				//根据fd 获取 NetConn
			case fd = <-ne.fd_conn :
				if c, exists := ne.conns[fd]; !exists {
					ne.conns[fd] = &NetConn {
						state :	NC_STATE_CLOSED,
						is_watched : false,
						sent_bytes : 0,
						recv_bytes : 0,
						sent_pack_num : 0,
						recv_pack_num : 0,
					}
				}
				ne.return_conn <- ne.conns[fd]


			case port := <-ne.new_port :
				if p, exists := ne.ports[port]; !exists {
					ne.ports[port] = &NetPort{
						ne : ne,
						port : port,
						shutdown : false,
					}
				}
				ne.return_port <- port

			default:
				/* do nothing */
			}
		}
	}()
	return ne
}


/* 启动一个NetEvent实例
*	
*		func (ne *NetEvent) Run() (err error)
*	
*/

func (ne *NetEvent) Run() error {
}


/* 获取一个端口实例(NetPort), 如果该实例还不存在，会新建一个
*	
*		func (ne *NetEvent) Port(uint port_num) (port *NetPort)
*	
*/
func (ne *NetEvent) Port(port_num uint) (*NetPort) {
	ne.new_port <- port_num
	port_num <- ne.return_port
	return ne.ports[port_num]
}


/* 关闭一个端口实例(NetPort), 会触发 NetPort.OnShutdown
*
*		func (ne *NetEvent) Shutdown(port_num uint) 
*/
func (ne *NetEvent) Shutdown(port_num uint) {
	if p, exists := ne.ports[port_num]; exists {
		if p != nil {
			p.Shutdown()
		}
		ne.ports[port_num] = nil
	}
}


/* 获取一个新的 *NetConn
*
*		func (ne *NetEvent) NewConn() (*NetConn, err error)
*/
func (ne *NetEvent) NewConn() (*NetConn, error) {
	ne.new_conn <- true
	for {
		select {
		case c := <-ne.return_conn:
			return c, nil

		case _ = <-ne.full_fd:
			return 0, errors.New("fd full")
		}
	}
}


/* 获取一个连接实例(NetConn), 跟 func (ne *NetEvent) Port() 一样,如果该实例还不存在，会新建一个
*  fd 不能超过最大连接数 MaxConnNum
*
*		func (ne *NetEvent) Conn(uint32 fd) (conn *NetConn)
*/
func (ne *NetEvent) Conn(fd uint32) (*NetConn) {
	if fd >= ne.MaxConnNum {
		return nil
	}
	ne.fd_conn <- fd
	nc := <-ne.return_conn
	return nc
}


/* 关闭一个连接实例(NetConn), 会触发 NetConn.OnClose
*
*		func (ne *NetEvent) Close(port_num uint32) (error) 
*/
func (ne *NetEvent) Close(fd uint32) (error) {
	if fd >= ne.MaxConnNum {
		return errors.New("bad fd")
	}
	return ne.Conn(fd).Close()
}





////////////////////////////////////////
//	NetPort
////////////////////////////////////////
type	NetPort	struct	{

	/*
	* OnStart	回调函数, 当 Listen() 被调用时，会触发该函数
	*/
	OnStart	func()error


	/*
	* OnShutdown	回调函数，当 Shutdown() 被调用时，会触发该函数
	*/
	OnShutdown	func()error


	/*
	* OnConn	回调函数, 一个端口被监听之后，当一个连接被accept之后，会调用该函数
	*/
	OnConn	func(fd uint32)error


	/*
	* OnRecv	回调函数，一个连接被成功accept之后，该函数会作为默认的OnRecv 被赋值给对应的 NetConn.OnRecv 
	*/
	OnRecv	func(fd uint32, data []buff)error


	/*
	* OnClose	回调函数，一个连接被成功accept之后，该函数会作为默认的OnClose 被赋值给对应的 NetConn.OnClose
	*/
	OnClose	func(fd uint32)error


	shutdown	bool

	/* 
	* NetEvent instance 
	*/
	ne	*NetEvent

	/* 
	* port to be listen 
	*/
	port	uint

}

/* 关闭端口
*
*/
func	(np *NetPort) Shutdown() {
	np.shutdown = true
}


/* 监听端口
*		func (np *NetPort) Listen(host string) (error)
*/
func	(np *NetPort) Listen(host string) error {
	var (
		err	error
		new_conn	uint32
		conn net.Conn
		tcp_addr	*net.TcpAddr
		listener	*net.TCPListener
	)

	np.OnStart()

	tcp_addr, err = net.ResolveTCPAddr("tcp", host_port)
	if err != nil {
		return err
	}

	listener, err = net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		return err
	}

	for !np.shutdown {
		listener.SetDeadLine(time.Now().Add(-1 * time.Second))
		//创建新连接
		if conn, err = listener.Accept() ; err != nil {
			//端口accept出现异常，关闭该端口
			if !isTimeout(err) {
				np.shutdown = true
			}
			continue
		}

		new_conn, err = np.ne.NewConn()
		if err != nil {
			//fd 池已满，拒绝连接
			conn.Close()
			continue
		}
		new_conn.sock = conn
		new_conn.from_port = np.port
		new_conn.OnConn = np.OnConn
		new_conn.OnRecv = np.OnRecv
		new_conn.OnClose = np.OnClose
		new_conn.to_state(NC_STATE_CONNECTED)
		go new_conn.Watch()
	}

	np.OnShutdown()
}

/* 
* 判断一个error是否为 timeout error
*/
func	isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}


/* 停止监听端口 
*		func (np *NetPort) Shutdown()
*/
func	(np *NetPort) Shutdown() {
	np.shutdown = true
}




//////////////////////////////////////
// NetConn
/////////////////////////////////////
type	NetConn	struct	{
	/*
	* OnConn	回调函数
	*/
	OnConn	func(fd uint32)error


	/*
	* OnRecv	回调函数，当有数据进来的时候，会先启用 OnPackeEof 进行自动分包，分包完成并得到一个完整的数据包之后，会触发该函数并传递data过来
	*/
	OnRecv	func(fd uint32, pack []byte)error


	/*
	* OnClose	回调函数，当连接断开的时候，先清理完该连接所占用的资源，然后调用该函数
	*/
	OnClose	func(fd uint32)error


	/*
	* OnPackEof,	自动分包函数
	*/
	OnPackEof	func(stream	[]byte)(int, error)


	/* 是否已经被监听
	*/
	is_watched	bool


	/* net 连接
	*/
	sock	net.Conn

	state	int	//INIT, CONNECTED, CLOSED

	/*
	* 来自哪个端口
	*/
	from_port	uint

	sent_bytes	uint64		//发送字节数
	recv_bytes	uint64		//接收字节数
	sent_pack_num	uint64		//发送数据包(Send() 函数调用次数)
	recv_pack_num	uint64		//接收数据包(OnRecv() + Recv() 函数调用次数)

}

const	NC_STATE_CLOSED = 0
const	NC_STATE_INIT = 1
const	NC_STATE_CONNECTED = 2

/* 转换状态
*
*/
func	(nc *NetConn) to_state(state int) {
}

/* 连接远端服务器
*
*		func (nc *NetConn) Connect(host string, port uint) error 
*/
func	(nc *NetConn) Connect(host string, port uint) error {
	// 原子操作更新
	// 连接, 
			//如果成功，返回nil	state = ESTABLISHED
			//如果失败，返回 erorr, 并更新 state = CLOSED
}


/* 监听连接(启动一个routine监听该连接，当有数据到来/连接被断开时，该routine会调用相应的回调函数)
*
*		func (nc *NetConn) Watch() 
*/
func	(nc *NetConn) Watch() error {
	// 原子操作更新 is_watched = true
	// 启动线程监听该连接, 
			//如果成功，返回 nil
			//如果失败，返回 error, 并更新 is_watched = false
			////!!! 注意在从 nc 结构中读取变量时，先判断成员是否为nil, 然后赋值给局部变量
}


/* 获取连接信息（客户端ip, 客户端port, 服务端 host, 服务端 port, 已发送字节数，已接受字节数，已发送数据包，已接受数据包)
*
*		func (nc *NetConn) Info() (info map[string]string, err error)
*/
func	(nc *NetConn) Info() (map[string]string, error) {
	return nc.sock.Info()
}


/* 向该连接发送数据
*
*		func (nc *NetConn) Send(data []byte) (sent_bytes int, err error)
*/
func	(nc *NetConn) Send(data []byte) (int, error) {
	return nc.sock.Send(data)
}


/* 从该连接中读取数据(不会触发OnRecv)
*
*		func (nc *NetConn) Recv(buff []byte) (recv_bytes int, err error) 
*/
func	(nc *NetConn) Recv(buff []byte) (int, error) {
	// 如果连接已经被监听，返回error(“is watched")
	// 如果连接未被监听，
		//原子操作 is_watched = true
		//读取数据直到完成，
		//is_watched = false
		//返回数据
}


/* 关闭连接(不会触发OnClose)
*
*		func (nc *NetConn) Close() (error)
*/
func	(nc *NetConn) Close() error {
	// 如果is_watch = true, 异步通知监听线程结束
	// 全部 field 置为初始值
}

