<?php
    class IoMysql extends _CoreIo{
        public function __construct($params) {
            $this->host = @$params[0];
            $this->port = @$params[1];
            $this->user = @$params[2];
            $this->passwd = @$params[3];
        }

        public function read("select * from $db.$table where id=$id"){}
        public function write("update $db.$table replace values(id,data,ctime,version) ($id, $data, now, version++)");
    }

    class IoMemcached extends _CoreIo{
        public function __construct($params) {
            $this->host = @$params[0];
            $this->port = @$params[1];
        }
        public function read("get($id)"){}
    }


    class IoRedis extends _CoreIo{
        public function __construct($params) {
            $this->host = @$params[0];
            $this->port = @$params[1];
        }
        public function read("get($id)"){}
    }

    class IoMpi extends _CoreIo{
        const STOR_TYPE_MYSQL = 1; 
        const STOR_TYPE_REDIS = 2;

        private $storage;
        private $storage_type;

        private $op_mysql_db;
        private $op_mysql_table;

        public function __construct($params){
            $this->storage = $params[0];  //mysql or redis
            $this->storage_type = $params[1];  //storage type

            $this->table_router = $params[2];//table_router
            $this->cache = $params[3];   //memcached
        }
        private function _read_storage($id) {
            if ($this->storage_type == self::STORE_TYPE_MYSQL) {
                return $this->storage->read("select * from ".$this->op_mysql_db.".".$this->op_mysql_table." where id=".$id.";");
            }else if($this->storage_type == self::STORE_TYPE_
        }

        public function read($id, $index){
            if (!$this->cache->read(array("k"=>$id, "op"=>"get"))) {
                return $this->storage($id);       
            }
        }
    }


