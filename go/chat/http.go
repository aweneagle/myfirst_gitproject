package chat


type http struct {
	server	string	// host":"port
	sock	*net.Conn
	keep_alive	bool	//whether to keep alive or not
}

type http_request interface {
}

type group_join struct {
	gid uint64
	uid	uint64
}

type group_leave struct {
	gid uint64
	uid	uint64
}

type group_on_server struct {
	gid uint64
	host	string
	port	string
}

type group_off_server struct {
	gid	uint64
	host	string
	port	string
}

type user_login struct {
	uid	uint64
	host	string
	port	string
}

type user_logout struct {
	uid	uint64
}

func (h *http) read_request() http_request {
}

func (h *http) write_response(resp string){
}

func (h *http) end(){
	h.sock.Close()
}

/* send a http request , and recv a response
*/
func request(uri string) {
}


func create_http(c *net.Conn){
	http := new (http)
	http.sock = c
	keep_alive = true
}



