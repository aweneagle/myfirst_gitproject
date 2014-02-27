<?php
	class	IoLinefile	extends _CoreIo{
		public function __construct($params){}
		public function read($options=null){}
		public function write($data){
			if (!is_string($data)) {
                throw new Exception( "wrong input for writing into IoLineFile",CORE_ERR_IO_WRITE);
            }
			$this->data[] = $data;
		}
		protected function flush_normally(){
			foreach ($this->data as $line) {
				echo $line . "\n";
			}
			$this->data = array();
		}
	}
?>
