/* version 1.0
 * 
 * a chat server, which can run on serval servers at the same time
 * when server start,  clients need to do follow things to comunicate with each other:
 *	1. login on server with a `userid`
 *	2. join into a chat `group`
 *	3. send message to this group
 *	4. other in this group will receive messages sent in step 3.
 *	
 *
 * by awen. 2014.10.26
 */

package chat

import "os"
import "runtime"
import "fmt"
import "strings"
import "strconv"


// `uid` is short for `userid`
// `gid` is short for `groupid`

type	uid_t	int32		// userid type //
type	gid_t	int32		// group id type //
type	size_t	int64		// size type //
type	cmd_t	int8		// command type 



type server struct {
	// users and groups //
	online map[uid_t] *connect				// online users , uid => *connect
	online_token map[uid_t] string				// online users' token,  uid => tokens 
	online_status map[uid_t] map[string]size_t		// online users' status, uid => {"sent":1, "recv":2, ...} //
	groups map[gid_t] map[uid_t] bool			// group members //

	on_other_line map[uid_t] string				// users on other peer servers, uid => host+port //

	peer_servers_proxy map[string] *connect			// peer servers' message proxy channels,  host+port => connect 
	peer_servers_cmd  map[string] *http			// peer servers' http command channels,	  host+port => http 

}

/*start a server*/
func (s *server) Start() {}



/*
 *  message protocol :
 *	the first package must be login package , and the others are all data packages
 *	[login package][data package][data package] ...
 *
 *	login package 
 *		[version : 2bytes][userid : 4/8 bytes][token : 32 bytes]
 *
 *	data package 
 *		user -> group
 *		[len : 1/4/8 bytes][sender : 4/8 bytes][receiver : 4/8 bytes(bid[2] == 1))][data : x bytes]
 *		user -> user
 *		[len : 1/4/8 bytes][sender : 4/8 bytes][receiver : 4/8 bytes(bid[2] == 0))][data : x bytes]
 *
 */
const SRV_CONN_BUFF_SIZE = 1024


type  connect struct {
	sock *net.Conn
	buff	[SRV_CONN_BUFF_SIZE]byte			// data buffer 
	buff_len	size_t					// current buffer length
	buff_is_empty	chan bool				// to inform writer "the buff data has been sent out succefully"
}

/* read data from socket, then write it into empty buff , and then wait writer routine to read it out  until buff is empty again*/
func (c *connect) read_in() {}

/* copy data from buff , then send them to all users, and finaly inform reader routine to read new data into buff again */
func (c *connect) write_out(){}






type  http struct {
	sock *net.Conn
	cmds map[string] cmd_t		//all commands definition
	keep_alive	bool		//if or not to keep sock alive after a http request is down
}

/* read in a http uri */
func (h *http) read_request() string {}

/* build a command from a http uri */
func (h *http) build_command(uri string) command {}

/* write back response */
func (h *http) write_response(respose string) {}

/* do a http query and return result as a string */
func (h *http) query (uri string) {}





type command struct {
	op	cmd_t		//command operation
	params	map[string]string	//params of a command
	result	string		//result after this command execute
}

/* execute a command and return result as a string */
func (c *connect) exec (s *server) string {}
