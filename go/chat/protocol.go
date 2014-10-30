/* protocols of chat server
 *
 *
 * by awen , 2014.10.28
 */


package chat


import "encoding/binary"

/* panic error no */
const P_ERR_READ_SOCK = 1
const P_ERR_TOO_LARGE_PACKAGE = 2
const P_ERR_WRONG_LOGIN_PACK_LEN = 3
const P_ERR_WRONG_DATA_LEN = 4
const P_ERR_RELOGIN = 5


/* 
 * packages : 
 *
 * client <-> server 
 * server <-> server
 *
 *		1.0  version
 *	login	[package_len (2|8bytes)] [version (2bytes) = "10"] [userid (4|8bytes)] [token (32bytes)] [role 1byte]
 *	data	[package_len (2|8bytes)] [receiver (4|8bytes)] [data bytes ...]
 *	heart_beat	[package_len = "00" (2bytes)]
 * 
 *		2.0	version
 *	login	[package_len (2|8bytes)] [version (2bytes) = "20"] [userid (4|8bytes)] [token (32bytes)] [role 1byte]		
 *	data	[package_len (2|8bytes)] [cmd = 1 (1byte)] [receiver (4|8bytes)] [data bytes ...]
 *	heart_beat	[package_len = 3 (2bytes)] [cmd = 2 (1byte)]
 * 
 */
const CONN_MAX_BUFF_LEN = 1024

type connect struct {
	sock	*net.Conn
	buff	[CONN_MAX_BUFF_LEN]byte
	buff_len	uint64		//total data read from socket 

	pack	struct {
		data	[]byte
		data_len	uint64
		pack_len	uint64
	}

	user_login	bool
	user	pk_login
}





/* login package */
const	ROLE_PROXY = 0
const	ROLE_CLIENT = 1

type pk_login struct {
	version	string
	userid	uint64
	token	string
	role	uint8	//ROLE_CLIENT (client -> server), ROLE_PROXY (server -> server)
}

/* heart package */
type pk_heartbeat struct {
}

/* data package */
const	RECV_TYPE_USER = 1
const	RECV_TYPE_GROUP = 0

type pk_data struct {
	receiver	uint64
	receiver_type	uint4	// RECV_TYPE_USER ,  RECV_TYPE_GROUP
	data	[]byte

	orig_pack	[]byte		/* original package, used for copying a whole package */
}



/* create a new connect object to handle packages from client or proxy server
 * 
 * @param	c	*net.Conn, socket
 * @return	conn	*connect
 */
func CreateConn(c *net.Conn) conn *connect{
	conn := new (connect)
	conn.sock = c
	conn.buff_len = 0

	conn.pack.data = nil
	conn.pack.data_len = 0

	conn.user_login = false
	return conn
}


/* read in a whole package from socket into buff
 *
 */
func (c *connect) read_in_package() {

	bytes, err := c.sock.Read(c.buff[0 : ] )
	if err != nil {
		panic(P_ERR_READ_SOCK)
	}

	c.pack.data_len , num_bytes := bytes2num_2_8(c.buff[0 : 8])
	c.pack.pack_len = num_bytes + c.pack.data_len
	if c.pack.pack_len > CONN_MAX_BUFF_LEN {
		panic(P_ERR_TOO_LARGE_PACKAGE)
	}

	c.pack.data = c.buff[num_bytes : num_bytes + c.pack.data_len]

	for bytes < num_bytes + c.pack.data_len {
		n, err := c.sock.Read(c.buff[bytes : ])
		bytes += n

		if err != nil {
			panic(P_ERR_READ_SOCK)
		}

	}
	c.buff_len = bytes

}


/* fetch login info 
 *	login	[package_len (2|8bytes)] [version (2bytes) = "10"] [userid (4|8bytes)] [token (32bytes)] [role 1byte]
 */
func (c *connect) Login(){
	c.read_in_package()

	if c.pack.data_len != 40 || c.pack.data_len != 44 {
		panic(P_ERR_WRONG_LOGIN_PACK_LEN)
	}
	var bytes int
	var role uint64
	i := 2

	c.user.version = string(c.pack.data[0 : i])
	c.user.userid,bytes = bytes2num_4_8(c.pack.data[i : ])

	i += bytes
	c.user.token = string(c.pack.data[i : i+32])

	i += 32
	role,_ = binary.Uvarint(c.pack.data[i : i+1])
	c.user.role = uint8(role)

}

/* write buff data into socket 
 *
 */
func (c *connect) Write (buff []byte) {
	c.sock.Write(buff)
}




/**************** protocol  version 1.0 **********************
 *	data	[package_len (2|8bytes)] [receiver (4|8bytes)] [data bytes ...]
 *	heart_beat	[package_len = "00" (2bytes)]
 */
func (c *connect) ReadIn_V1 () interface {} {
	c.read_in_package()

	if c.pack.data_len == 0 {
		/* heartbeat */
		return new(pk_heartbeat)

	} else if c.pack.data_len < 4{
		panic(P_ERR_WRONG_DATA_LEN)

	} else {
		/* data frame */
		data := new (pk_data)

		var num_bytes int
		var receiver uint64
		receiver, num_bytes = bytes2num_4_8(c.pack.data[0 : ])
		data.receiver, data.receiver_type = uint64_to_receiver(receiver)

		data.data = c.pack.data[num_bytes : c.pack.data_len - num_bytes]

		data.orig_pack = c.buff[0 : c.pack.pack_len]

		return data
	}
}

func (c *connect) Handle (cmd interface{}, s *server) {
	switch cmd.(type) {
	case pk_login:
		panic(P_ERR_RELOGIN)

	case pk_data:
		/* send out data package */
		switch cmd.receiver_type {
		case RECV_TYPE_USER:
			s.send_to_user(cmd);

		case RECV_TYPE_GROUP:
			s.send_to_group(cmd);

		default:
			fmt.Println("[WARNING] wrong receiver type")

		}

	case pk_heartbeat:
		/* heart beat */
	}
}






/************ to find out "how much bytes need to represent a number" , or "what number is represented by given bytes" *************/

func bytes2num_4_8(buff []byte) number uint64, bytes_len int{
}

func bytes2num_2_8(buff []byte) number uint64, bytes_len int{
}

func num2bytes_2_8(buff []byte, number uint64) bytes_len int {
}

func num2bytes_4_8(buff []byte, number uint64) bytes_len int {
}

func uint64_to_receiver( number uint64 ) receiver uint64, receiver_type uint4 {
}

func receiver_to_uint64( receiver uint64, receiver_type uint4 ) uint64 number {
}

/*************************************************************************************************************************************/


type http struct {
	server	string	// host":"port
	sock	*net.Conn
	keep_alive	bool	//whether to keep alive or not
}

func (h *http) read_in() interface{} {
}

func (h *http) write_out( buff []byte ){
}

func (h *http) end(){
}

/* send a http request , and recv a response
*/
func request(uri string) {
}


func CreateHttp(c *net.Conn){
	http := new (http)
	http.sock = c
	keep_alive = true
}



