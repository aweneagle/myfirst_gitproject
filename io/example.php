<?php
    class IoExample extends _CoreIo{

        public function read($key=null){
            if ($key === null) {
                return $this->data;
            } else {
                return @$this->data[$key];
            }
        }

        protected function flush_normally(){
            echo json_encode($this->data);
            $this->data = array();
        }

        protected function flush_return(){
            $tmp = $this->data;
            $this->data = array();
            return json_encode($tmp);
        }

        public function write($data){
            if (!is_array($data)) {
                $data = @json_decode($data, true);
                if (empty($data)) {
                    throw new Exception(
                    "wrong data for writing into IoExample",
                    CORE_ERR_IO_WRITE );
                }
            }

            foreach ($data as $k => $v) {
                $this->data[$k] = $v;
            }

        }
    }
?>
