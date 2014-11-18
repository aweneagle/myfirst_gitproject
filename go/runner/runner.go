/* runner
 * 
 * 创建一个routine专门用于接受/处理请求, 对需要同步进行的func进行排队处理
 *
 */

package	runner

import	"sync/atomic"
import	"errors"
import	"time"

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
func NewMutx(ca int32) *Mutx {
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
	} else if new_m <= m.ca && m.num >= 0 {
	/*
	if m.num >= 0 && new_m <= m.ca {
		return atomic.AddInt32(&m.num, 1)
		*/
		/* 系统繁忙 */
		return MUTX_ERR_SYS_BUSY
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
	if num := atomic.AddInt32(&m.num, -1) ; num < 0 {
		return MUTX_ERR_CLOSED
	} else {
		return num
	}
}

/* 判断信号量是否已经关闭
 */
func (m *Mutx) IsClosed() bool {
	return m.num < 0
}



type Runner struct {
	/* Quit 信号量 */
	q_mutx	*Mutx
	/* Request 信号量 */
	r_mutx	*Mutx

	/* 终止处理器
	*/
	quit	chan bool

	/* 终止完成
	 */
	quit_done	chan bool

	/* 紧急终止 */
	to_quit	bool

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
	p.r_mutx = NewMutx(65535)
	p.q_mutx = NewMutx(1)
	p.event_in = make(chan Event)
	p.event_out = make(chan Event)
	p.quit = make(chan bool)
	p.quit_done = make(chan bool)
	p.to_quit = false
	go p.run()
	return p
}


/* 向该处理器发送一个请求 , 并等待, 直到处理器返回结果
 *
 */
func (p *Runner) Request (in Event) error {
	var (
		m int32
		slp uint64 = 1
	)
	if p.to_quit {
		return errors.New("runner is shutdown")
	}
	for {
		if m := p.r_mutx.Up(); m > 0 {
			//println("sending request")
			p.event_in <- in
			//println("sending result")
			_ = <-p.event_out
			//println("return result...")
			p.r_mutx.Down()

			/* 最后一个离开 Request 的 routine需要通知 Runner 可以关闭服务了 */
			if p.r_mutx.IsClosed() {
				p.quit <- true
			}
			return nil
		}
		if m == MUTX_ERR_CLOSED {
			return errors.New("runner closed")

		} else if m == MUTX_ERR_CAP_FULL || m == MUTX_ERR_SYS_BUSY {
			/* sleep for a few nanoseconds */
			//return nil
			time.Sleep( time.Duration(slp) * time.Nanosecond)
			slp += 1
			if slp > 128 {
				slp = 1
			}
		} else {
			return errors.New("sys busy")
		}
	}
}


/* 终止处理器
 */
func (p *Runner) Quit() error {
	/* 只允许一个routine向 p.quit 发送 信号 */
	if p.q_mutx.Up() > 0 {

		/* 关闭Request, 拒绝后续的请求 */
		p.to_quit = true

		/* 关闭mutx */
		p.r_mutx.Down()

		/* 如果已经没有其他线程在使用服务，可以直接关闭 */
		if p.r_mutx.IsClosed() {
			p.quit <- true
		}
		/* 交给最后一个离开Request的线程执行 p.quit <- true, 见 func (p *Runner) Request (in Event) error */
		_ = <-p.quit_done
		return nil
	}
	return errors.New("runner is shutting down")
}


func (p *Runner) run () {
	for {
		select {

			/* 收到终止信号 */
		case _ = <-p.quit:
			println("quiting done...")
			close(p.event_in)
			close(p.event_out)
			close(p.quit)
			select {
			/* 通知未离开 Quit 的线程 */
			case p.quit_done <- true:
				/* do nothing */
			}
			return

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

