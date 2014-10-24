/* version 1.0
 * simple server
 *
 * by awen. 2014.10.24
 */

package main
import	"./web"

func main() {
	defer web.Exception()
	s := web.Server("0.0.0.0", "9999")
	s.Start()
}
