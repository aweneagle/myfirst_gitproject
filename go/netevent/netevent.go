/*
 * net event driver
 *
 * 这是个微型网络事件框架，通过注册事件监听函数的方式，来简化网络应用开发的工作
 *
 * 使用示例:
 *		//给一个端口绑定一个驱动实例
 *		ne,_ := netevent.Listen("*:8888")
 *
 *		//注册handler， 监听该端口的"新增连接"事件
 *		ne.OnConnect	= func (fd uint32) {  
 *		//添加自动分包规则
 *			ne.SetPackEof(fd, netevent.PackEof { Eof:"_*_*_" } );
 *		}
 *
 *		//注册handler， 监听该端口的"数据包到达"事件
 *		ne.OnRecv		= func (pack []byte, fd uint32) {  
 *			//发送数据包
 *			ne.Send(fd, pack);
 *		}
 * 
 *		//注册handler,	监听该端口的"连接关闭"事件
 *		ne.OnClose		= func (fd uint32) {  
 *			//清理自动分包规则
 *			ne.CleanPackEof(fd);
 *		}
 *
 */

package	netevent
import	"net"
import	"strings"
import	"strconv"
import	"time"
import	"os"
import	"errors"
import	"io"

type net_event_driver struct{

	/* 回调函数：返回的error只用于net_event_driver的error_log*/
	OnRecv	func(uint32, []byte) error
	OnClose	func(uint32) error
	OnConn	func(uint32) error

	/* 
	ErrorLog, DebugLog 文件路径要求:
	
	log 文件格式支持按日期分隔: %Y为年份, %m为月份, %d为日期, %H为小时, %i 为分钟,  最小粒度为分钟
	例如：
		~/file/path.%Y%m%d.log	将	分隔日志为：
			~/file/path.20141107.log
			~/file/path.20141108.log
			~/file/path.20141109.log
			......
	*/

	/*
	ErrorLog:

	如果该文件路径为空，错误将输出到stderr；否则输出到该文件中
	错误日志的格式固定为：CHAT|2014-11-01 11:11:11|ERROR|$msg,  $msg 为错误消息内容

	默认值为 ""
	*/

	ErrorLog	string


	/*
	DebugLog:

	debug内容：每个连接的 OnConn, OnClose 事件
	debug日志的格式：CHAT|2014-11-01 11:11:11|DEBUG|$fd|(recv|conn|close)|(read_bytes:$read_bytes)|(pack_bytes:$pack_bytes)
		这里 (recv|conn|close) 指该字段为括号内的三个标识之一,  
		$fd 为事件所在的连接, 
		$read_bytes,$pack_bytes 只有在recv事件时才记录，$read_num 为连接在本次事件中从socket中读取出来的全部字节数, 
		$pack_bytes 为连接在本次事件中返回的包的字节数

	默认值为 ""
	*/
	DebugLog	string	//当debug = true时，如果该文件路径为空，日志内容将输出到stdout；否则输出到该文件中



	/* Debug:	

	是否开启dbeug模式

	默认为 true 
	*/
	Debug		bool


	/*
	ConnBuffSize:

	单个连接所用到的reading buffer大小; 当连接启用自动分包(见method  SetPackEof() )时，该值必须比最大的数据包大，超过该大小的数据包无法完整组装, 将不自动分包，直接把数据传递给
	OnRecv() 回调函数, 这时应用层需要自行处理分包问题

	默认值为  8K
	*/
	ConnBuffSize	int


	/* 
	IdleTime:

	一个连接闲置的最长时间, 它意味着一个连接超过 IdelTime 之后仍没有收到数据(或发送数据)，该连接将会被服务器自动关闭, 如果设置为0，则不进行自动清理

	默认值为 0, 单位为 second
	*/
	IdleTime	int


	/*
	MaxConnNum:

	最大连接数，这个值决定了fd的范围，fd的范围是： [0,MaxConnNum)

	默认值为 65535
	*/
	MaxConnNum	uint32





	host_port	string	//监听的端口

	conns	map[uint32] *connect

	last_fd	uint32//上一次被使用的fd

}

