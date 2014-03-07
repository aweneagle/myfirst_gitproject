<?php

    class IoMemcached extends _CoreIo{
        public function __construct($params) {
            $this->host = @$params[0];
            $this->port = @$params[1];
        }
        public function read($index = null){
            $conn = new Memcached();
            $conn->addServer($this->host, $this->port);
            if ($conn) {
                return $conn->get($index);
            }
        }
        public function write($data) {
            if (@$data['key']) {
                $conn = new Memcached();
                $conn->addServer($this->host, $this->port);
                if ($conn) {
                    return $conn->set($data['key'], @$data['val']);
                }
            }
            return false;
        }
        public function flush_normally(){}
    }
