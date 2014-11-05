package chat
import	"net"
const P_ERR_READER_SIZE_TOO_LARGE = 1

const	STREAM_BUFF_SIZE = 128
type stream struct {
	sock	*net.Conn	//tcp socket
	buff	[STREAM_BUFF_SIZE]byte
	buff_len	int
}

func (r *stream) create_stream(c *net.Conn) {
	r = new stream
	r.buff_len = 0
	r.sock = c
	return r
}

func (r *stream) Read (buff []byte) int, error{
	if r.buff_len == 0 {
		return r.sock.Read(buff)
	} else {
		buff_read_len := copy(buff[0, r.buff_len], r.buff[0, r.buff_len])
		sock_read_len := r.sock.Read(buff[r.buff_len : ])
		r.buff_len = 0
		return sock_read_len + buff_read_len, nil
	}
}

func (r *stream) Write(buff []byte) int, error {
	return r.sock.Write(buff)
}

func (r *stream) Close() error {
	return r.sock.Close()
}

func (r *stream) push_back(buff []byte) {
	if len(buff) > STREAM_BUFF_SIZE {
		panic(P_ERR_READER_SIZE_TOO_LARGE)
	}
	r.buff_len = copy(r.buff[0:], buff)
}
