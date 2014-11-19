package	main
import	"./netevent"
import	"fmt"
import	"time"
import	"sync/atomic"

func test(fd uint32, ne *netevent.NetEvent, num int, psize int) uint64 {
	var (
		//psize	int = 1024
		cost	uint64 = 0
		sent	uint64 = 0
		succ	int32 = 0
		fail	int32 = 0

		total	int32	= 0

		done	chan	bool

		err	error

		bytes	int = 0

		pack	[]byte
		f	func(*uint64, *uint64, *int32, *int32)
		test_begin	time.Time
		test_end	time.Time
		test_cost	uint64	= 0
	)

	fmt.Println("connect connect:", fd)
	pack = make([]byte, psize)
	done = make(chan bool)
	for i := 0; i < psize; i ++ {
		pack[i] = byte('a')
	}

	cost = 0
	f = func(cost *uint64, sent *uint64, succ *int32, fail *int32) {
		var (
			begin	time.Time
			end	time.Time
		)
		begin = time.Now()
		bytes, err = ne.Send(fd, pack)
		end = time.Now()

		if err != nil {
			atomic.AddInt32(fail, 1)
		} else {
			atomic.AddInt32(succ, 1)
			atomic.AddUint64(sent, uint64(bytes))
		}
		atomic.AddUint64(cost, (uint64(end.Second()) - uint64(begin.Second())) * 1000000000)
		atomic.AddUint64(cost, uint64(end.Nanosecond()) - uint64(begin.Nanosecond()))

		if atomic.AddInt32(&total, -1) <= 0 {
			done <- true
		}
	}

	total = int32(num)
	test_begin = time.Now()
	for i := 0; i < num; i ++ {
		go f(&cost, &sent, &succ, &fail)
	}

	_ = <-done
	test_end = time.Now()
	test_cost += (uint64(test_end.Second()) - uint64(test_begin.Second())) * 1000000000
	test_cost += uint64(test_end.Nanosecond()) - uint64(test_begin.Nanosecond())

	println( "num:", num, "cost:", cost, "sent:", sent, "succ:", succ, "fail:", fail, "test_cost:", test_cost )

	return test_cost
}

func main() {
	var	(
		num	int
		bytes	int
	)
	fmt.Println("please input bytes:")
	fmt.Scanf("%d", &bytes)
	fmt.Println("please input routine:")
	fmt.Scanf("%d", &num)

	ne := netevent.Init()
	ne.Debug = true;
	ne.OnConn = func(fd uint32) error {
		var	total uint64 = 0
		for i := 0; i < 10 ; i ++  {
			total += test(fd, ne, num, bytes)
		}
		println("total:" , total)
		println("avarage:" , total/10)
		return nil
	}
	ne.OnClose = func(fd uint32) error {
		ne.Close(fd)
		fmt.Println("connect closed:", fd)
		return nil
	}
	ne.OnRecv = func(fd uint32, pack []byte) error {
		return nil
	}

	ne.Listen("127.0.0.1:8888")


}
