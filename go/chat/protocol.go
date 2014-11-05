/* protocols of chat server
 *
 *
 * by awen , 2014.10.28
 */


package chat


import	"encoding/binary"
import	"errors"

/* panic error no */
const P_ERR_READ_SOCK = 1
const P_ERR_TOO_LARGE_PACKAGE = 2
const P_ERR_WRONG_LOGIN_PACK_LEN = 3
const P_ERR_WRONG_DATA_LEN = 4
const P_ERR_RELOGIN = 5
const P_ERR_WRONG_LOGIN_ROLE = 6
const P_ERR_PROXY_NOT_ALLOW_HEARTBEAT = 7
const P_ERR_WRONG_PROTOCOL_VERSION = 8
const P_ERR_TOKEN_TOO_LONG = 9


/* 
 * packages : 
 *
 * client <-> server 
 * server <-> server
 *
 *		1.0  version
 *	user_login	[protocol (4 bytes) = "CHAT"] [package_len (2|8bytes)] [version (2bytes) = "10"] [role = "1" 1byte] [userid (4|8bytes)] [token (32bytes)] 
 *	proxy_login	[protocol (4 bytes) = "CHAT"] [package_len (2|8bytes)] [version (2bytes) = "10"] [role = "0" 1byte] [server (32bytes)] [pwd (32bytes)] 
 *	data	[package_len (2|8bytes)] [receiver (4|8bytes)] [data bytes ...]
 *	heart_beat	[package_len = "00" (2bytes)]
 * 
 * 
 */
const CONN_MAX_BUFF_LEN = 1024

const PROTOCOL_VERSION_1 = "10"

type connect struct {
	sock	*stream
	read_buff	[]byte
	r_buff_len	uint64		//total data read from socket 


	pack	struct {
		data	[]byte
		data_len	uint64
		pack_len	uint64
	}


	cmd	command
	login	command

	is_login	bool
	role	uint64

	version	string

	_cmds	struct	{
		user_login	pk_user_login

		proxy_login	pk_proxy_login

		v1_data	v1_pk_data	//transport data

		v1_heartbeat	v1_pk_heartbeat	//heartbeat
	}
}

type command interface {
	init_from_buff(c *connect) error
	send_to_conn(c *connect) error
}

/* create a new connect object to handle packages 
 * 
 * @param	c	*stream, buffered socket
 * @return	conn	*connect
 */
func create_connect(c *stream) *connect{
	conn := new (connect)
	conn.sock = c
	conn.r_buff_len = 0

	conn.read_buff = make([]byte, CONN_MAX_BUFF_LEN)
	conn.pack.data = nil
	conn.pack.data_len = 0

	conn.is_login = false

	return conn
}


/* handle login package from client or proxy
 *	
 */
func (c *connect) accept_login() (command , error){
	if !c.is_login {
		c.read_in_package()
		return c.fetch_cmd()
	}
	return nil, errors.New("not login")
}

/*  read in a command from peer
 *
 */
func (c *connect) read_cmd() (command , error){
	c.read_in_package()
	return c.fetch_cmd()
}

/* send a command to peer
 *
 */
func (c *connect) write_cmd (cmd command) error {
	return cmd.send_to_conn(c)
}



/*
* user login package 
* [data_len (2bytes)] [protocol (4bytes)] [version (2bytes)] [role (1byte) = 1] [userid (4|8 bytes)] [token (32 bytes)]
*/
const	ROLE_PROXY	=	"0"
const	ROLE_USER	=	"1"

type pk_user_login struct {
	userid	uint64
	token	string
}

func (p *pk_user_login) init_from_buff (c *connect) error {

	i := 7
	var bytes int
	p.userid, bytes = bytes2num_4_8(c.pack.data[i : ])

	i += bytes
	p.token = string(c.pack.data[i : i+32])

	return nil
}

func (p *pk_user_login) send_to_conn (c *connect) error {
	var buff [64]byte
	if len(p.token) > 32 {
		return errors.New("login token too long")
	}

	copy(buff[2 : 6], []byte("CHAT"))
	i := 6
	copy(buff[i : i+2], c.version)
	i += 2
	copy(buff[i : i+1], ROLE_USER)
	i += 1
	bytes := num2bytes_4_8(buff[i:], p.userid)
	i += bytes
	tl := len(p.token)
	copy(buff[i : i+tl], p.token)
	i += tl
	num2bytes_4_8(buff[0 : 2],  uint64(i - 2))

	c.sock.Write(buff[0 : i])

	return nil
}


