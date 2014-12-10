/* netevent ,  a mini network event frame work
 *
 * by awen, 2014.11.28
 */
package	ne
import	"strconv"
import	"fmt"
import	"errors"


/*********************************************************
*						NetEvent						 *
**********************************************************/
type NetEvent struct {

	/* onstart call back 
	 *
	 * 当 Run() 被调用时，首先会触发 OnStart 回调函数, 然后将剩余未启动的端口监听线程起起来; 
	 * 用户也可以在 OnStart 函数中手动监听端口
	 * 
	 * 调用 Listen()函数不会触发 OnStart 回调函数
	 */
	OnStart	func()error


	/* onshutdown call back
	 *
	 * 当进程捕获到 USR2 信号时，该函数首先会被触发, 然后再进行NetEvent内部资源的销毁
	 */
	OnShutdown func()error


	/* server host */
	Host	string


	/* 端口监听 */
	ports	map[uint] *NetPort


	/* 连接 */
	conns	map[uint32] *NetConn

}


/* 获取一个 NetEvent 实例 */
func Init() *NetEvent {
	ne := &NetEvent{
		OnStart : nil,
		OnShutdown : nil,
		Host : "",
	}
	ne.ports = make( map[uint] *NetPort )
	ne.conns = make( map[uint32] *NetConn )
	return ne
}


/* 启动一 NetEvent 实例 
 *
 * 该函数首先触发 ne.OnStart 回调函数, 然后逐一监听 ne.ports 中未监听的端口
 */
func (ne *NetEvent) Run(){
	if ne.OnStart != nil {
		ne.OnStart()
	}

	var (
		p	uint,
		np	NetPort,
		err	error,
		i	int = 0,
		ports_num	int = 0,
		caller	func(np *NetPort),
	)
	caller = func(np *NetPort) {
		if err := np.listen(); err != nil {
			ne.LogError(err.Error())
			return
		}
	}

	ports_num = len(ne.ports)
	for p, np := range ne.ports {
		i ++
		np.HostPort = ne.Host + ":" + strconv.FormatUint64(uint64(p), 10)

		if i < ports_num {
			go caller(np)
		} else {
			caller(np)
		}
	}
}

/* 记录错误信息
*/
func (ne *NetEvent) LogError(err string) {
	fmt.Println(err)
}


/* 获取一个端口实例 
* (目前尚未做线程安全处理)
*/
func (ne *NetEvent) Port(port uint) (*NetPort, error) {
	new_p := &NetPort {
		OnStart	: nil,
		OnShutdown : nil,
		OnConn : nil,
		OnRecv : nil,
		OnClose : nil,
		HostPort : "",
	}
	if p, exists := ne.ports[port]; exists {
		return p, nil
	}
	ne.ports[port] = new_p
	return new_p, nil
}

/* 获取一个连接实例 */
func (ne *NetEvent) Conn(fd uint32) (*NetConn, error) {
	new_conn := &NetConn {
		OnRecv : nil,
		OnClose : nil,
		OnPackEof : nil,
	}
	if c, exists := ne.conns[fd]; exists {
		return c, nil
	}
	ne.conns[fd] = new_conn
	return new_conn, nil
}


///////////////////
// port functions 
///////////////////
/* 启动一个线程监听指定端口
* 除了 ne.Conn(fd).OnXxx 回调函数以外，不能用于其它回调函数
*
* @param	port	端口
* @param	async	是否异步, 如果为false, 则程序将阻塞在这里，直到程序收到 USR2 退出信号
*/
const	LISTEN_ASYNC = false
const	LISTEN_SYNC = true
func (ne *NetEvent) Listen(port int, async bool) error {
	if p, exists := ne.ports[port]; !exists {
		return p, errors.New("bad port")
	}
	caller := func(np *NetPort) error {
		if err := np.listen(); err != nil {
			ne.LogError(err.Error())
			return err
		}
		return nil
	}
	if (async == LISTEN_ASYNC) {
		go caller(p)
		return nil
	} else {
		return caller(p)
	}
}


/* 通知指定端口的监听线程 结束 并等待,直到监听线程成功结束
* 除了 ne.Conn(fd).OnXxx 回调函数以外，不能用于其它回调函数
*
* @param	port	端口
* @param	error
*/
func (ne *NetEvent) Shutdown(port int) error {
	if p, exists := ne.ports[port]; !exists {
		return p, errors.New("bad port")
	}
	p.shutdown <- true
	_ <- p.exit
	return nil
}


//////////////////////
// connect functions 
//////////////////////
/* 主动连接远端服务器, 必须是还未建立socket的
*
* @param	fd	
* @param	host, string
* @param	port, int
* @error	"1. fd不存在； 2. fd已经建立好socket; 3. 网络错误"
*/
func (ne *NetEvent) Connect(fd uint32, host string, port uint) error { return nil }

