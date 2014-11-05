/* server of chat 
 *
 * by awen, 2014.10.30
 */

package chat

import	"net"
import	"strconv"


type client struct {
 conn *connect
 http *http

 host	string
 port	uint64

}


/* create a new client
 * 
 * @param	host,	server host
 * @param	port,	server port
 */
func NewClient(host string, port uint64) (*client, error) {
	cli := new (client)
	sock,err := net.Dial("tcp", host + ":" + strconv.FormatUint(port, 10))
	if err != nil {
		return nil, err
	}

	cli.conn = create_connect(create_stream(&sock))
	cli.host = host
	cli.port = port
	return cli, nil
}


/* login as a user 
 *
 * @param	userid
 * @param	passwd
 */
func (c *client) UserLogin(userid uint64, passwd string) error {
	var login pk_user_login
	login.userid = userid
	login.token = passwd
	return c.conn.write_cmd(&login)
}


/* send message to user
 *
 */
func (c *client) SendToUser(msg string, receiver uint64) error {
	var data v1_pk_data
	data.init_pack([]byte(msg), receiver, RECV_TYPE_USER)
	return c.conn.write_cmd(&data)
}


/* send message to group 
 */
func (c *client) SendToGroup(msg string, receiver uint64) error {
	var data v1_pk_data
	data.init_pack([]byte(msg), receiver, RECV_TYPE_GROUP)
	return c.conn.write_cmd(&data)
}


/* read message from server 
 */
func (c *client) Read( message []byte ) uint64 {
	c.conn.read_buff = message
	c.conn.read_in_package()
	return c.conn.r_buff_len
}