/*
* proxy login package
* [data_len (2bytes)] [protocol (4bytes) = "CHAT"] [version (2bytes)] [role (1byte) = 1] [userid (4|8 bytes)] [token (32 bytes)]
*/

type pk_proxy_login struct {
	server	string
	pwd		string
}

func trim_null_characters(bytes []byte) string {
	strlen := len(bytes)
	if strlen == 0 {
		return ""
	}

	begin := 0
	end := strlen
	for i:=0 ; i < strlen && bytes[i] == 0; i++ {
		begin ++
	}
	for j:=strlen - 1 ; j >= 0 && byte[j] == 0; j -- {
		end --
	}

	return string(bytes[begin : end])
}


func (p *pk_proxy_login) init_from_buff(c *connect) error {
	i = 7
	p.server = trim_null_characters(c.pack.data[i : i+32])

	i += 32
	p.pwd = trim_null_characters(c.pack.data[i : i+32])

	return nil
}

func (p *pk_proxy_login) send_to_conn(c *connect) error {
	var buff [128]byte	//[data_len (2bytes)] [version (2bytes)] [role (1byte)] [server (32bytes)] [token (32 bytes)]
	if len(p.token) > 32 {
		return errors.New("wrong proxy login token")
	}

	i := 2
	copy(buff[i : i+2], c.version)
	i += 2
	copy(buff[i : i+1], ROLE_PROXY)
	i += 1
	copy(buff[i : i+32], p.server)
	i += 32
	copy(buff[i : i+32], p.pwd)
	i += 32
	num2bytes_2_8(buff[0 : 2], i - 2)

	c.sock.Write(buff[0 : i])

	return nil

}


/************************* version 1.0 package ********************/
/* 
* heart package  
*
*/
type v1_pk_heartbeat struct {
	userid	uint64
}

func (p *v1_pk_heartbeat) init_from_buff(c *connect) error {
	p.userid = c.login.userid
	return nil
}

func (p *v1_pk_heartbeat) send_to_conn(c *connect) error {
	var buff [32] byte
	num2bytes_2_8(buff[0:], 0)
	_, err := c.sock.Write(buff[0 : 2])
	if err != nil {
		return errors.New("heart beat:" + err.Error())
	}
	return nil
}


/* 
* data package 
* [package_len (2|8bytes)] [receiver (4|8bytes)] [data bytes ...]
*/
const	RECV_TYPE_USER = 1
const	RECV_TYPE_GROUP = 0

type v1_pk_data struct {
	receiver	uint64
	receiver_type	uint8	// RECV_TYPE_USER ,  RECV_TYPE_GROUP
	data	[]byte

	_orig_pack	[]byte		/* original package, used for copying a whole package */
}

func (p *v1_pk_data) init_pack(data []byte, receiver uint64, receiver_type uint8){
	p.receiver = receiver
	p.receiver_type = receiver_type

	data_size := len(data)

	receiver_num := receiver_to_uint64(receiver, receiver_type)

	var tmp_buff [8]byte
	receiver_len := num2bytes_4_8(tmp_buff, receiver_num)

	packsize_len := num2bytes_2_8(tmp_buff, receiver_len + data_size)

	p._orig_pack = make([]byte, packsize_len + receiver_len + data_size)
	p.data = p._orig_pack[packsize_len + receiver_len : ]
}

func (p *v1_pk_data) init_from_buff(c *connect) error {
	receiver, num_bytes = bytes2num_4_8(c.pack.data[0 : ])
	p.receiver, p.receiver_type = uint64_to_receiver(receiver)

	p.data = c.pack.data[num_bytes : c.pack.data_len - num_bytes]

	p._orig_pack = c.read_buff[0 : c.pack.pack_len]

	return nil
}

func (p *v1_pk_data) send_to_conn(c *connect) error {
	bytes, err := c.sock.Write(p._orig_pack)
	if err != nil {
		return errors.New("send data pack:" + err.Error())
	}
	return nil
}


