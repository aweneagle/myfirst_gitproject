package main
import "fmt"
import "regexp"
import "strings"

type a struct {
	m map[string]string
}

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

	a := new (a)
	if a.m == nil {
		fmt.Println("nil found")
	}


}

