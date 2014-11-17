/* net event driver 
 *
 * 微型网络事件驱动
 *
 * by awen, 2014.11.17
 */

import	"../runner"
import	"net"

type connect struct {

	/* 网络事件处理器 */
	runner	*runner.Runner

	/* socket */
	sock	*net.Conn

}

/* 启动守护 routine, 监听网络socket, 并获取数据
 *
 */
func (c *connect) Run () {
	var r = &read_data{ conn:c }
	for {
		c.runner.Request(r)
	}
}

/* 发送请求，并等待响应 */
func (c *connect) Request(request []byte, response []byte) error {
	var r = &send_data{ conn:c }
	return c.runner.Request(r)
}

/* 关闭连接 */
func (c *connect) Close() {
	var r = &close_conn{ conn:c }
	return c.runner.Request(r)
}


/* 关闭网络连接 */
type close_conn struct {
	conn	*connect
}

func (s *close_conn) Handle() {
}


/* 向 网络连接 发送数据 */
type send_data struct {
	conn	*connect
}

func (s *send_data) Handle() {
}



/* 从 网络连接 中读取数据 */
type read_data struct {
	conn	*connect
}

func (s *read_data) Handle() {
}
