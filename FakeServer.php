<?php
use Exception;
class FakeServer {
    private $parent2child = null;
    private $child2parent = null;

    //子进程pid
    private $child;

    public function __destruct() {
        if ($this->child) {
            //parent
            pcntl_waitpid($this->child);
            @unlink($this->parent2child);
            @unlink($this->child2parent);
        }
    }

    public function serve() {
        /*
         * 安全模式下无法在  /tmp 目录下创建文件， 详情见 posix_mkfifo() 函数说明
         */
        $this->parent2child = tempnam("/tmp", "fake_server");
        $this->child2parent = tempnam("/tmp", "fake_server");
        if (!posix_mkfifo($this->parent2child)) {
            throw new Exception("failed to create pipe of:" . $this->parent2child);
        }
        if (!posix_mkfifo($this->child2parent)) {
            throw new Exception("failed to create pipe of:" . $this->child2parent);
        }
        if (!$this->p2c_fp = fopen($this->parent2child)) {
            throw new Exception("failed to open pipe of:" . $this->parent2child);
        }
        if (!$this->c2p_fp = fopen($this->child2parent)) {
            throw new Exception("failed to create pipe of:" . $this->child2parent);
        }

        if (($this->child = pcntl_fork()) == 0) {
            //child
            $c2p = fopen($this->child2parent, "w");

        } else {
            usleep(15);
        }
    }

    public function response(){}
    public function when(){}
    protected function resp_when($when){ return "hello world"; }
}
