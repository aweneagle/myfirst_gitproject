<?php

    class IoMpi extends _CoreIo{

        private $storage;
        private $cache;

        public function __construct($params){
            $this->storage = $params[0];  //mysql or redis
            $this->cache = $params[1];   //memcached
        }

        public function read($id = null){
            $res = $this->cache->read($id);
            if ($res) {
                return $res;
            }
            $res = $this->storage->read($id);
            $this->cache->write(array("key"=>$id, "val"=>$res));
            return $res;
        }

        public function write($data){
            $this->cache->write(array("key"=>$data["key"], "val"=>$data["val"]));
            $this->storage->write(array("key"=>$data["key"], "val"=>$data["val"]));
            return true;
        }

        public function flush_normally(){
        }
        
    }


    class IoMpiMysql extends IoMysql{
        private $db;
        private $table;
        private $field;
        private $mod;   //multi tables

        public function __construct($params){
            $this->db = $params[4];
            $this->table = $params[5];
            $this->field = $params[6];
            $this->mod = $params[7];
        }

        public function read($uid=null){
            return parent::read("select " . $this->field . " from " . $this->db . "." . $this->table . " where uid=" . $uid );
        }

        public function write($data){
            return parent::write("update " . $this->db . "." . $this->table . " set " . $this->field . "=" . $data["key"] . " where uid=" . $data["val"] );
        }
    }
