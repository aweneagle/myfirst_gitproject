<?php
namespace Tools;
use Exception;
use Closure;

/*
 * 模拟一个httpserver来返回数据，供phpunit测试
 */
class FakeHttpServer
{
    const DEAMO = 1;
    const NORMAL = 2;


    private $to_clean = true;
    private $post_handlers = [];
    private $get_handlers = [];

    public function post($path, Closure $func)
    {
        $this->post_handlers[$path] = $func;
    }

    public function get($path, Closure $func)
    {
        $this->get_handlers[$path] = $func;
    }

    public function request($path, Closure $func)
    {
        $this->post_handlers[$path] = $this->get_handlers[$path] = $func;
    }

    private function get_handler($method, $path)
    {
        if ($method == "GET") {
            $handlers = $this->get_handlers;
        } else {
            $handlers = $this->post_handlers;
        }
        if (isset($handlers[$path])) {
            return $handlers[$path];
        } else {
            return false;
        }
    }

    public $debug = true;

    public $default_response_headers = [
        "HTTP/1.0 {status}",
        "Server: nginx/1.6.2",
        "Vary: Accept-Encoding",
    ];

    private $host;
    private $port;

    private $rule = [];

    /* 是否在监听中 */
    protected $listening = false;

    /* 子进程PID */
    private $child = null;

    /* 子进程结束时的状态 */
    private $child_status = null;

    /*
     * unix socket 消息通道
     */
    protected $channel = null;

    public function __construct($host, $port)
    {
        $this->host = $host;
        $this->port = $port;
    }

    /**
     * response() 根据条件返回数据
     *
     * @param   array $when, 请求需要满足的条件  [$field => $preg]
     *              $field: 'post', 'get', 'header', 'body', 'url'
     *              $preg:  正则表达式
     *
     * @param   array $return, 指定要返回数据, [$field => $data]
     *              $field: 'header', 'body'
     *              $data: 字符串
     *              
     */
    public function response(array $when, array $return)
    {
        $this->rule[] = [
            'when' => $when,
            'return' => $return,
        ];
    }

    protected function recv()
    {
        return trim(fgets($this->channel));
    }

    protected function send($msg)
    {
        if ($this->debug) {
            echo $msg . "\n";
        }
        fwrite($this->channel, $msg . "\n");
    }

    protected function halt($errmsg)
    {
        $this->send($errmsg);
        exit(1);
    }

    public function quit($signal)
    {
        $this->listening = false;
        if ($this->debug) {
            echo "signal [$signal] received \n";
        }
    }

    private function debug($msg)
    {
        if ($this->debug) {
            echo "child[{$this->child}] $msg\n";
        }
    }

    public function start($style = self::NORMAL)
    {
        if ($style == self::DEAMO) {
            //父进程创建一个子进程, 并退出
            $this->debug = false;
            if (($this->child = pcntl_fork()) === -1) {
                throw new Exception("failed to call pcntl_fork()");
            }
            if ($this->child == 0) {
                //子进程
                sleep(1);//等待父进程退出
                $this->serve();
            } else {
                //父进程
                echo "deamo is created succefully\n";
                $this->to_clean = false;
                exit(0);
            }
        } else {
            $this->start();
        }
    }

    /**
     * start() 启动server
     *
     * @return false, 启动失败;  true, 启动成功
     */
    private function serve()
    {
        //创建通信通道
        $pipe = stream_socket_pair(STREAM_PF_UNIX, STREAM_SOCK_STREAM, STREAM_IPPROTO_IP);
        if (!$pipe) {
            throw new Exception("failed to call stream_socket_pair()");
        }
        //父进程创建一个子进程, 并等待
        if (($this->child = pcntl_fork()) === -1) {
            throw new Exception("failed to call pcntl_fork()");
        }

        // 子进程
        if ($this->child == 0) {
            $this->channel = $pipe[0];
            stream_set_blocking($this->channel, 1);
            //子进程监听端口
            $ln = socket_create(AF_INET, SOCK_STREAM, 0);
            if (!$ln) {
                //通知父进程“失败",即将退出
                $this->halt("failed to call socket_create()");
            }
            //注册退出监听函数
            declare(ticks = 1);
            if (!pcntl_signal(SIGUSR2, [$this, "quit"])) {
                $this->halt("failed to set signal handler pcntl_signal()");
            }
            socket_setopt($ln, SOL_SOCKET, SO_REUSEADDR, 1);
            if (!socket_bind($ln, $this->host, $this->port)) {
                //通知父进程“失败",即将退出
                $this->halt("failed to call socket_bind()");
            }
            if (!socket_listen($ln)) {
                //通知父进程“失败",即将退出
                $this->halt("failed to call socket_listen()");
            }
            if (!socket_set_nonblock($ln)) {
                //通知父进程“失败",即将退出
                $this->halt("failed to call socket_set_noblock()");
            }

            //子通知父进程成功
            $this->listening = true;
            $this->send("OK");
            while ($this->listening) {
                $conn = socket_accept($ln);
                if ($conn === false) {
                    usleep(5000);
                    continue;
                }
                if ($conn == false) {
                    $this->debug("child: failed to call socket_accept()");
                    $this->listening = false;
                    break;
                }
                $this->handle($conn);
            }
            socket_close($ln);
            exit(0);

        } else {
            //父进程
            $this->channel = $pipe[1];
            stream_set_blocking($this->channel, 1);
            $child_msg = $this->recv();
            if ($child_msg != "OK") {
                //如果子进程失败， 父进程,等待子进程结束，然后抛异常
                pcntl_waitpid($this->child, $this->child_status);
                throw new Exception("failed: child [" . $this->child . "] " . $child_msg);

            } else {
                //如果子进程成功， 父进程继续
                $this->debug("child ready");
                return;
            }
        }
    }