const	PACK_EOF_MAX_NUM	=	6	//每个连接所能拥有的最大的自动分包规则数量
const	CONN_RUNNING = 1
const	CONN_CLOSED = 2
type connect struct {
	sock	net.Conn
	buff	[]byte

	is_shutup	bool
	is_shutdown	bool

	pack_eof	[PACK_EOF_MAX_NUM] i_net_pack_eof	//自动分包规则
	pack_eof_num	uint8	//自动分包规则数量
	last_op_time	int64	//上一次操作时间(send or recv)

	state	uint8	//connect state:  ST_CONN_INI, ST_CONN_RUNNING, ST_CONN_TO_CLOSE
}


const CFG_MAX_CONN_NUM = 65535

/* 生成一个 event driver, 绑定在指定端口上
 *
 * @param	host_port, ip和端口用":"隔开, 格式:"*:8888"
 */
func Init (host_port string) *net_event_driver {
	ne := new(net_event_driver)

	ne.host_port = host_port
	ne.MaxConnNum = CFG_MAX_CONN_NUM
	ne.conns = make(map[uint32]*connect, ne.MaxConnNum)
	ne.ConnBuffSize = 8 * 1024
	ne.Debug = true
	ne.DebugLog = ""
	ne.ErrorLog = ""
	ne.IdleTime = 0
	ne.last_fd = 0

	ne.OnClose = nil
	ne.OnConn = nil
	ne.OnRecv = nil

	return ne
}


const ERR_BAD_FD = "bad fd"
const ERR_CONN_BAD_STATE = "connect bad state"


/* 启动 event driver 
 *
 */
func (ne *net_event_driver) Start() error {
	//如果最大连接数被用户设置过，则需要重新申请连接池内存
	if ne.MaxConnNum != CFG_MAX_CONN_NUM {
		ne.conns = make(map[uint32]*connect, ne.MaxConnNum)
	}

	tcp_addr, addr_err := net.ResolveTCPAddr("tcp", ne.host_port)
	if addr_err != nil {
		ne.log_error(addr_err.Error())
		return addr_err
	}

	listener, lis_err := net.ListenTCP("tcp", tcp_addr)
	if lis_err != nil {
		ne.log_error(lis_err.Error())
		return lis_err
	}


	for {
		conn, err := listener.Accept()
		if err != nil {
			ne.log_error(err.Error())
			continue
		}

		// 开启线程处理该链接
		ne.log_debug(1, "here")
		new_fd, err := ne.new_conn(conn)
		ne.log_debug(1, "here")
		if err != nil {
			//添加新连接失败
			ne.log_error(err.Error(), "Start")
			return err
		}

		go ne.handle_conn(new_fd)
	}
	return nil
}



/* 在一个连接上设置自动分包规则
 * 该函数会先清除之前的所有分包规则，然后将目前的规则设置上去;
 * 从左往右，第一个到 PACK_EOF_MAX_NUM 个自动分包规则会被应用上，后面的丢弃；
 * 一个连接上可以有多个分包规则，从左往右数起，第一个先匹配上的规则将被应用上；如果所有规则遍历完之后仍得不到数据包，将返回“包未找到"的错误( error(netevent.ERR_PACK_NOT_FOUND) )
 * 
 * @param	fd
 * @param	pack_eof_rule
 */
func (ne *net_event_driver) SetPackEof (fd uint32, peof ... i_net_pack_eof) error {
	conn := ne.conns[fd]
	if conn == nil {
		ne.log_error(ERR_BAD_FD, "SetPackEof", strconv.FormatUint(uint64(fd), 10))
		return errors.New(ERR_BAD_FD)
	}

	conn.pack_eof_num = 0
	for _, p := range peof {
		if conn.pack_eof_num >= PACK_EOF_MAX_NUM {
			break;
		}
		conn.pack_eof[conn.pack_eof_num] = p
		conn.pack_eof_num ++
	}
	return nil
}


/* 清理一个连接上的自动分包规则
 *
 * @param	fd
 */
func (ne *net_event_driver) CleanPackEof (fd uint32) error {
	conn := ne.conns[fd]
	if conn == nil {
		ne.log_error(ERR_BAD_FD, "SetPackEof", strconv.FormatUint(uint64(fd), 10))
		return errors.New(ERR_BAD_FD)
	}

	var i uint8
	for i = 0; i < conn.pack_eof_num; i ++ {
		conn.pack_eof[i] = nil
	}
	return nil
}