/* fetch command from package 
*
*/
func (c *connect) fetch_cmd() (command, error) {
	if !c.is_login {

		//fetch login command
		if c.pack.data_len != 43 || c.pack.data_len != 47 {
			return nil, errors.New("wrong login package len")
		}

		c.version = string(c.pack.data[4 : 6])	// login: version
		c.role,_ = string(c.pack.data[6 : 7])	//login: role

		switch c.role {

		// user login (  userid, token )
		case ROLE_USER:
			c.login = & c._cmds.user_login
			c.cmd = & c._cmds.user_login


		// proxy login ( server, passwd )
		case ROLE_PROXY:
			c.login = & c._cmds.proxy_login
			c.cmd = & c._cmds.proxy_login

		default:
			panic(P_ERR_WRONG_LOGIN_ROLE)
		}

		c.is_login = true


	} else {
		//fetch normal command
		switch c.version {
		case "10":
			// heartbeat (userid)
			if c.pack.data_len == 0 {
				c.cmd = & c._cmds.v1_pk_heartbeat

				if c.role == ROLE_PROXY {
					panic(P_ERR_PROXY_NOT_ALLOW_HEARTBEAT)
				}

			// send_data (sender<for log>, receiver, receiver_type, data, _orig_pack)
			} else {
				c.cmd = & c._cmds.v1_pk_data

			}

		default:
			panic(P_ERR_WRONG_PROTOCOL_VERSION)

		}
	}
	c.cmd.init_from_buff(c)

	return c.cmd
}



/* read in a whole package from socket into buff
 *
 */
func (c *connect) read_in_package() {

	bytes, err := c.sock.Read(c.read_buff[0 : ] )
	if err != nil {
		panic(P_ERR_READ_SOCK)
	}

	var num_bytes int
	c.pack.data_len , num_bytes = bytes2num_2_8(c.read_buff[0 : 8])
	c.pack.pack_len = num_bytes + c.pack.data_len
	if c.pack.pack_len > CONN_MAX_BUFF_LEN {
		panic(P_ERR_TOO_LARGE_PACKAGE)
	}

	c.pack.data = c.read_buff[num_bytes : c.pack.pack_len]

	for bytes < c.pack.pack_len {
		n, err := c.sock.Read(c.read_buff[bytes : ])
		bytes += n

		if err != nil {
			panic(P_ERR_READ_SOCK)
		}

	}
	c.r_buff_len = bytes

}



/************ to find out "how much bytes need to represent a number" , or "what number is represented by given bytes" *************/
func bytes2num_x_8(buff []byte, fix_len uint8) (uint64, int) {
	if fix_len < 8 {
		var tmp_buff [8]byte

		if buff[0] & 128 {	//10000000, frist bit is '1', then this num is represent by 8 bytes
			copy(tmp_buff[0:8], buff[0:8])
			tmp_buff[0] &= 127	// set the first bit as 0
			return binary.Varuint(tmp_buff[0:8]), 8

		} else {	//00000000, frist bit is '0', then this num is represent by 4 bytes
			copy(tmp_buff[0:fix_len], buff[0:fix_len])
			return binary.Varuint(tmp_buff[0:fix_len]), fix_len
		}
	}
	return 0, -1
}

func num2bytes_x_8(buff []byte, fix_len uint8, number uint64) int {
	max_len := 8
	max_x := uint64(math.Pow(2, fix_len * 8 - 1) - 1)
	max_8 := uint64(math.Pow(2, max_len * 8 - 1) -1)
	if number <= max_x {
		/* fix_len bytes */
		binary.PutUvarint(buff, number)
		return fix_len

	} else if number <= max_8 {
		binary.PutUvarint(buff, number)
		buff[0] |= 128	//10000000, first bit should be '1'
		return max_len

	} else {
		return -1	//error 
	}

}

func bytes2num_4_8(buff []byte) (uint64, int){
	return bytes2num_x_8(buff, 4)
}

func bytes2num_2_8(buff []byte) (number uint64, bytes_len int){
	return bytes2num_x_8(buff, 2)
}

func num2bytes_2_8(buff []byte, number uint64) int {
	return num2bytes_x_8(buff, 2)
}

func num2bytes_4_8(buff []byte, number uint64) int {
	return num2bytes_x_8(buff, 4)
}

func uint64_to_receiver( number uint64 ) (uint64, uint8) {
	if number & 1 {
		return RECV_TYPE_USER, number >> 1
	} else {
		return RECV_TYPE_GROUP, number >> 1
	}
}

func receiver_to_uint64( receiver uint64, receiver_type uint8 ) uint64 {
	receiver <<= 1
	switch receiver_type {
	case RECV_TYPE_USER :
		return receiver | 1

	case RECV_TYPE_GROUP :
		return receiver

	default:
		return 0
	}
}

/*************************************************************************************************************************************/