    /*
     * 处理连接，返回http包
     */
    protected function handle($socket)
    {
        socket_getpeername($socket, $raddr, $rport); 
        $this->debug("Received Connection from $raddr:$rport"); 

        $body = null;
        $headers = [];
        $state = "head";
        $left_str = "";
        $len = 0;
        $body_len = 0;
        $path = '/';
        $method = null;
        $query = null;

        //连接是否关闭 (连接关闭时，socket_read() 返回 "")
        while ($str = socket_read($socket, 1024, PHP_BINARY_READ)) {
            if ($state == "body") {
                $body = $body . $str;
            } else {
                $str = $left_str . $str;
                $len = strlen($str);
                $pos = 0;
                for ($i = 0; $i < $len; $i ++) {
                    if ($str[$i] == "\r" && $str[$i+1] == "\n") {
                        //每个header后面跟了一个\r\n
                        $header = $headers[] = substr($str, $pos, $i - $pos);
                        //确定body的长度
                        if (preg_match('/^content-length:\s(\d+)*$/i', $header, $match)) {
                            $body_len = intval($match[1]);
                        }
                        //确定path
                        if (preg_match('/^(GET|POST|HEAD)\s+(\/.*)\s/', $header, $match)) {
                            $method = $match[1];
                            $path = $match[2];
                            if ($qpos = strpos($path,"?")) {
                                $query = substr($path, $qpos + 1);
                                $path = substr($path, 0, $qpos);
                            }
                        }
                        $pos = $i + 2;
                        //\r\n\ header 和 body 之间多了一个 \r\n
                        if ($str[$pos] == "\r" && $str[$pos+1] == "\n") {
                            $state = "body";
                            $pos = $pos + 2;
                            $body = substr($str, $pos);
                            $pos = strlen($str);
                            break;
                        }
                    }
                }

                if ($pos < $len) {
                    $left_str = substr($str, $pos);
                }
            }

            //检查http包是否已经完整接收到了
            if ($body_len <= strlen($body)) {
                if (($handler = $this->get_handler($method, $path))) {
                    //处理请求:
                    $req = new ____Request____;
                    $resp = new ____Response____;
                    $req->method = $method;
                    $req->headers = $headers;
                    $req->path = $path;
                    parse_str($body, $req->POST);
                    parse_str($query, $req->GET);

                    try {
                        $handler($req, $resp);
                        //响应数据包
                        $headers = $this->default_response_headers;
                        //写入 Status 
                        $headers[0] = str_replace("{status}", $resp->status, $headers[0]);
                        $http = implode("\r\n", $headers) . "\r\n\r\n" . $resp->body;
                        if (!socket_write($socket, $http, strlen($http))) {
                            $this->send("child: failed to call fwrite() on connection");
                        }
                    } catch (Exception $e) {
                        $this->send("child: " . $e->getMessage());
                    }

                } else {
                    $this->send("child: unknown request");
                }
                socket_shutdown($socket);
                break;
            }
        }
        socket_close($socket);
    }

    private function select_response($url, array $headers, $body)
    {
        return $this->rule[0];
    }

    private function match($preg, array $contents)
    {
        foreach ($contents as $c) {
            if (!preg_match($preg, $c)) {
                return false;
            }
        }
        return true;
    }

    private function clean()
    {
        if ($this->channel) {
            fclose($this->channel);
            $this->channel = null;
        }
    }

    public function stop()
    {
        posix_kill($this->child, SIGQUIT);
        while ($msg = $this->recv()) {
            $this->debug($msg);
        }
        pcntl_waitpid($this->child, $this->child_status);
        $this->child = null;
    }

    public function __destruct()
    {
        if (!$this->to_clean) {
            return;
        }
        $this->clean();
        if ($this->child) {
            pcntl_waitpid($this->child, $this->child_status);
        }
    }
}

class ____Request____ {
    public $path = null;
    public $headers = [];
    public $POST = [];
    public $GET = [];
}

class ____Response____ {
    private $data = [
        "status" => 200,
        "body" => null,
        "headers" => [],
    ];
    public function __get($name)
    {
        if ($name == "body") {
            if (is_array($this->data["body"])) {
                return json_encode($this->data["body"]);
            }
            return $this->data["body"];
        }
        if ($name == "status") {
            return $this->get_status_str($this->data[$name]);
        }
        if (isset($this->data[$name])) {
            return $this->data[$name];
        } else {
            return null;
        }
    }

    private function get_status_str($status) {
        switch ($status) {
        case 200:
            return "200 OK";

        case 404:
            return "404 Not Found";

        default:
            throw new Exception("unknown status:" . $status);
        }
    }

    public function __set($name, $value) {
        switch ($name) {
        case "body":
            if (!is_string($value) && !is_array($value)) {
                throw new Exception("wrong data type for Response::body");
            }
            break;

        case "headers":
            if (!is_array($value)) {
                throw new Exception("wrong data type for Response::body");
            }
            foreach ($value as $line) {
                if (!is_string($line)) {
                    throw new Exception("wrong data type for Response::headers");
                }
            }
            break;

        case "status":
            if (!is_numeric($value) || intval($value) != $value) {
                throw new Exception("wrong data type for Response::status");
            }
            break;

        default:
            throw new Exception("unknown fileld of Request" . $name);
            break;
        }

        $this->data[$name] = $value;
    }
}
