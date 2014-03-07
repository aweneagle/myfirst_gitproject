<?php
    class IoRunner extends _CoreIo{
        private $io = null;
        private $map = null;
        public function __construct($params){
            self::assert(isset($params[0]), CORE_ERR_IO_CREATING, "no io");
            self::assert(isset($params[1]), CORE_ERR_IO_CREATING, "no map");
            self::assert(is_object($params[0]) && $params[0] instanceof _CoreIo, CORE_ERR_IO_CREATING, "wrong io obj");
            self::assert(method_exists($params[1], "mapin") && method_exists($params[1], "mapout") , CORE_ERR_IO_CREATING, "no mapin() function or mapout() function");
            $this->io = $params[0];
            $this->map = $params[1];
        }

        public function write($data){
            return $this->io->write($this->map->mapin($data)); 
        }

        public function read($index=null){
            return $this->io->read($this->map->mapout($index));
        }

        public function flush_normally(){}
    }