/* 发送一个数据包到一个连接
 *
 * @param	fd
 * @param	pack
 */
func (ne *net_event_driver) Send (fd uint32, pack []byte) (int, error) {
	conn,exist := ne.conns[fd]
	if !exist {
		ne.log_error(ERR_BAD_FD, strconv.FormatUint(uint64(fd), 10))
		return 0, errors.New(ERR_BAD_FD)
	}
	if conn.state == CONN_RUNNING {
		return conn.sock.Write(pack)
	}
	return 0, errors.New(ERR_CONN_BAD_STATE)
}


/* 关闭一个连接
 *
 * @param	fd
 */
func (ne *net_event_driver) Close (fd uint32) error {
	conn, exist := ne.conns[fd]
	if !exist {
		ne.log_error(ERR_BAD_FD, strconv.FormatUint(uint64(fd), 10))
		return errors.New(ERR_BAD_FD)
	}
	conn.state= CONN_CLOSED
	return nil
}


/* 停止向一个连接读取数据
 *
 * @param	fd
 */
func (ne *net_event_driver) ShutDown (fd uint32) error {
	conn,exist := ne.conns[fd]
	if !exist {
		ne.log_error(ERR_BAD_FD, strconv.FormatUint(uint64(fd), 10))
		return errors.New(ERR_BAD_FD)
	}
	conn.is_shutdown = true
	if conn.is_shutup {
		conn.state = CONN_CLOSED
	}
	return nil
}


/* 停止向一个连接写入数据
 *
 * @param	fd
 */
func (ne *net_event_driver) ShutUp (fd uint32) error {
	conn,exist := ne.conns[fd]
	if ; !exist {
		ne.log_error(ERR_BAD_FD, strconv.FormatUint(uint64(fd), 10))
		return errors.New(ERR_BAD_FD)
	}
	conn.is_shutup = true
	if conn.is_shutdown {
		conn.state = CONN_CLOSED
	}
	return nil
}


/* 发起一个连接, 返回fd
 *
 * @param	host_port, ip和端口用":"隔开，格式:"*:8888"
 */
func (ne *net_event_driver) Connect (host_port string) (uint32, error) {
	conn, err := net.Dial("tcp", host_port)
	if err != nil {
		ne.log_error(err.Error(), "Connect")
		return 0, err
	}

	// 开启线程处理该链接
	new_fd, new_err := ne.new_conn(conn)
	if new_err != nil {
		//添加新连接失败
		return 0, new_err
	}

	go ne.handle_conn(new_fd)
	return new_fd, nil
}




/* 记录错误日志，如果记录失败，错误信息将连同失败原因一起输出到stderr
 * 错误日志的格式是固定的：CHAT|2014-11-01 11:11:11|error|
 *
 */
func (ne *net_event_driver) log_error (msg ... string) {
	now := time.Now()
	line := "CHAT|" + now.Format("2006-01-02 15:04:05") + "|ERROR|" + strings.Join(msg, "|")

	if ne.ErrorLog == "" {
		os.Stderr.WriteString(line + "\n")
		return
	}

	log_file,_ := ne.tran_log_path(ne.ErrorLog, now)

	f, err := os.OpenFile(log_file, os.O_APPEND | os.O_WRONLY |os.O_CREATE, 0600)
	if err != nil {
		os.Stderr.WriteString(line + "\n" + err.Error() + "\n")
		return
	}

	defer func(){
		err := f.Close()
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
	}()

	_, err = f.WriteString(line + "\n")
	if err != nil {
		os.Stderr.WriteString(line + "\n" + err.Error() + "\n")
		return
	}

	return

}


/* 记录debug日志，如果记录失败，debug信息将连同失败原因一起输出到stderr
 * debug日志的格式固定为：CHAT|2014-11-01 11:11:11|DEBUG|$fd|(recv|conn|close)|(read_bytes:$read_bytes)|(pack_bytes:$pack_bytes)
 */
