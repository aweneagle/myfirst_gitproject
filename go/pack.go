/* a mini web event-driven frame work
*
*	by awen, 2014.12.08
*/

package "netevent"
import	"net"

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
	* OnStart 回调函数，当 Run() 被调用时，会触发该函数
	*/
	OnStart	func()error


	/*
	* OnShutdown	回调函数，当 NetEvent收到 USR2 信号退出时，会触发该函数
	*/
	OnShutdown	func()error

}

/* 初始化一个NetEvent 实例
*	
* 		func Init() (*NetEvent)
*	
*/

func Init() *NetEvent {
	return &NetEvent{}
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
*		func (ne *NetEvent) Port(uint port_num ) (port *NetPort)
*	
*/
func (ne *NetEvent) Port(port_num uint) (*NetPort) {
}


/* 关闭一个端口实例(NetPort), 会触发 NetPort.OnShutdown
*
*		func (ne *NetEvent) Shutdown(port_num uint) (error) 
*/
func (ne *NetEvent) Shutdown(port_num uint) (error) {
}



/* 获取一个连接实例(NetConn), 跟 func (ne *NetEvent) Port() 一样,如果该实例还不存在，会新建一个
*
*		func (ne *NetEvent) Conn(uint32 fd) (conn *NetConn)
*/
func (ne *NetEvent) Conn(fd	uint32) (*NetConn) {
}


/* 关闭一个连接实例(NetConn), 会触发 NetConn.OnClose
*
*		func (ne *NetEvent) Close(port_num uint32) (error) 
*/
func (ne *NetEvent) Close(fd uint32) (error) {
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

}


/* 监听端口
*		func (np *NetPort) Listen() (error)
*/
func	(np *NetPort) Listen() error {
}


/* 停止监听端口, 不会触发 NetPort.OnShutdown
*		func (np *NetPort) Shutdown() (error)
*/
func	(np *NetPort) Shutdown() error {
	//异步回调
}




//////////////////////////////////////
// NetConn
/////////////////////////////////////
type	NetConn	struct	{

	/*
	* OnRecv	回调函数，当有数据进来的时候，会先启用 OnPackeEof 进行自动分包，分包完成并得到一个完整的数据包之后，会触发该函数并传递data过来
	*/
	OnRecv	func(pack []byte)error


	/*
	* OnClose	回调函数，当连接断开的时候，先清理完该连接所占用的资源，然后调用该函数
	*/
	OnClose	func()error


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

	is_conn	bool

}


/* 连接远端服务器
*
*		func (nc *NetConn) Connect(host string, port uint) error 
*/
func	(nc *NetConn) Connect(host string, port uint) error {
	// 原子操作更新 is_conn = true
	// 连接, 
			//如果成功，返回nil
			//如果失败，返回 erorr, 并更新 is_conn = false
}


/* 监听连接(启动一个routine监听该连接，当有数据到来/连接被断开时，该routine会调用相应的回调函数)
*
*		func (nc *NetConn) Watch() 
*/
func	(nc *NetConn) Watch() error {
	// 原子操作更新 is_watched = true
	// 启动线程监听该连接, 
			//如果成功，返回 nil
			//如果失败，返回 error, 并更新 is_conn = false
			////!!! 注意在从 nc 结构中读取变量时，先判断成员是否为nil, 然后赋值给局部变量
}


/* 获取连接信息（客户端ip, 客户端port, 服务端 host, 服务端 port, 已发送字节数，已接受字节数，已发送数据包，已接受数据包)
*
*		func (nc *NetConn) Info() (info map[string]string, err error)
*/
func	(nc *NetConn) Info() ( map[string]string, error) {
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

