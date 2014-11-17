/* runner
 * 
 * 创建一个routine专门用于接受/处理请求, 对需要同步进行的func进行排队处理
 *
 */

package	runner

import	"sync/atomic"
import	"errors"

/* 有限容量的信号量 
 * 
 * 一个信号量共有三个操作：
 *   Up()	增加信号量 num += 1, 直到容量已满
 *   Down()	减少信号量 num -= 1
 */
type Mutx struct {
	num	int32	//当前信号量
	ca	int32	//容量
}

/* 初始化一个信号量 
 */
func Mutx(ca int32) *Mutx {
	return  &Mutx{ ca:ca, num:0 }
}

const	MUTX_ERR_CLOSED = -1
const	MUTX_ERR_SYS_BUSY = -2
const	MUTX_ERR_CAP_FULL = -3

/* up 增加信号量，直至达到容量 m.ca 为止
 * 
 */
func (m *Mutx) Up () int32 {
	new_m := m.num + 1
	if m.num >= 0 && new_m <= m.ca && atomic.CompareAndSwapInt32(&m.num, m.num, new_m) {
		return new_m
	}
	if new_m <= m.ca {
		/* 系统繁忙 */
		return MUXT_ERR_SYS_BUSY
	} else if m.num < 0 {
		return MUTX_ERR_CLOSED
	} else {
		/* 容量已达到上限 */
		return MUTX_ERR_CAP_FULL
	}
}

/* down 减少信号量
 */
func (m *Mutx) Down() int32 {
	if atomic.AddInt32(&m.num, -1) < 0 {
		return MUTX_ERR_CLOSED
	}
}

/* 判断信号量是否已经关闭
 */
func (m *Mutx) IsClosed() {
	return m.num < 0
}



type Runner struct {
	/* Request 信号量 */
	r_mutx	*Mutx
	/* 有多少 routine 在执行 Quit */
	quitting int32

	/* 有多少 routine 在执行 Request*/
	running	int32

	/* 终止处理器
	*/
	quit	chan bool

	/* 关闭 Request */
	shutdown	bool

	/* 仍留在 Request() 函数内的 routine，通过 final_quit 告诉 runner 可以关闭服务了 
	 *
	 */
	final_quit	chan bool

	/* 通知 请求routine 数据已经处理完，可以取了
	 */

	/* 接受 请求routine 的请求
	 */
	event_in	chan Event

	/* 返回 请求结果
	*/
	event_out	chan Event

}

type Event interface {
	/* 事件处理 */
	Handle ()
}

/* 初始化一个处理器, 并启动它 */
func Init () *Runner {
	p := &Runner{}
	p.r_mutx = Mutx(65535)
	p.q_mutx = Mutx(1)
	p.event_in = make(chan Event)
	p.event_out = make(chan Event)
	p.quit = make(chan bool)
	p.final_quit = make(chan bool)
	p.running = 0
	p.quitting = 0
	p.shutdown = false
	go p.run()
	return p
}


/* 向该处理器发送一个请求 , 并等待, 直到处理器返回结果
 *
 */
func (p *Runner) Request (in Event) error {
	if p.r_mutx.Up() > 0 {
		//println("sending request")
		p.event_in <- in
		//println("sending result")
		_ = <-p.event_out
		//println("return result...")
		p.r_mutx.Down()

		/* 最后一个离开 Request 的 routine需要通知 Runner 可以关闭服务了 */
		if p.shutdown == true && p.r_mutx.IsClosed() {
			p.quit <- true
			//println("final quitting start...")
		}
		//println("leaving Requesst ...", p.running)
		return nil
	}
	return errors.New("sys busy")
}


/* 终止处理器
 */
func (p *Runner) Quit() error {
	atomic.AddInt32(&p.quitting, 1)
	/* 只允许一个routine向 p.quit 发送 信号 */
	if p.quitting == 1 {

		/* 关闭Request, 拒绝后续的请求 */
		p.shutdown = true

		atomic.AddInt32(&p.running, -1)
		//println("quiting[", p.running, "]")
		p.quit <- true
		return nil
	}
	return errors.New("runner is shutting down")
}


func (p *Runner) run () {
	for {
		select {

			/* 收到来自 Request 中的 routine 的终止信号 */
		case _ = <-p.final_quit:
			println("quiting done...", p.running)
			close(p.event_in)
			close(p.event_out)
			close(p.quit)
			return


			/* 收到来自 Quit 的终止信号 */
		case _ = <-p.quit:

			/* 还有routine未离开Request函数 */
			if p.running >= 0 {
				/* 最后一个离开Request函数的 routine 发来通知*/
				//println("final quitting done...")
				break

			} else {
				println("quiting done...", p.running)
				close(p.event_in)
				close(p.event_out)
				close(p.quit)
				return
			}

			/* 收到请求，处理*/
		case e := <-p.event_in:
			//println("recv request")
			e.Handle()
			//println("handle request")
			p.event_out <- e
			//println("return result")

		default:
		}


	}
}

