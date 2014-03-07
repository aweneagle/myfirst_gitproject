<?php
    include "include.php";

    class MapMysql{
        public function mapin($data){
            return array("query"=>"replace into db_pepper_admin.fortest set data=? , id=?", "columns" => array($data["val"], $data["key"]));
        }
        public function mapout($data){
        }
    }

    class MapNull{
        public function mapin($data){
            return $data;
        }
        public function mapout($data){}
    }

    class Runner{
        public $io = null;
        public $run_times = 20;

        public $data_size = 128;    // unit data size, 256 byte
        public $data_range = 13;     // 256b, 512b, 1k, 2k, 4k, 8k, 16k, 64k, 128k

        public $sec_round = 6;      // the round() num of micro seconds

        public $result = array();    // [data_size][run_no] => cost_time

        public function run(){
            $this->result = array();
            $newdata = $this->_create_data();
            for ($i = 0; $i < $this->data_range; $i ++) {
                for ($j = 0; $j < $this->run_times; $j ++) {

                    $begin = $this->_microtime();
                    core_write($this->io, array("key"=>1, "val"=>$newdata));
                    $end = $this->_microtime();

                    $this->result[$i][$j] = number_format($end - $begin, $this->sec_round);

                }
                $newdata .= $newdata;
            }
        }

        public function show(array $csv){
            foreach ($csv as $i => $line) {
                $csv[$i] = implode(",", $line);
            }
            return implode("\n", $csv);
        }

        public function invert($csv_str){
            $csv_arr = array_filter(explode("\n", $csv_str));
            if (!empty($csv_str)) {
                $str = array_shift($csv_arr);
            } else {
                $str = "null";
            }
            if (empty($csv_arr)) {
                return $str;
            }

            foreach ($csv_arr as $i => $line) {
                $csv_arr[$i] = explode(",", $line);
            }

            $new = array();
            foreach ($csv_arr as $y => $line) {
                foreach ($line as $x => $val) {
                    $new[$x][$y] = $val;
                }
            }

            foreach ($new as $i => $line) {
                $new[$i] = implode(",", $line);
            }
            $str .= "\n" . implode("\n", $new);
            return $str;
        }

        private function _microtime(){
            list($msec, $sec) = explode(" ", microtime());
            return ((float)$msec + (float)$sec);
        }

        private function _create_data(){
            $d = 'a';
            for($i = 0; $i < $this->data_size; $i ++){
                $d .= 'a';
            }
            return $d;
        }
    }

    core_errtype(CORE_ERRTYPE_EXCEPTION);
    core_default_init();
    $redis = new Runner();
    $mysql = new Runner();
    $memcached = new Runner();
    $mysql_web = new Runner();
    $redis_web = new Runner();
    $memcached_web = new Runner();

    $mysql->io = core_open('IoRunner(IoMysql(127.0.0.1, 3306, fanqu, vanchu2010), MapMysql())');
    $redis->io = core_open('IoRunner(IoRedis(127.0.0.1, 4444), MapNull())');
    $memcached->io = core_open('IoRunner(IoMemcached(127.0.0.1, 9999), MapNull())');

    $mysql_web->io = core_open('IoRunner(IoMysql(192.168.1.51, 3306, fanqu, vanchu2010), MapMysql())');
    $redis_web->io = core_open('IoRunner(IoRedis(192.168.1.51, 9999), MapNull())');
    $memcached_web->io = core_open('IoRunner(IoMemcached(192.168.1.51, 9998), MapNull())');
    $mysql->run();
    $redis->run();
    $memcached->run();
    $mysql_web->run();
    $redis_web->run();
    $memcached_web->run();

    $xaixs = array();
    $xaixs[0] = "x坐标/y坐标";
    $size = 0;
    $pow = pow(10, $mysql->sec_round);
    foreach ($mysql->result as $i => $res){
        if ($size == 0) {
            $size = $mysql->data_size;
        } else {
            $size += $size;
        }
        $xaixs[] = ($size > 1024) ? ($size / 1024)."k" : $size . "b";

        $mysql->result[$i] = $pow * ((float) array_sum($res) / (float)count($res));
    }

    foreach ($redis->result as $i => $res){
        $redis->result[$i] = $pow * ((float) array_sum($res) / (float)count($res));
    }
    foreach ($memcached->result as $i => $res){
        $memcached->result[$i] = $pow * ((float) array_sum($res) / (float)count($res));
    }

    foreach ($mysql_web->result as $i => $res){
        $mysql_web->result[$i] = $pow * ((float) array_sum($res) / (float)count($res));
    }
    foreach ($redis_web->result as $i => $res){
        $redis_web->result[$i] = $pow * ((float) array_sum($res) / (float)count($res));
    }
    foreach ($memcached_web->result as $i => $res){
        $memcached_web->result[$i] = $pow * ((float) array_sum($res) / (float)count($res));
    }

    $csv = array();
    array_unshift($mysql->result, "mysql");
    array_unshift($redis->result, "redis");
    array_unshift($memcached->result, "memcached");
    array_unshift($mysql_web->result, "mysql_web");
    array_unshift($redis_web->result, "redis_web");
    array_unshift($memcached_web->result, "memcached_web");
    $csv[] = array("io性能");
    $csv[] = $xaixs;
    $csv[] = $mysql->result;
    $csv[] = $redis->result;
    $csv[] = $memcached->result;
    $csv[] = $mysql_web->result;
    $csv[] = $redis_web->result;
    $csv[] = $memcached_web->result;

    echo $mysql->invert($mysql->show($csv));

