/* version 1.0
 * 
 * a simple server package 
 * 
 * by awen, 2014.10.23
 */
package web
import "fmt"
import "time"
import "os"
import "net"
import "runtime"
import "strconv"
import "strings"

const LOGIN_PACKAGE_LEN = 8		//must be of length 1 ~ 8


/* panic exception handler
 *
 */
func Exception () {
	str := recover()
	stack := make([]byte, 1024)
	bytes := runtime.Stack(stack, true)
	fmt.Println("[ERROR]", str)
	fmt.Printf("[PANIC] of %d bytes: %s", bytes, stack)
	os.Exit(-1)
}

func CheckError(err error, messages ... string) {
	if err != nil {
		logmsg := ""
		for _,msg := range messages {
			logmsg += msg
		}
		checkError(err, logmsg)
	}

}

func CheckWarning(err error, messages ... string) bool {
	if err != nil {
		logmsg := ""
		for _,msg := range messages {
			logmsg += msg
		}
		return checkWarning(err, logmsg)
	}
	return true
}

/* if error != nil, panic
 *
 */
func checkError(err error, messages ... string) {
	if err != nil {
		logmsg := ""
		for _,msg := range messages {
			logmsg += msg
		}
		panic (err.Error)
		fmt.Println("[ERROR]" + logmsg)
	}
}

/* log error message 
 *
 */
func logError(errors ... string) {
	var err = ""
	for _,msg := range errors {
		err += "|" + msg
	}
	fmt.Println("[ERROR]", err)
}


/* if error != nil, print out error message continue
 * 
 * @return	if error != nil , return false;  else return true
 */
func checkWarning(err error, messages ... string) bool {
	if err != nil {
		logmsg := ""
		for _,msg := range messages {
			logmsg += msg
		}
		fmt.Printf("[WARNING]%s|%s\n", err.Error(), logmsg)
		return false
	}
	return true
}



/****************************
 * client side functions 
 ****************************/
const CLI_DEBUG_ON = true	//to open debug log 
const CLI_BUFF_SIZE = 32	//bytes

type client struct {
	sock	net.Conn
	userid	uint64
	host	string
	port	string
}

/* create a client instance
 *
 */
func Client (host string, port string) *client {
	cli := new (client)
	cli.host = host
	cli.port = port
	return cli
}

func (c *client) Login (userid uint64) {
	conn , err := net.Dial("tcp", c.host + ":" + c.port)
	checkError(err)
	c.sock = conn
	c.userid = userid

	userid_str := strconv.FormatUint(userid, 10)
	userid_str = strings.Repeat("0", LOGIN_PACKAGE_LEN - len(userid_str)) + userid_str

	_, err = c.sock.Write([]byte(userid_str))
	checkError(err)
}

func (c *client) StrLogin(userid string) {
	uid, err := strconv.ParseUint(userid, 10, 64)
	checkError(err)
	c.Login(uid)
}


/* send message 
 *
 */
func (c *client) Send(msg []byte) {
	var err error
	n := 0
	for length := len(msg); length > 0; length -= n {
		n , err = c.sock.Write(msg[n:])
		if !checkWarning(err) {
			c.sock.Close()
			return
		}
	}
}


/* recv message , return bytes we got
 *
 */
func (c *client) Recv(msg []byte) int {
	if len(msg) < CLI_BUFF_SIZE {
		logError("buff size is not enough, need size " + strconv.FormatInt(int64(CLI_BUFF_SIZE), 10) + ", given " + strconv.FormatInt(int64(len(msg)), 10))
		return  0
	}
	n, err := c.sock.Read(msg[0:])
	checkError(err)
	return n
}





/****************************
 * server side functions 
 ****************************/
const SRV_DEBUG_ON = true		//to open debug log
const SRV_CONN_BUFF_SIZE = 16		//bytes 
const SRV_MONITOR_SLEEP = 2			//seconds


/* create a server instance 
 *
 */
func Server (host string, port string) *server {
	s := new(server)
	s.host = host
	s.port = port
	s.online = make(map[uint64] *connect)
	return s
}


/* run a server instance 
 *
 */
func (s *server) Start () {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.host + ":" + s.port)
	checkError(err)

	listener, l_err := net.ListenTCP("tcp", tcpAddr)
	checkError(l_err)
	go s.monitor()

	for {
		conn, a_err := listener.Accept()
		if !checkWarning(a_err,"accept") {
			continue;
		}

		go s.addConn(conn);

	}

}


/* monitor server's status
 *
 */
func (s *server) monitor() {
	for {
		fmt.Println("[MONITOR]total conn ", len(s.online))
		time.Sleep(SRV_MONITOR_SLEEP * time.Second)
	}
}


/* add a connection into server 
 *
 * when a new conn comes in, server will wait for it's login package forever until received it, if something wrong happended before login package
 * arrived, or if wrong login package received, addConn() will fail, and warning will be print out, other wise new conn will be add into server as a member of "online members"
 *
 * @return true or false
 */
func (s *server) addConn(conn net.Conn) bool {
	var buff [LOGIN_PACKAGE_LEN]byte
	for length := 0; length < LOGIN_PACKAGE_LEN; {
		n, err := conn.Read(buff[length:])

		if !checkWarning(err, "addconn") {
			return false
		}

		length += n
	}

	userid, int_err := strconv.ParseUint(string(buff[0:]), 10, 64)
	if !checkWarning(int_err, "parse_uid") {
		return false
	}

	if _,exist := s.online[userid]; exist {
		logError("userid already login, id=", string(buff[0:]))
		return false
	}

	c := new (connect)
	c.sock = conn
	c.buff = make (chan []byte)
	c.buff_is_empty = make (chan bool)
	c.userid = userid

	s.online[userid] =  c

	go c.readIn(s)
	go c.writeOut(s)

	return true
}



/* read message from socket 
 *
 */
func (c *connect) readIn(s *server) {
	var buff [SRV_CONN_BUFF_SIZE]byte
	for {
		n, err := c.sock.Read(buff[0:])

		if !checkWarning(err, "readin") {
			c.sock.Close()
			delete(s.online, c.userid)
			return
		}

		c.buff <- buff[0:n]

		buff_is_empty := <-c.buff_is_empty
		fmt.Printf("data received\tbuff_is_empty:%t\tbytes:%d\n", buff_is_empty, n)
	}
}


/* write out message into other sockets 
 *
 */

func (c *connect) writeOut(s *server) {
	for {
		buff := <-c.buff


		for userid, peer :=  range s.online {

			if c.userid == userid  {
				continue
			}

			_, err := peer.sock.Write(buff[0:])
			if !checkWarning(err, "writeout") {
				break;
			}
		}

		c.buff_is_empty <- true
	}
}




type server struct {
	online	map[uint64] *connect

	host	string
	port	string
}

type connect struct {
	sock	net.Conn
	buff	chan []byte
	buff_is_empty	chan bool
	userid	uint64
}

