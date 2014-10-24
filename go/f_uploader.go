/* version 1.0
 * simple server
 *
 * by awen. 2014.10.24
 */

package main
import	"./web"
import	"fmt"
import	"os"
import	"io"
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

	fi, err := os.Open(file)
	web.CheckError(err)

	buf := make([]byte, buff_size)

	for {
		n, err := fi.Read(buf)

		if err != nil && err != io.EOF {
			panic("failed while reading ...")
		}

		if n == 0 {
			break;
		}

		c.Send(buf[0:n])
	}
	fi.Close()
	fmt.Println("succ!")
	os.Exit(0)

}
