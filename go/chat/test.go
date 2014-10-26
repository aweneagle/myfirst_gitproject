package main
import "fmt"

type test_t int

func  main(){
	var a test_t
	a = 1
	fmt.Println("a is ", a)

	b := make (map[test_t]string)
	b[24] = "awen"

	fmt.Println("b is ", b)
}