func (ne *net_event_driver) log_debug (fd uint32, tag string, msg ... string) {
	now := time.Now()
	line := "CHAT|" + now.Format("2006-01-02 15:04:05") + "|DEBUG|" + strconv.FormatUint(uint64(fd), 10) + "|" +tag + "|" + strings.Join(msg, "|")

	if ne.ErrorLog == "" {
		os.Stderr.WriteString(line + "\n")
		return
	}

	log_file,_ := ne.tran_log_path(ne.ErrorLog, now)

	f, err := os.OpenFile(log_file, os.O_APPEND | os.O_WRONLY |os.O_CREATE, 0600)
	if err != nil {
		os.Stderr.WriteString(line + "\n" + err.Error() + "\n")
		return
	}

	defer func(){
		err := f.Close()
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
	}()

	_, err = f.WriteString(line + "\n")
	if err != nil {
		os.Stderr.WriteString(line + "\n" + err.Error() + "\n")
		return
	}

	return

}


/* 将文件路径如 "/tmp/file.%Y%m%d.log" 转换成当前时间对应的 实际路径 ，如 "/tmp/file.20141107.log"
 */
func (ne *net_event_driver) tran_log_path(log_path string, now time.Time) (string, error) {
	return "/tmp/file.20141107.log", nil
}


const ERR_SERVER_FULL_FILLED = "server already full filled"
/* 从 last_fd 的位置开始，搜索整个 conns 列表，找出一个空闲的fd
 * 如果没找到，将返回“服务器已满载”的错误errors.New(ERR_SERVER_FULL_FILLED)
 */
func (ne *net_event_driver) new_sock_fd() (uint32, error) {
	found := false
	i := ne.last_fd + 1

	for  i != ne.last_fd {
		if i >= ne.MaxConnNum {
			i = 0
		}

		if _, exists := ne.conns[i] ; !exists {
			found = true
			break
		}

		i ++
	}

	if found {
		return i, nil
	} else {
		return 0, errors.New(ERR_SERVER_FULL_FILLED)
	}
}


/* 从fd中读取数据，并调用 OnRecv 
 *
 * @param	fd	uint32
 */
func (ne *net_event_driver) conn_read_data(fd uint32) {
	conn, exist := ne.conns[fd]
	if !exist {
		ne.log_error(ERR_BAD_FD)
		return
	}
	var (
		phead	int
		plen	int
		perror	error
		err	error
		res	error
		pack	[]byte

		head	int = 0
		tail	int
		r	int
		bytes	int
		i		uint8
	)

	// 读取一段数据
	bytes, err = conn.sock.Read(conn.buff[head : ])
	tail = head + bytes

	if bytes > 0 {
		if conn.pack_eof_num == 0 {
			pack = conn.buff[head : tail]
			head = 0
			if ne.OnRecv != nil {
				if res = ne.OnRecv(fd, pack); res != nil {
					ne.log_error(res.Error())
				}
			}

		} else {
			head = 0	//rewind

			for head < tail {

				perror = errors.New(ERR_PACK_NOT_FOUND)
				// 自动分包将返回第一个数据包的 begin下标 和 长度 len
				for i = 0 ; i < conn.pack_eof_num; i ++ {
					peof := conn.pack_eof[i]
					if peof != nil {
						phead, plen, perror = peof.fetch(conn.buff[head : tail])
						//这里返回的phead是相对于 buff[head : tail]的下标，需要转换为相对于 buff[0 : ] 的下标
						if perror == nil {

							//校验自定义包规则的返回值是否正确
							if phead > 0 && plen > 0 && (phead + plen) <= (tail - head) {
								phead += head
								break;

							} else {
								ne.log_error(ERR_ILLEAGLE_PACK_EOF)
								perror = errors.New(ERR_ILLEAGLE_PACK_EOF)
							}
						}
					}
				}

				if perror == nil {	//解析到数据包

					//包的位置在 head[] 下标处
					if phead == head {
						pack = conn.buff[head : head + plen]
						head += plen

					} else {
						pack = conn.buff[head : phead]
						head = phead
					}

				} else {	//整个buff中都未找到package
					if head == 0 {
						pack = conn.buff[head : tail]
						head = tail

					} else {
						pack = nil
						r = tail - head
						//另存剩余的数据
						copy(conn.buff[0 : r], conn.buff[head : tail])
						head = r

						break;
					}
				}

				if ne.OnRecv != nil {
					if res = ne.OnRecv(fd, pack); res != nil {
						ne.log_error(res.Error())
					}
				}

			}
		}
	}


	//客户端关闭了连接 
	if err == io.EOF {
		/* do nothing */
	//网络连接异常
	} else if err != nil {
		ne.log_error(err.Error())
	}
	ne.Close(fd)

}


