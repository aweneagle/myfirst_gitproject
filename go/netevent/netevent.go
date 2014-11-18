/* net event driver 
 *
 * 微型网络事件驱动
 *
 * by awen, 2014.11.17
 */

package	netevent

import	"../runner"
import	"net"
import	"reflect"
import	"errors"
import	"time"
import	"os"
import	"strings"
import	"strconv"
import	"io"
import	"bytes"


type netevent struct {

	/* 回调函数：返回的error只用于netevent的error_log*/
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

	默认为 false 
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



	/* 在线连接 */
	conns	map[uint32]*connect

	host_port	string	//监听的端口

	last_fd	uint32//上一次被使用的fd


}


const CFG_MAX_CONN_NUM = 65535

/* 生成一个 event driver, 绑定在指定端口上
 *
 * @param	host_port, ip和端口用":"隔开, 格式:"*:8888"
 */
func Init () *netevent {
	ne := new(netevent)

	ne.MaxConnNum = CFG_MAX_CONN_NUM
	ne.conns = make(map[uint32]*connect, ne.MaxConnNum)
	ne.ConnBuffSize = 8 * 1024
	ne.Debug = false
	ne.DebugLog = ""
	ne.ErrorLog = ""
	ne.IdleTime = 0
	ne.last_fd = 0

	ne.OnClose = nil
	ne.OnConn = nil
	ne.OnRecv = nil

	return ne
}



/* 连接远端服务器，并监听数据
 *
 * @param	host_port	string, 格式： "127.0.0.1:8888"
 */
func (ne *netevent) Dial(host_port string) error {
	var (
		conn net.Conn
		err	error
		new_fd	uint32
	)
	conn, err = net.Dial("tcp", host_port)
	if err != nil {
		ne.log_error("Dial: Dial", err.Error())
		return err
	}

	new_fd, err = ne.add_connect(conn)
	if err != nil {
		ne.log_error("Dial: add_connect", err.Error())
		if err = conn.Close(); err != nil {
			ne.log_error("Dial: conn.Close", err.Error())
		}
		return err
	}

	return ne.run_connect(new_fd)
}



/* 监听端口
 *
 * @param	host_port	string, 格式： "127.0.0.1:8888"
 */
func (ne *netevent) Listen(host_port string) error {
	var (
		conn net.Conn
		err	error
		new_fd	uint32
		tcp_addr	*net.TCPAddr
		listener	*net.TCPListener
	)

	tcp_addr, err = net.ResolveTCPAddr("tcp", host_port)
	if err != nil {
		ne.log_error("Listen: net.ResolveTCPAddr", err.Error())
		return err
	}

	listener, err = net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		ne.log_error("Listen: net.ListenTCP", err.Error())
		return err
	}

	for {
		if conn, err = listener.Accept() ; err != nil {
			ne.log_error("Listen: Accept", err.Error())
			continue
		}
		if new_fd, err = ne.add_connect(conn); err != nil {
			ne.log_error("Listen: add_connect", err.Error())
			if err = conn.Close(); err != nil {
				ne.log_error("Listen: Close", err.Error())
			}
			continue
		}

		go ne.run_connect(new_fd)
	}
}


/* 连接远端服务器, 返回fd
 *
 * @param	host_port	string, 格式： "127.0.0.1:8888"
 */
func (ne *netevent) Connect(host_port string) (uint32, error) {
	var (
		conn net.Conn
		err	error
		new_fd	uint32
	)
	conn, err = net.Dial("tcp", host_port)
	if err != nil {
		ne.log_error("Connect: Dial", err.Error())
		return 0, err
	}

	if new_fd, err = ne.add_connect(conn); err != nil {
		ne.log_error("Connect: add_connect", err.Error())
		if err = conn.Close(); err != nil {
			ne.log_error("Connect: conn.Close", err.Error())
		}
		return 0, err
	}

	go ne.run_connect(new_fd)

	return new_fd, nil

}


/* 向一个连接发送数据包
 *
 */
func (ne *netevent) Send(fd uint32, pack []byte) (int, error) {
	if c,exists := ne.conns[fd] ; exists {
		return c.Send(pack)
	}
	return 0, errors.New("bad fd")
}


/* 向一个连接发送数据包，并等待返回结果
 *
 */
func (ne *netevent) Request(fd uint32, req []byte, resp []byte) (int, error) {
	if c,exists := ne.conns[fd] ; exists {
		return c.Request(req, resp)
	}
	return 0, errors.New("bad fd")
}


/* 关闭一个连接
 *
 * @param	fd
 */
func (ne *netevent) Close (fd uint32) error {
	if _, exist := ne.conns[fd] ; !exist {
		return errors.New("bad fd")
	}
	return ne.conns[fd].Close()
}

