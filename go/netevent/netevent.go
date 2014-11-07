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
 *		ne.OnConnect	= func (fd uint64) {  
 *		//添加自动分包规则
 *			ne.SetPackEof(fd, netevent.PackEof { Eof:"_*_*_" } );
 *		}
 *
 *		//注册handler， 监听该端口的"数据包到达"事件
 *		ne.OnRecv		= func (pack []byte, fd uint64) {  
 *			//发送数据包
 *			ne.Send(fd, pack);
 *		}
 * 
 *		//注册handler,	监听该端口的"连接关闭"事件
 *		ne.OnClose		= func (fd uint64) {  
 *			//清理自动分包规则
 *			ne.CleanPackEof(fd);
 *		}
 *
 */

package	netevent
import	"net"
import	"strings"
import	"time"
import	"os"

type net_event_driver struct{
	OnRecv	func(uint64, []byte) error
	OnClose	func(uint64) error
	OnConn	func(uint64) error

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




	host_port	string	//监听的端口

	conns	map[uint64] conn

}

const	PACK_EOF_MAX_NUM	=	6	//每个连接所能拥有的最大的自动分包规则数量
type conn struct {
	sock	net.Conn
	buff	[]byte

	pack_eof	[PACK_EOF_MAX_NUM] i_net_pack_eof	//自动分包规则
	last_op_time	int	//上一次操作时间(send or recv)
}


/* 生成一个 event driver, 绑定在指定端口上
 *
 * @param	host_port, ip和端口用":"隔开, 格式:"*:8888"
 */
func Init (host_port string) *net_event_driver {
	ne := new(net_event_driver)

	ne.host_port = host_port
	ne.conns = make(map[uint64]conn)
	ne.ConnBuffSize = 8 * 1024
	ne.Debug = true
	ne.DebugLog = ""
	ne.ErrorLog = ""
	ne.IdleTime = 0

	return ne
}

func (ne *net_event_driver) TLogErr (str string) {
	ne.log_error(str)
}

/* 启动 event driver 
 *
 */
func (ne *net_event_driver) Start() error {
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
		buff := make([]byte, 1024)
		conn.Read(buff)
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
func (ne *net_event_driver) SetPackEof (fd uint64, peof ... i_net_pack_eof) error {
	return nil
}

/* 清理一个连接上的自动分包规则
 *
 * @param	fd
 */
func (ne *net_event_driver) CleanPackEof (fd uint64) error {
	return nil
}


/* 发送一个数据包到一个连接
 *
 * @param	fd
 * @param	pack
 */
func (ne *net_event_driver) Send (fd uint64, pack []byte) error {
	return nil
}


/* 关闭一个连接
 *
 * @param	fd
 */
func (ne *net_event_driver) Close (fd uint64) error {
	return nil
}


/* 关闭一个连接的输出端
 *
 * @param	fd
 */
func (ne *net_event_driver) ShutDown (fd uint64) error {
	return nil
}


/* 发起一个连接
 *
 * @param	host_port, ip和端口用":"隔开，格式:"*:8888"
 */
func (ne *net_event_driver) Connect (host_port string) (uint64, error) {
	return 0, nil
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
func (ne *net_event_driver) log_debug (fd uint64, tag string, args ... string) {
}


/* 将文件路径如 "/tmp/file.%Y%m%d.log" 转换成当前时间对应的 实际路径 ，如 "/tmp/file.20141107.log"
 */
func (ne *net_event_driver) tran_log_path(log_path string, now time.Time) (string, error) {
	return "/tmp/file.20141107.log", nil
}



/**************************** 自动分包 ***********************************/

/* 自动分包规则
 *
 * 从字节流中的[0]位开始,根据规则依次搜索到包头,包尾,然后返回包头, 包尾的下标[begin][end];
 * 如果字节流中没有符合该包规则的数据段，返回"包未找到"的错误( error(netevent.ERR_PACK_NOT_FOUND) )
 *
 */
const	ERR_PACK_NOT_FOUND	=	"package not found"
type i_net_pack_eof interface {
	fetch (stream []byte) (uint64, uint64, error)
}


/* 自动分包规则: 尾部分隔符
 *
 */
type PackEofTail struct {
	Eof	string	//分隔符
}
func (pe *PackEofTail) fetch (stream []byte) (uint64, uint64, error) {
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

