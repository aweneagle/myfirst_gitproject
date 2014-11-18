
package main

import	"./runner"
import	"time"
import	"fmt"

type	event struct {
	num	uint64
}

func (e *event) Handle () {
	e.num += 1
}

func test(r *runner.Runner, e func(), num int) {
	begin := time.Now()
	var i int
	for i = 0 ; i < num; i ++ {
		if err := r.Request(e); err != nil {
			break
		}
	}
	end := time.Now()
	fmt.Println(end.Second() - begin.Second(), end.Nanosecond() - begin.Nanosecond(), i)
}

func main () {
	r := runner.Init()
	e := &event{ num:0 }
	var (
		num	int
		times	int
	)

	println("please input test() times:")
	fmt.Scanf("%d", &times)
	println("please input num of each test():")
	fmt.Scanf("%d", &num)

	for i := 0 ; i < times; i ++ {
		go test(r, e.Handle, num)
	}

	for i := 0; i < 10; i ++ {
		go func() {
			time.Sleep(100000 * time.Nanosecond)
			println("quit start")
			if err := r.Quit(); err == nil {
				println("quit success ....")
			} else {
				println("quit err:", err.Error())
			}
		}()
	}

	fmt.Scanln()

}