/* 
 * 把一个网络连接加入连接池, 返回的error已经被记录到 ErrorLog 中
 */
func (ne *net_event_driver) new_conn (conn net.Conn) (uint32, error) {
	new_fd, err := ne.new_sock_fd()
	if err != nil {
		ne.log_error(err.Error())
		return 0, err
	}

	new_conn := &connect { sock: conn }
	new_conn.buff = make([]byte, ne.ConnBuffSize)
	new_conn.last_op_time = time.Now().Unix()
	new_conn.pack_eof_num = 0
	new_conn.state = CONN_RUNNING
	new_conn.is_shutdown = false
	new_conn.is_shutup = false
	ne.conns[new_fd] = new_conn

	// 调用 OnConn 回调函数
	if ne.OnConn != nil {
		if err = ne.OnConn(new_fd); err != nil {
			ne.log_error(err.Error())
		}
	}
	return new_fd, nil
}

/* 
 * 清理该连接在连接池里所占用的内存（不关闭该连接对应的socket), 返回的error已经被记录到 ErrorLog 中
 */
func (ne *net_event_driver) clean_conn (fd uint32) error {
	if _, exist := ne.conns[fd]; !exist {
		return errors.New(ERR_BAD_FD)
	}
	delete(ne.conns, fd)
	return nil
}


/* 
 * 处理连接
 */
func (ne *net_event_driver) handle_conn(new_fd uint32) {
	 var (
		 err	error
		 res	error
	 )
	 new_c := ne.conns[new_fd]
	 conn := new_c.sock
	 for {
		 switch new_c.state {

		 case CONN_RUNNING:
			 if ! new_c.is_shutdown {
				 ne.conn_read_data(new_fd)
			 }

		 case CONN_CLOSED:
			 ne.clean_conn(new_fd)
			 if err = conn.Close(); err != nil {
				 ne.log_error(err.Error())
			 }
			 if ne.OnClose != nil {
				 if res = ne.OnClose(new_fd); res != nil {
					 ne.log_error(res.Error())
				 }
			 }
			 break;

		 default:
			 ne.log_error("wrong state", "handle_conn", strconv.FormatUint(uint64(new_fd), 10), strconv.FormatUint(uint64(new_c.state), 10))
			 if err = ne.clean_conn(new_fd); err != nil {
				 ne.log_error(err.Error())
			 }
			 if err = conn.Close(); err != nil {
				 ne.log_error(err.Error())
			 }
			 break;


		 }
	 }

 }


/**************************** 自动分包 ***********************************/

/* 自动分包规则
 *
 * 从字节流中的[0]位开始,根据规则依次搜索到包头,包尾,然后返回包头, 包尾的下标[begin], 和包的长度 length
 * 如果字节流中没有符合该包规则的数据段，返回"包未找到"的错误( error(netevent.ERR_PACK_NOT_FOUND) )
 *
 */
const	ERR_PACK_NOT_FOUND	=	"package not found"
const	ERR_ILLEAGLE_PACK_EOF = "illeagle pack eof"
type i_net_pack_eof interface {
	fetch (stream []byte) (int, int, error)
}


/* 自动分包规则: 尾部分隔符
 *
 */
type PackEofTail struct {
	Eof	string	//分隔符
}
func (pe *PackEofTail) fetch (stream []byte) (int, int, error) {
	return 0, 0, nil
}


/* 自动分包规则：首尾分隔符
 *
 */
type PackEofHeadTail struct {
	HeadEof	string	//首部分隔符
	TailEof	string	//尾部分隔符
}


/* 自动分包规则：HTTP post 包
 */
type PackEofHttpPost struct {
}


/* 自动分包规则: HTTP get 包
 */
type PackEofHttpGet struct {
}


type PackEofTest struct {
}
func (pe *PackEofTest) fetch (stream []byte) (int, int, error) {
	return 0, 4, nil
}
