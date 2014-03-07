<?php

    class IoRedis extends _CoreIo{
        public function __construct($params) {
            $this->host = @$params[0];
            $this->port = @$params[1];
        }
        public function read($index=null){
            $conn = new Redis();
            if ($conn->connect($this->host, $this->port)) {
                return $conn->get($index);
            } else {
                return false;
            }
        }

        public function write($data){
            $conn = new Redis();
            if (@$data['key'] && $conn->connect($this->host, $this->port)) {
                return $conn->set($data['key'], @$data['val']);
            }
            return false;
        }

        public function flush_normally(){}
    }
