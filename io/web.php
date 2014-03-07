<?php

    class IoWeb extends _CoreIo{
        public function read($key=null){
            return @$this->data[$key];
        }

        public function write($data){
            return $this->data[@$data["key"]] = @$data["val"];
        }

        public function flush_normally(){
            echo @json_encode($this->data);
        }
    }
