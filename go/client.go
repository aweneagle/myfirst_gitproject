/* version 1.0
 * simple server
 *
 * by awen. 2014.10.24
 */

package main
import	"./web"
import	"fmt"
import	"time"
import	"os"
const buff_size = 64


func main() {
	defer web.Exception()

	if len(os.Args) < 4 {
		fmt.Println("usage: port userid char")
		os.Exit(-1)
	}

	port := os.Args[1]
	userid := os.Args[2]
	char := []byte(os.Args[3])

	c := web.Client("0.0.0.0", port)

	c.StrLogin(userid)

	go func () {
		var msg [buff_size]byte
		padding := []byte("im here to get the right order ....")
		for i:= 0; i < len(padding); i ++ {
			msg[i] = padding[i]
		}

		for i:= len(padding); i<buff_size; i++ {
			msg[i] = char[0]
		}

		for {
			c.Send(msg[0:])	// 1 * 33
			time.Sleep(1 * time.Second)
		}
	}()

	go func () {
		for {
			var msg[512] byte
			total := c.Recv(msg[0:])
			fmt.Printf("[CLIENT] received %s [%d]\n", msg[0:total], total)
		}
	}()

	var input string

	fmt.Scanln(&input)

}
