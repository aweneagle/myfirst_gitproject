package main
import	"fmt"
import	"errors"
func main() {
	a := errors.New("A")
	b := a
	fmt.Println(a == b)
}
