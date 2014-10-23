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



/* panic exception handler
 *
 */
func Exception () {
	str := recover()
	stack, bytes := runtime.Stack()
	fmt.Println("[ERROR]", str)
	fmt.Printf("[PANIC] of %d bytes: %s", bytes, stack)
}

/* if error != nil, panic
 *
 */
func checkError(err error) {
	if err != nil {
		panic (err.Error)
	}
}



/****************************
 * client side functions 
 ****************************/
const CLI_DEBUG_ON = true	//to open debug log 

/* create a client instance
 *
 */
type client struct {
	sock	net.Conn
	userid	int32
}

func Client (host string, port string) {
}





/****************************
 * server side functions 
 ****************************/
const SRV_DEBUG_ON = true		//to open debug log
const SRV_CONN_BUFF_SIZE = 32		//bytes 


/* create a server instance 
 *
 */
func Server (host string, port string) *server {
	s := new(server)
	s.host = host
	s.port = port
	return s
}


/* run a server instance 
 *
 */
func (s *server) Start () {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.host + ":" + s.port)
	checkError(err)

}


type server struct {
	online	map[int32] *conn
	host	string
	port	string
}

type conn struct {
	sock	net.Conn
	buff	chan []byte
	buff_is_filled	chan bool
	userid	int32
}

