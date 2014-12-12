/* a mini web event-driven frame work
*
*	by awen, 2014.12.08
*/
package main
import	"net"
import	"errors"
import	"time"
import	"sync/atomic"
func main() {
	ne := &NetEvent{}
	ne.Port(9999).Listen("127.0.0.1")
}

///////////////////////////////////////
// NetEvent 
//////////////////////////////////////
type NetEvent struct {
	state	uint32	//NE_STATE_SHUTDOWN, NE_STATE_RUNNING

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
	* 每个连接的 buff 大小
	*/
	ConnBuffSize	uint64

	/* 
	* 最新fd ( 0 ~ MaxConnNum )
	*/
	new_fd	uint32

	/* 
	* 请求一个新的NetConn
	*/
	new_conn	chan uint32	//新NetConn的初始状态

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

const	NE_STATE_SHUTDOWN = 0
const	NE_STATE_RUNNING = 1

/* 初始化一个NetEvent 实例
*	
* 		func Init() (*NetEvent)
*	
*/

func (ne *NetEvent) try_to_init() {
	ne_state := ne.state
	if ne_state != NE_STATE_SHUTDOWN {
		return
	}
	if !atomic.CompareAndSwapUint32(&ne.state, ne_state, NE_STATE_RUNNING) {
		return
	}

	ne.MaxConnNum = 65535
	ne.ConnBuffSize = 1024
	ne.new_fd = 0
	ne.new_conn = make(chan uint32)
	ne.fd_conn = make(chan uint32)
	ne.return_conn = make(chan *NetConn)
	ne.full_fd = make(chan bool)
	ne.new_port = make(chan uint)
	ne.return_port = make(chan uint)
	/* 启动 创建新连接 服务 */
	go func() {
		var	(
			found	bool	=	false
			fd	uint32	=	0
			new_conn	*NetConn = nil
			new_conn_st	uint32	=	0
			port	uint	=	0
		)
		for {
			select {
			case	new_conn_st = <-ne.new_conn :
				//找到可用的NetConn 或 生成新的 NetConn
				last_fd := ne.new_fd
				found = false
				for !found {
					// 生成新的 NetConn
					if c, exists := ne.conns[ne.new_fd]; !exists {
						ne.conns[ne.new_fd] = &NetConn {
							state :	new_conn_st,
							buff_size : ne.ConnBuffSize,
							fd : ne.new_fd,
							buff : make([]byte, ne.ConnBuffSize),
						}
						new_conn = ne.conns[ne.new_fd]
						found = true
					// 重用closed状态下的NetConn
					} else if c.to_state( new_conn_st ) {
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
						buff_len : 0,
						buff_head : 0,
						buff_size : ne.ConnBuffSize,
						sent_bytes : 0,
						recv_bytes : 0,
						sent_pack_num : 0,
						recv_pack_num : 0,
						from_port : 0,
						is_closing : false,
						fd : fd,
						buff : make([]byte, ne.ConnBuffSize),
					}
				}
				ne.return_conn <- ne.conns[fd]


			case port = <-ne.new_port :
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


/* 获取一个端口实例(NetPort), 如果该实例还不存在，会新建一个
*	
*		func (ne *NetEvent) Port(uint port_num) (port *NetPort)
*	
*/
func (ne *NetEvent) Port(port_num uint) (*NetPort) {
	ne.try_to_init()
	ne.new_port <- port_num
	port_num <- ne.return_port
	return ne.ports[port_num]
}


/* 关闭一个端口实例(NetPort), 会触发 NetPort.OnShutdown
*
*		func (ne *NetEvent) Shutdown(port_num uint) 
*/
func (ne *NetEvent) Shutdown(port_num uint) {
	ne.try_to_init()
	if p, exists := ne.ports[port_num]; exists {
		if p != nil {
			p.Shutdown()
		}
	}
}


/* 获取一个待监听的 *NetConn
*
*/
func (ne *NetEvent) accept_conn() (*NetConn, error) {
	ne.new_conn <- NC_STATE_TOWATCH
	for {
		select {
		case c := <-ne.return_conn:
			return c, nil

		case _ = <-ne.full_fd:
			return 0, errors.New("fd full")
		}
	}
}


/* 获取一个新的 *NetConn
*
*		func (ne *NetEvent) NewConn() (*NetConn, err error)
*/
func (ne *NetEvent) NewConn() (*NetConn, error) {
	ne.try_to_init()
	ne.new_conn <- NC_STATE_INIT
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
	ne.try_to_init()
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
	ne.try_to_init()
	if fd >= ne.MaxConnNum {
		return errors.New("bad fd")
	}
	if c, e := ne.conns[fd]; e && c != nil {
		return c.Close()
	}
	return errors.New("empty fd")
}


/* 监听端口
*
*/
func (ne *NetEvent) listen(port uint) error {
	if p, e := ne.ports[port]; (e && p != nil) {
		err := p.listen(port)

		//关闭所有来自该端口的连接
		for fd, c := range ne.conns {
			if c != nil && c.from_port == port {
				c.Close()
			}
		}
		return err
	}
	return errors.New("empty port")
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

	/*
	* state: NP_STATE_SHUTDOWN, NP_STATE_LISTING, NP_STATE_INIT
	*/
	state	uint32


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
const	NP_STATE_SHUTDOWN = 0
const	NP_STATE_LISTENING = 1
const	NP_STATE_INIT = 2


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
	return np.ne.listen(host)
}

func	(np *NetPort) listen(host string) error {
	var (
		err	error
		new_conn	uint32
		conn net.Conn
		tcp_addr	*net.TcpAddr
		listener	*net.TCPListener
		host_port	string
		np_state	uint32
	)
	np_state = np.state
	if np_state != NP_STATE_SHUTDOWN {
		return errors.New("wrong port state")
	}
	if !atomic.CompareAndSwapUint32(&np.state, np_state, NP_STATE_INIT) {
		return errors.New("failed to init port")
	}


	host_port = host + ":" + strconv.FormatUint(np.port)
	tcp_addr, err = net.ResolveTCPAddr("tcp", host_port)
	if err != nil {
		np.state = NP_STATE_SHUTDOWN
		return err
	}

	listener, err = net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		np.state = NP_STATE_SHUTDOWN
		return err
	}

	np.state = NP_STATE_LISTENING
	if np.OnStart != nil {
		np.OnStart()
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

		new_conn, err = np.ne.accept_conn()
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
		go new_conn.Watch()
	}
	np.state = NP_STATE_SHUTDOWN
	if np.OnShutdown != nil {
		np.OnShutdown()
	}
	return nil
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



	/* net 连接
	*/
	sock	net.Conn

	/* 
	* 自动分包相关
	*/
	buff_head	int
	buff_len	int

	state	uint32	//INIT, CONNECTED, CLOSED

	/*
	* 来自哪个端口
	*/
	from_port	uint

	/* 
	* 来自NetEvent的fd
	*/
	fd	uint32

	/*
	* read buffer
	*/
	buff	[]byte

	sent_bytes	uint64		//发送字节数
	recv_bytes	uint64		//接收字节数
	sent_pack_num	uint64		//发送数据包(Send() 函数调用次数)
	recv_pack_num	uint64		//接收数据包(OnRecv() + Recv() 函数调用次数)

	/* 
	* 是否正在关闭中
	*/
	is_closing	bool

}

/* 共可容纳10个不同的状态，见to_state()函数*/
const	NC_STATE_CLOSED = 0
// 未监听的连接状态
const	NC_STATE_INIT = 1
const	NC_STATE_CONNECTED = 2
// 监听中的状态
const	NC_STATE_TOWATCH = 3
const	NC_STATE_WATCHING = 4
// Recv阻塞状态
const	NC_STATE_RECVING = 5


func	(nc *NetConn) state_to_init(){
	nc.sent_bytes = 0
	nc.recv_bytes = 0
	nc.sent_pack_num = 0
	nc.recv_pack_num = 0
	nc.buff_len = 0
	nc.buff_head = 0
	nc.from_port = 0
	nc.is_closing = false
}

func	(nc *NetConn) state_to_protected() {
	nc.state_to_init()
}

func	(nc *NetConn) state_to_closed() {
}

/* 转换状态
*
*/
func	(nc *NetConn) to_state(state uint32) bool {
	old_state := nc.state
	swi := old_state * 10 + state
	closed_to_init := NC_STATE_CLOSED * 10 + NC_STATE_INIT
	init_to_connected := NC_STATE_INIT * 10 + NC_STATE_CONNECTED
	connected_to_closed := NC_STATE_CONNECTED * 10 + NC_STATE_CLOSED

	closed_to_towatch := NC_STATE_CLOSED * 10 + NC_STATE_TOWATCH
	towatch_to_watching := NC_STATE_TOWATCH * 10 + NC_STATE_WATCHING
	watching_to_closed := NC_STATE_WATCHING * 10 + NC_STATE_CLOSED

	connected_to_recving := NC_STATE_CONNECTED * 10 + NC_STATE_RECVING
	recving_to_connected := NC_STATE_RECVING * 10 + NC_STATE_CONNECTED


	var state_func func() = nil

	switch swi {

	////// 未监听的连接状态转换
	case closed_to_init :
		state_func = nc.state_to_init
	case init_to_connected :
		/* do nothing */
	case connected_to_closed :
		state_func = nc.state_to_closed

	////// 监听下的连接状态转换
	case closed_to_towatch :
	case towatch_to_watching :
	case watching_to_close :
		/* do nothing */

	////// 从未监听到监听状态的转换
	case connected_to_watching :

	/////// recving阻塞状态 <-> connected状态
	case recving_to_connected :
	case connected_to_recving :

	default:
		return false
	}

	if atomic.CompareAndSwapUint32( &nc.state, old_state, state) {
		if state_func != nil {
			state_func()
		}
		return true
	} else {
		return false
	}
}

/* 连接远端服务器
*
*		func (nc *NetConn) Connect(host string, port uint) error 
*/
func	(nc *NetConn) Connect(host string, port uint) error {
	if (nc.state == NC_STATE_CLOSED && nc.to_state(NC_STATE_INIT)) {
		var (
			conn	net.Conn
			err	error
		)
		conn, err = net.Dial(host + ":" + strconv.FormatUint(port, 10))
		if err != nil {
			nc.to_state(NC_STATE_CONNECTED)
			nc.to_state(NC_STATE_CLOSED)
			return err
		}

		nc.sock = conn
		nc.to_state(NC_STATE_CONNECTED)
		return nil

	} else {
		return errors.New("conn in used")
	}
}


/* 监听连接，当有数据到来/连接被断开时，该函数会调用相应的回调函数
*
*		func (nc *NetConn) Watch() 
*/
func	(nc *NetConn) Watch() error {
	if !nc.to_state(NC_STATE_WATCHING) {
		return 0, errors.New("conn wrong state")
	}
	if nc.OnConn != nil {
		nc.OnConn(nc.fd)
	}
	for !nc.is_closing {
		bytes, err := nc.read_pack()
		if err == nil {
			if nc.OnRecv != nil {
				nc.OnRecv(nc.fd, bytes)
			}
			nc.recv_pack_num += 1
		} else {
			if len(bytes) > 0 {
				if nc.OnRecv != nil {
					nc.OnRecv(nc.fd, bytes)
				}
				nc.recv_pack_num += 1
			}
		}
	}
	nc.to_state(NC_STATE_CLOSED)
	err := nc.sock.Close()
	if nc.OnClose != nil {
		nc.OnClose(nc.fd)
	}
	return err
}

/* 读取数据包
*
*/
func	(nc *NetConn) read_pack() ([]byte, error) {
	var (
		err	error	=	nil
		buff_head	int	=	0
		buff_len	int	=	0
		plen	int		=	0
		read_len	int
		conn	*NetConn
	)

	// 如果没有自动分包规则，直接将数据传递给OnRecv
	if nc.OnPackEof == nil {
		read_len, err = nc.sock.Read(nc.buff[nc.buff_head + nc.buff_len : ])
		nc.recv_bytes += read_len
		read_len += nc.buff_len
		buff_head = nc.buff_head
		nc.buff_len = 0
		nc.buff_head = 0
		if err == nil {
			return nc.buff[buff_head : read_len], nil
		} else {
			return nc.buff[buff_head : read_len], err
		}

	// 如果有自动分包规则，过滤分包规则并传递给OnRecv
	} else {

		for {
			buff_head = nc.buff_head
			buff_len = nc.buff_len
			if buff_len > 0 {
				plen, err = nc.OnPackEof(nc.buff[nc.buff_head : nc.buff_head + nc.buff_len])

				// 如果现有的buff中找到了数据包，直接返回
				if err == nil && plen > 0 {
					nc.buff_head += plen
					nc.buff_len -= plen
					return nc.buff[buff_head : plen + buff_head], nil
				}
			}

			// 从socket中读数据进buffer
			read_len, err = nc.sock.Read(nc.buff[nc.buff_head + nc.buff_len : ])
			nc.recv_bytes += read_len
			nc.buff_len += read_len

			if err != nil {
				buff_len = nc.buff_len
				nc.buff_head = 0
				nc.buff_len = 0
				return nc.buff[buff_head : buff_head + buff_len], err
			}

			//读到buff已经满了，仍然未找到package
			if nc.buff_len + nc.buff_head >= nc.buff_size {

				//整个buff已满
				if nc.buff_head == 0 {
					return nc.buff[0 : ], errors.New("no package")

				//buff头部还有空闲位置，将内容整体往头部偏移长度 buff_head
				} else {
					copy(nc.buff[0 : nc.buff_len], conn.buff[ nc.buff_head : nc.buff_head + nc.buff_len])
					nc.buff_head = 0
				}
			}
		}
	}
}


/* 获取连接信息（客户端ip, 客户端port, 服务端 host, 服务端 port, 已发送字节数，已接受字节数，已发送数据包，已接受数据包)
*
*		func (nc *NetConn) Info() (info map[string]string, err error)
*/
func	(nc *NetConn) Info() (map[string]string, error) {
	if nc.state == NC_STATE_CLOSED {
		return nil, errors.New("unused conn")
	}
	info := make(map[string]string)
	info["sent_bytes"] = strconv.FormatUint64(nc.sent_bytes, 10)
	info["recv_bytes"] = strconv.FormatUint64(nc.recv_bytes, 10)
	info["sent_pack_num"] = strconv.FormatUint64(nc.sent_pack_num, 10)
	info["recv_pack_num"] = strconv.FormatUint64(nc.recv_pack_num, 10)
	info["peer_addr"] = nc.sock.RemoteAddr().String()
	info["local_addr"] = nc.sock.LocalAddr().String()
	return info, nil
}


/* 向该连接发送数据
*
*		func (nc *NetConn) Send(data []byte) (sent_bytes int, err error)
*/
func	(nc *NetConn) Send(data []byte) (int, error) {
	if nc.state != NC_STATE_CONNECTED || nc.state != NC_STATE_WATCHING || nc.state == NC_STATE_WATCHING && nc.is_closing {
		return 0, errors.New("conn wrong state")
	}
	sent_bytes, err := nc.sock.Send(data)
	nc.sent_bytes += sent_bytes
	nc.sent_pack_num += 1
	return sent_bytes, err
}


/* 从该连接中读取数据(不会触发OnRecv)
*
*		func (nc *NetConn) Recv(buff []byte) (recv_bytes int, err error) 
*/
func	(nc *NetConn) Recv(buff []byte) (int, error) {
	if !nc.to_state(NC_STATE_RECVING) {
		return 0, errors.New("conn wrong state")
	}
	buff_len, err := nc.sock.Read(buff)
	nc.recv_bytes += buff_len
	nc.recv_pack_num += 1
	nc.to_state(NC_STATE_CONNECTED)
	return buff_len, err
}


/* 关闭连接(不会触发OnClose)
*
*		func (nc *NetConn) Close() (error)
*/
func	(nc *NetConn) Close() error {
	if nc.state == NC_STATE_TOWATCH || nc.state == NC_STATE_WATCHING {
		nc.is_closing = true
		return nil
	} else {
		if nc.to_state(NC_STATE_CLOSED) {
			return nc.sock.Close()
		}
		return errors.New("failed to Close conn")
	}
}

