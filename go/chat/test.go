package main
import "fmt"
import "regexp"
import "strings"
import	"math"

func main() {
	tmp := "HTTP/1.1 OK\r\n"
	tmp += "Content-type: json\r\n"
	tmp += "Content-data: json\r\n"
	tmp += "Content-format: json\r\n\r\n"
	//tmp += "here we got the http body"
	reg,err := regexp.Compile("^HTTP.*OK(\\r\\n.*)*\\r\\n\\r\\n(.*)$")

	/*
	tmp := "POST /example.html HTTP/1.1\r\n"
	tmp += "Content-type: json\r\n\r\n"
	tmp += "a=1&b=2"
	reg,err := regexp.Compile("^POST\\s+([^\\s]+)\\s+(HTTP[^\\s]+)\\r\\n(.*)\\r\\n\\r\\n(.*)$")
	*/

	str := []byte(tmp)
	fmt.Println(tmp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	match := reg.FindSubmatch(str)
	for _,m := range match {
		fmt.Println("preg fetch: --------------------\n", string(m) )
	}

	params := make(map[string]string)
	for _,chip := range strings.Split("a=1&b=2", "&") {
		p := strings.Split(chip,"=")
		params[p[0]] = p[1]
	}
	fmt.Println(params)

	var buff [4]byte
	buff[0] = 'a'
	buff[1] = 'b'
	buff[2] = 'c'
	buff[3] = 'd'
	newstr := string(buff[0:2])
	buff[0] = 'e'
	fmt.Println(string(buff[:]), newstr)

	var (
		a uint64
		b uint64
	)
	a = uint64(math.Pow(2,63))
	b = uint64(math.Pow(2,63)) - 1

	fmt.Println(a, b, a + b)

}

