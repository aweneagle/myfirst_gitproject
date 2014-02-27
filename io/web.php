<?php
    class   IoWeb   extends _CoreIo{
        public function __construct(){
            $this->data = array_merge($this->data, $_GET);
            $this->data = array_merge($this->data, $_POST);
        }
		public function read($key=null){
            if ($key === null) {
                return $this->data;
            } else {
                return @$this->data[$key];
            }
        }

		public function write($data){
            if (is_string($data) && ($data = json_decode($data,true)) === false) {
                throw new Exception("wrong input data for io", CORE_ERR_IO_WRITE);
            }
            foreach ($data as $k => $v) {
                $this->data[$k] = $v;
            }
		}

		protected function flush_normally(){
		}
	}
?>