/* 开启一个线程，监听fd上的数据
*
* @param	fd
* @error	"1. OnRecv 未被设置; 2. 连接已经被监听; 3.对应的NetConn不存在"
*/
func (ne *NetEvent) Watch(fd uint32) error { return nil }


/* 主动连接远端服务器, 该函数功能与 NetEvent::Connect() 一样，唯一不同的是它会启动一个线程监听fd上的数据( 等价于执行： ` Connect(fd); Watch(fd); ` )
*
* @param	fd
* @param	host, string
* @param	port, int
* @error	"1. OnRecv 未设置; 2. 连接已经被监听; 3. 对应的NetConn不存在"
*/
func (ne *NetEvent) Dial(fd uint32, host string, port uint) error { return nil }

/* 

/* 取连接的详细信息（见 NetConnInfo )
*
* @param	fd, uint32
* @error	"1. 对应的NetConn不存在"
*/
func (ne *NetEvent) Info(fd uint32) error { return nil }

/* 
* func (ne *NetEvent) Send( fd uint32, data []byte) ( sent_bytes int, err error ) 
* 发送数据到指定连接
*
* @param	fd
* @param	data
* @error	"1. 对应的NetConn不存在; 2. 网络错误"
*/
func (ne *NetEvent) Send(fd uint32, data []byte) (int, error) { return 0, nil }

/*
* func (ne *NetEvent) Recv( fd uint32, buff []byte) (recv_bytes int, err error) 
* 从一个连接中读取数据(该函数必须在连接未被监听的情况下才有效)
*
* @param	fd
* @param	data
* @error	"1. 对应的NetConn不存在; 2. 网络错误; 3.该fd已经被监听"
*/
func (ne *NetEvent) Recv(fd uint32, data []byte) (int, error) { return 0, nil }

/* 关闭连接
* func (ne *NetEvent) Close( fd uint32 ) error 
*
* @param	fd
* @error	"1. 对应的NetConn不存在; 2. 网络错误; 3.该fd已经被监听"
*/
func (ne *NetEvent) Close(fd uint32) error { return nil }




/*********************************************************
*						NetPort							 *
**********************************************************/

type NetPort struct {

	/* 当 NetEvent.Listen(port) 被调用时，启动的新线程会首先触发该函数, 然后再进行监听工作
	*
	*/
	OnStart	func()error

	/* 当 NetEvent 退出的时候，会通知各个监听线程退出，监听线程退出之前会调用该函数
	*/
	OnShutdown	func()error


	/* 新连接创建后，会将该 OnConn 函数设置为新连接的 OnConn 回调函数
	*/
	OnConn	func(fd uint32)error

	/* 新连接创建后，会将该 OnRecv 函数设置为新连接的 OnRecv 回调函数
	*/
	OnRecv	func(fd uint32, pack []byte)error

	/* 新连接创建后，会将该 OnClose 函数设置为新连接的 OnClose 回调函数
	*/
	OnClose	func(fd uint32)error


	HostPort	string	//ip端口

	is_listening	bool	//是否在監聽

}

/* 監聽端口, 接受並創建連接 */
func (np *NetPort) listen() error { return nil }




/*********************************************************
*						NetConn							 *
**********************************************************/

type NetConn struct {

	/* 这里的 OnRecv, OnClose 被调用之后，不会再调 NetPort.OnRecv 和 NetPort.OnClose（事实上NetPort.OnRecv，NetPort.OnClose只是作为临时变量，是永远不会被调用的)
	*/
	OnRecv	func(fd uint32, pack []byte)error

	OnClose	func(fd uint32)error

	/* 自动分包函数  func(stream []byte)(pack_len int, err error)
	*
	* 该函数需要确保：
	*	1. stream的头部(stream[0 : ])即包头
	*	2. 从stream的头部算起, 能在stream中找到完整的数据包，并返回包的长度
	*	3. 否则返回error
	*/
	OnPackEof	func(stream []byte) (int, error)

}


/* 连接远程服务，并启动一个监听线程监听数据
 *
 */
func (nc *NetConn) Connect(host string, port int) error { return nil }


/* 连接远程服务，同时监听该连接，直到连接被关闭或者程序退出
*
*/
func (nc *NetConn) Dial(host string, port int) error { return nil }

/* 向该连接发送数据 
*
*/
func (nc *NetConn) Send(data []byte) (int, error) { return 0, nil }


/* 关闭该连接
*
*/
func (nc *NetConn) Close() error { return nil }


type NetConnInfo struct {
	SysFd	uint32
	SysProcId	uint32
	RemoteHost	string
	RemotePort	uint32
	LocalHost	string
	LocalPort	uint32
	SentBytes	uint64
	RecvBytes	uint64
	SentPacks	uint64
	RecvPacks	uint64
}


/* 获取connect信息 
*/
func (nc *NetConn) Info() (*NetConnInfo, error) { return nil, nil }

