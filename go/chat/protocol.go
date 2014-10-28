/* protocols of chat server
 *
 *	client <-> server,		see		read_login_package(), read_data_from_client(), write_data_to_client()
 *
 *	proxy_server <-> server,	see	read_login_package(), read_data_from_proxy(), write_data_to_proxy()
 *
 *
 * by awen , 2014.10.28
 */


package chat

import "encoding/binary"

/* read package 
 * package struct:
 *		[package_len (2|8 bytes)][data bytes ...]
 *
 */
 func read_package (c *net.Conn, pack *data) bytes uint64 {

	 /* why data-buff write begin with 8 ?
	  *   "to resove a space for some additional information so that we will not need to copy or 
	  * move the whole data buffer memory when the "data frame" is changed"
	  */

	 /* read [package_len (2|8 byte)] */
	 _, err := c.Read(pack.buff[8:16]);
	 if err != nil {
		 panic (err.Error())
	 }

	 pack_len, bytes := get_number_from_bytes_2_8(pack.buff[8:16])

	 /* 8 byte space resoved for additional information , 8 bytes for [package_len] itself */
	 if pack_len + 8 + 8 > SRV_CONN_BUFF_SIZE {
		 panic ("too large package")
	 }

	 pack.bytes = pack_len

	 /* small package */
	 if pack.bytes <= 8 {
		 return
	 }

	 c.Read(pack.buff[8:


 }

/* login package , the first data package we must received when connect established succefully.
 * it's sent from client side to server side
 * package struct:
 *		[package_len(2|8bytes)] [version (3byte)] [role (1byte)] [userid (4|8 bytes)] [token (32 bytes)]
 *
 *		version:	1.0.1		each bytes represents an integer  
 *		role:		1			ROLE_PROXY_CLIENT, ROLE_CLIENT
 *		userid:		1211		if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer
 *		token:		0afccbdedf...	a 32-bytes md5 string
 *
 * when wrong data received , function will panic 
 *
 */

func read_login_package (pack *data) version int32, role int8, userid uint64, token string {
}







/************************* comunications betweem client and server ***********************/

type receiver struct {
	id		uint64			// userid or groupid
	id_type	uint4			// id type, 1: userid, 2:group_id
}

type data struct {
	buff []byte
	bytes uint64
}

/* read a whole data package from client
 * package struct :
 *		[package_len (2|8 bytes)] [receiver (4|8 bytes)] [data bytes]
 *
 *		receiver:	1200		if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer
 *		data_len:	121			if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer (with the first bit turned to 0)
 *		data:		bytes
 *
 */
func read_data_from_client(pack *data, data *data, receiver *receiver) {
}

func write_data_to_server(pack *data, data *data, receiver *receiver) {
}


/* write a whole data package to client 
 * package struct :
 *		[package_len (2|8 bytes)] [sender (4|8 bytes)] [receiver (4|8 bytes)] [data bytes]
 *
 *		receiver:	1201		if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer
 *		sender:		1200		if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer
 *		data_len:	121			if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer (with the first bit turned to 0)
 *		data:		bytes
 */
func write_data_to_client(pack *data, data *data, sender uint64, receiver *receiver) {
}

func read_data_from_server(pack *data, receiver *receiver) bytes uint64, sender uint64{
}



/************ to find out "how much bytes need to represent a number" , or "what number is represented by given bytes" *************/

func get_number_from_bytes_2_8(buff []byte) number uint64, bytes_len int{
}

func get_number_bytes_len_2_8(number uint64) bytes_len int{
}



func get_number_from_bytes_4_8(buff []byte) number uint64, bytes_len int{
}

func get_number_bytes_len_4_8(number uint64) bytes_len int{
}



func put_number_into_bytes(buff []byte, bytes int, number uint64){
}

/*************************************************************************************************************************************/









/************************* comunications betweem proxy and server ***********************/

/* read a whole data package from proxy server 
 * package struct :
 *		[package_len (2|8 bytes)] [sender (4|8 bytes)] [receiver (4|8 bytes)] [data (`data_len` bytes)]
 *
 *		receiver:	1201		if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer
 *		sender:		1200		if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer
 *		data_len:	121			if first bit is 0, it is a 4-bytes integer; or it is a 8-bytes integer (with the first bit turned to 0)
 *		data:		bytes
 */

func read_data_from_proxy(pack *data, receiver *receiver) bytes uint64, sender uint64 {
}

func write_data_to_proxy(pack *data, sender uint64, receiver *receiver) {
}