/* 在一个连接上设置自动分包规则
 * 该函数会先清除之前的所有分包规则，然后将目前的规则设置上去;
 * 从左往右，第一个到 PACK_EOF_MAX_NUM 个自动分包规则会被应用上，后面的丢弃；
 * 一个连接上可以有多个分包规则，从左往右数起，第一个先匹配上的规则将被应用上；如果所有规则遍历完之后仍得不到数据包，将返回“包未找到"的错误( error(netevent.ERR_PACK_NOT_FOUND) )
 * 
 * @param	fd
 * @param	pack_eof_rule
 */
func (ne *netevent) SetPackEof (fd uint32, peof ... i_net_pack_eof) error {
	return nil
}


/* 清理一个连接上的自动分包规则
 *
 * @param	fd
 */
func (ne *netevent) CleanPackEof (fd uint32) error {
	return nil
}




/* 为一个连接启动守护 routine
*/
func (ne *netevent) run_connect(fd uint32) error {
	var (
		c	*connect
		exists	bool
		err	error
		read_len	int
	)
	if c, exists = ne.conns[fd] ; !exists {
		ne.log_error("run_connect: bad fd")
		return errors.New("run_connect: bad fd")
	}
	if c.OnRecv == nil {
		ne.log_error("run_connect: no OnRecv found")
		return errors.New("run_connect: no OnRecv found")
	}
	for {
		if c.shutdown {
			/* 关闭routine */
			if c.OnClose != nil {
				c.OnClose()
			}
			break
		}
		if read_len, err = c.listen_and_read(); err != nil {
			/* 客户端断开连接 | 其他异常情况出现 */
			if err == io.EOF {
				ne.log_debug(fd, "closed")
			} else {
				ne.log_error("run_connect: " + err.Error())
			}
			if err = c.Close(); err != nil {
				ne.log_error("run_connect: " + err.Error())
			}
			return err
		}
		if read_len > 0 {
			ne.log_debug(fd, "read", strconv.FormatInt(int64(read_len), 10))
		}
	}
	return nil
}

/* 添加一个连接
 */
func (ne *netevent) add_connect(conn net.Conn) (uint32, error) {
	var (
		new_fd	uint32
		err	error
		c	*connect
	)

	new_fd, err = ne.new_sock_fd()
	if err != nil {
		return 0, err
	}

	c = NewConn(conn, ne.ConnBuffSize)


	c.OnRecv = func (data []byte) error {
		if ne.OnRecv != nil {
			err := ne.OnRecv(new_fd, data)
			return  err
		}
		return errors.New("no OnRecv found")
	}

	c.OnClose = func () {
		if ne.OnClose != nil {
			_ = ne.OnClose(new_fd)
		}
		delete (ne.conns, new_fd)
	}

	ne.conns[new_fd] = c

	if ne.OnConn != nil {
		ne.OnConn(new_fd)
	}
	ne.log_debug(new_fd, "connect")

	return new_fd, nil

}

/* 记录错误日志，如果记录失败，错误信息将连同失败原因一起输出到stderr
 * 错误日志的格式是固定的：CHAT|2014-11-01 11:11:11|error|
 *
 */
