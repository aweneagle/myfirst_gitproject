/* version 1.0
 * simple server
 *
 * by awen. 2014.10.24
 */

package main
import	"./web"
import	"fmt"
import	"os"
const buff_size = 64


func main() {
	defer web.Exception()

	if len(os.Args) < 4 {
		fmt.Println("usage: port userid file")
		os.Exit(-1)
	}

	port := os.Args[1]
	userid := os.Args[2]
	file := os.Args[3]

	c := web.Client("0.0.0.0", port)

	c.StrLogin(userid)

	fi, err := os.Create(file)
	web.CheckError(err)

	buf := make( []byte, buff_size )

	for {
		bytes := c.Recv(buf[0:])

		_, err := fi.Write(buf[0:bytes])

		web.CheckError(err)

	}

	fi.Close()

	os.Exit(0)
}
