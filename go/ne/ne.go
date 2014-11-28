package	ne
/* netevent ,  a mini network event frame work
 *
 * by awen, 2014.11.28
 */


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
	ports	map[int] NetPort


	/* 连接 */
	conns	map[uint32] NetConn

}


/* 获取一个 NetEvent 实例 */
func Init() *NetEvent {	return nil }


/* 启动一 NetEvent 实例 
 *
 * 该函数首先触发 ne.OnStart 回调函数, 然后逐一监听 ne.ports 中未监听的端口
 */
func (ne *NetEvent) Run(){}


/* 获取一个端口实例 */
func (ne *NetEvent) Port(port int) (*NetPort, error) { return nil, nil }


/* 获取一个连接实例 */
func (ne *NetEvent) Conn(fd uint32) (*NetConn, error) { return nil, nil }


/* 启动一个线程监听指定端口
*
* @param	port	端口
* @param	is_sync	是否异步, 如果为false, 则程序将阻塞在这里，直到程序收到 USR2 退出信号
*/
func (ne *NetEvent) Listen(port int, is_sync bool) error { return nil }






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