func (ne *netevent) log_error (msg ... string) {
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
func (ne *netevent) log_debug (fd uint32, tag string, msg ... string) {
	if !ne.Debug {
		return
	}

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
func (ne *netevent) tran_log_path(log_path string, now time.Time) (string, error) {
	return "/tmp/file.20141107.log", nil
}

/* 生成新的fd */
func (ne *netevent) new_sock_fd() (uint32, error) {
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
		return 0, errors.New("socket full filled")
	}
}



/********************* connection **************************/
type connect struct {
	buff_size	int

	pack_eof	[PACK_EOF_MAX_NUM] i_net_pack_eof	//自动分包规则
	pack_eof_num	uint8	//自动分包规则数量
	pack_head	int

	buff_len	int

	/* 网络事件处理器 */
	runner	*runner.Runner

	/* socket */
	sock	net.Conn

	/* 关闭服务 */
	shutdown	bool


	/* data buffer */
	buff	[]byte


	/* OnRecv 事件回调 */
	OnRecv	func([]byte) error

	/* OnClose 事件回调 */
	OnClose	func()

	/* 数据监听 */
	read	read_data

}

const PACK_EOF_MAX_NUM = 6

func NewConn(sock net.Conn, buffsize int) *connect{
	c := & connect { shutdown: false, sock: sock}
	c.runner = runner.Init()
	c.buff = make([]byte, buffsize)

	c.OnRecv = nil
	c.OnClose = nil

	c.buff_size = buffsize
	c.pack_eof_num = 0
	c.pack_head = 0
	c.buff_len = 0


	c.read = read_data {
		sock : c.sock,
		buff : c.buff[0 : buffsize],
		read_len : 0,
		err	: nil,
	}
	return c
}


/* 向连接发送请求，并等待响应 
*/
func (c *connect) Request(req []byte, response []byte) (int, error) {
	var r = &request{ sock:c.sock, req: req, resp: response , err:nil, bytes:0 }
	if c.shutdown {
		return 0, errors.New("connect closed")
	}
	if err := c.runner.Request(r.Handle); err != nil {
		return 0, err
	} else {
		return r.bytes, r.err
	}
}


/* 向连接发送数据 
*/
func (c *connect) Send(pack []byte) (int, error) {
	var r = &send_data{ sock:c.sock , pack: pack, err:nil, bytes:0 }
	if c.shutdown {
		return 0, errors.New("connect closed")
	}
	if err := c.runner.Request(r.Handle); err != nil {
		return 0, err
	} else {
		return r.bytes, r.err
	}
}


/* 关闭连接 
*/
func (c *connect) Close() error {
	//关闭 Run, Request，停止接受事件
	c.shutdown = true
	//关闭 runner, 这里会等待所有事件完成才退出
	c.runner.Quit()
	return c.sock.Close()
}



/* 设置自动分包规则
 * 该函数会先清除之前的所有分包规则，然后将目前的规则设置上去;
 * 从左往右，第一个到 PACK_EOF_MAX_NUM 个自动分包规则会被应用上，后面的丢弃；
 * 一个连接上可以有多个分包规则，从左往右数起，第一个先匹配上的规则将被应用上；如果所有规则遍历完之后仍得不到数据包，将返回“包未找到"的错误( error(netevent.ERR_PACK_NOT_FOUND) )
 * 
 * @param	pack_eof_rule
 */
func (c *connect) SetPackEof(peof ... i_net_pack_eof) error {
	r := &set_pack_eof{ c:c, peof:peof }
	return c.runner.Request(r.Handle)
}


/* 清理自动分包规则
 *
 */
func (c *connect) CleanPackEof() error {
	r := &clean_pack_eof{ conn:c }
	return c.runner.Request(r.Handle)
}

type set_pack_eof struct {
	c	*connect
	peof	[]i_net_pack_eof
}

func (s *set_pack_eof) Handle () {
	s.c.clean_pack_eof()
	for i,peof := range s.peof {
		if i < PACK_EOF_MAX_NUM {
			s.c.pack_eof[i] = peof
		}
	}
	s.c.pack_eof_num = uint8(len(s.peof))
}

type clean_pack_eof struct {
	conn	*connect
}

func (c *clean_pack_eof) Handle () {
	c.conn.clean_pack_eof()
}


/* 清除自动分包规则 */
func (c *connect) clean_pack_eof() {
	var i uint8
	for i = 0; i < c.pack_eof_num; i ++ {
		c.pack_eof[i] = nil
	}
	c.pack_eof_num = 0
}



/* 向 网络连接 发送数据 */
type send_data struct {
	sock	net.Conn
	pack	[]byte
	err	error
	bytes	int
}

func (s *send_data) Handle() {
	s.bytes, s.err = s.sock.Write(s.pack)
}

/* 向 网络连接 发送数据，并等待请求返回 */
type request struct {
	sock	net.Conn
	req	[]byte
	resp	[]byte
	err	error
	bytes	int
}

func (s *request) Handle() {
	s.bytes = 0
	_, s.err = s.sock.Write(s.req)
	if s.err != nil {
		return
	}
	s.bytes, s.err = s.sock.Read(s.resp)
}



/* 从 网络连接 中读取数据, 并传递给回调函数 */
type read_data struct {
	sock	net.Conn
	buff	[]byte
	read_len	int
	err	error
}

func (s *read_data) Handle() {
	if err := s.sock.SetReadDeadline(time.Now().Add(10 * time.Microsecond)); err != nil {
		s.err = err
		return
	}
	s.read_len, s.err = s.sock.Read(s.buff[0 : ])

}

func (conn *connect) listen_and_read () (int, error) {

	var (
		err	error	=	nil
		phead	int		=	0	//package head
		buff_len	int	=	0	//buff length
		plen	int		=	0	//package length
		read_len	int
	)

	conn.read.buff = conn.buff[conn.buff_len : ]
	conn.runner.Request( conn.read.Handle )
	err = conn.read.err

	//读超时，直接返回
	if err != nil {
		_, has_method_timeout := reflect.TypeOf(err).MethodByName("Timeout")
		if has_method_timeout && err.(net.Error).Timeout() {
			return 0, nil
		}
	}

	// 从socket中读取数据
	buff_len = conn.buff_len
	phead = conn.pack_head
	read_len = conn.read.read_len


	buff_len += read_len

	// 如果没有自动分包规则，直接将数据传递给OnRecv
	if conn.pack_eof_num == 0 {
		plen = buff_len - phead

		conn.OnRecv(conn.buff[phead : phead + plen])
		phead += plen

	// 如果有自动分包规则，过滤分包规则并传递给OnRecv
	} else {

		for {
			plen = conn.fetch_package(conn.buff[phead : buff_len])
			if plen < 0 || phead + plen > buff_len {
				break
			}
			conn.OnRecv(conn.buff[phead : phead + plen])
			phead += plen
		}
	}


	if buff_len >= conn.buff_size {
		if phead == 0 {
			//当buff已满，仍未找到数据包的情况下，该buff内数据被丢弃，连接将被关闭
			return read_len, errors.New("wrong package")

		} else {
			if phead < buff_len {
				copy(conn.buff[0 : buff_len - phead], conn.buff[ phead : buff_len ])
			}
			buff_len -= phead
			phead = 0
		}
	} else if buff_len == phead {
		//所有的数据都已经被返回给OnRecv, 此时buff需要清空
		buff_len -= phead
		phead = 0
	}

	//log_debug(fd, "read", "-----final phead:", strconv.FormatUint(uint64(phead), 10), ",buff_len:", strconv.FormatUint(uint64(buff_len), 10))

	conn.pack_head = phead
	conn.buff_len = buff_len

	return read_len, err
}

/* 从buffer中自动抓包
 *
 * @return	package_lentgh, int
 */
func (c *connect) fetch_package (buff []byte) int {
	var (
		plen int
		i	uint8
	)
	for i = 0; i < c.pack_eof_num; i ++ {
		peof := c.pack_eof[i]
		if peof != nil {
			plen = peof.fetch(buff)
			if plen > 0 {
				return plen
			}
		}
	}
	return -1
}


/**************************** 自动分包 ***********************************/

/* 自动分包规则
 *
 * 从字节流中的[0]位开始,根据规则搜索包，如果数据包未找到，返回 -1, 否则返回数据包长度 len
 *
 *	fetch (stream []byte) (len int)
 */
const	ERR_PACK_NOT_FOUND	=	"package not found"
const	ERR_ILLEAGLE_PACK_EOF = "illeagle pack eof"
type i_net_pack_eof interface {
	fetch (stream []byte) int
}


/* 自动分包规则: 尾部分隔符
 *
 */
type PackEofTail struct {
	Eof	string	//分隔符
}
func (pe *PackEofTail) fetch (stream []byte) int {
	var (
		e	[]byte
		elen	int
		index	int
	)
	e = []byte(pe.Eof)
	elen = len(e)
	index = bytes.Index(stream, e)
	if index  == -1 {
		return -1
	} else {
		return index + elen
	}
}


/* 自动分包规则：首尾分隔符
 *
 */
type PackEofHeadTail struct {
	HeadEof	string	//首部分隔符
	TailEof	string	//尾部分隔符
}
func (pe *PackEofHeadTail) fetch (stream []byte) int {
	var (
		tail	[]byte
		tail_pos	int
		head_len	int
		tail_len	int
		stream_len	int
	)
	head_len = len(pe.HeadEof)
	tail_len = len(pe.TailEof)
	stream_len = len(stream)

	if stream_len < head_len + tail_len {
		return -1
	}

	//match the head
	if bytes.Compare(stream[0:head_len], []byte(pe.HeadEof)) != 0 {
		return -1
	}

	if tail_len == 0 {
		return stream_len

	} else {
		tail = []byte(pe.TailEof)
		tail_pos = bytes.Index(stream, tail)

		if tail_pos == -1 {
			return -1
		} else {
			return tail_pos + len(tail)
		}
	}

}


/* 自动分包规则：HTTP post 包
 */
type PackEofHttpPost struct {
	rule	PackEofHeadTail
}
func (pe *PackEofHttpPost) fetch (stream []byte) int {
	pe.rule.HeadEof = "POST"
	pe.rule.TailEof = ""
	return pe.rule.fetch(stream)
}


/* 自动分包规则: HTTP get 包
 */
type PackEofHttpGet struct {
	rule	PackEofHeadTail
}
func (pe *PackEofHttpGet) fetch (stream []byte) int {
	pe.rule.HeadEof = "GET"
	pe.rule.TailEof = "\r\n\r\n"
	return pe.rule.fetch(stream)
}


/* 自动分包规则: CHAT 登陆包
 */
type PackEofChat struct {
	rule	PackEofHeadTail
}
func (pe *PackEofChat) fetch (stream []byte) int {
	pe.rule.HeadEof = "CHAT"
	pe.rule.TailEof = "\r\n\r\n"
	return pe.rule.fetch(stream)
}

