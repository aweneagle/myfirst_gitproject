package chat
import	"strings"
import	"strconv"
import	"regexp"
import	"errors"

const P_ERR_HTTP_READ_ERROR = 1
const P_ERR_HTTP_REQUEST = 2
const P_ERR_HTTP_RESPONSE = 3
const P_ERR_HTTP_ACCEPT_REQUEST = 4

type http struct {
	buff	[1024]byte
	server_name	string
	version	string	//default as HTTP1.1
	sock	*stream
}

type http_request struct {
	uri	string
	params	map[string]string
}


func create_http(c *stream){
	http := new (http)
	http.sock = c
	http.version = "HTTP/1.1"
}

func (h *http) login()


func (h *http) read_request() *http_request {
	var buff [1024] byte
	bytes, err := h.sock.Read(buff[0:])
	if err != nil {
		panic(P_ERR_HTTP_READ_ERROR)
	}

	post_reg,_ := regexp.Compile("^POST\\s+([^\\s]+)\\s+(HTTP[^\\s]+)\\r\\n(.*)\\r\\n\\r\\n(.*)$")
	get_reg,_ := regexp.Compile("^GET\\s+([^\\s]+)\\s+(HTTP[^\\s]+)\\r\\n(.*)\\r\\n\\r\\n(.*)$")
	match := post_reg.FindSubmatch(buff[0 : bytes])

	req := new (http_request)
	req.params = make(map[string]string)
	if match == nil {
		match = get_reg.FindSubmatch(buff[0 : bytes])
		if match == nil {
			panic(P_ERR_HTTP_ACCEPT_REQUEST)
		}

		path := strings.Split(string(match[1]), "?")
		req.uri = path[0]
		if len(path) == 2 {
			for _,param := range strings.Split(string(path[1]), "&") {
				p := strings.Split(param, "=")
				if len(p) == 2 {
					req.params[p[0]] = p[1]
				}
			}
		}

	} else {
		req.uri = string(match[1])
		for _,param := range strings.Split(string(match[3]), "&") {
			p := strings.Split(param, "=")
			if len(p) == 2 {
				req.params[p[0]] = p[1]
			}
		}
	}
	return req

}

func (h *http) write_response(body string){
	content_len := strconv.FormatUint(uint64(len(body)), 10)
	resp := h.version + " 200 OK\r\n"
	resp += "Content-type: text/json\r\n"
	resp += "Content-len: " + content_len + "\r\n"
	resp += "Server: chat\r\n\r\n"
	resp += body

	buff := []byte(resp)
	_, err := h.sock.Write(buff)
	if err != nil {
		panic(P_ERR_HTTP_RESPONSE)
	}
}

func (h *http) end(){
	h.sock.Close()
}

/* send a http request , and recv a response
*/
func (h *http) request(uri string) (string, error){
	req := "GET " + uri + " HTTP/1.1\r\n"
	req += "Host: unknown\r\n"
	req += "\r\n\r\n"

	buff := []byte(req)
	_, err := h.sock.Write(buff)
	if err != nil {
		panic(P_ERR_HTTP_REQUEST)
	}

	var r_buff [4096]byte
	var bytes int
	bytes, err = h.sock.Read(r_buff[0:])
	if err != nil {
		return "", err
	}

	reg, _ := regexp.Compile("^HTTP.*OK(\\r\\n.*)*\\r\\n\\r\\n(.*)$")

	match := reg.FindSubmatch(r_buff[0:])
	if len(match) != 3 {
		return "", errors.New(string(r_buff[:]))
	}
	return string(match[2]), nil
}
