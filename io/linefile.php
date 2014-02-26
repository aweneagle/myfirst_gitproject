<?php
	class	IoLinefile	extends _CoreIo{
		public function __construct($params){}
		public function read($options=null){}
		public function write($data){
			Core::_assert(is_string($data), CORE_ERR_IO_WRITE, "wrong input for writing into IoLineFile");
			$this->data[] = $data;
		}
		protected function flush_normally(){
			foreach ($this->data as $line) {
				echo $line . "\n";
			}
			$this->data = array();
		}
		protected function flush_return(){
			$data = $this->data;
			$this->data = array();
			return $data;
		}
	}
?>