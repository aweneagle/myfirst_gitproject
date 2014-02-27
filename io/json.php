<?php
	class	IoJson	extends _CoreIo{
		public function read($key=null){
            if ($key === null) {
                return $this->data;
            } else {
                return @$this->data[$key];
            }
        }

		public function write($data){
            if ($data != null) {
                if (is_string($data)) {
                    $data = @json_decode($data,true);
                    if ( !is_array($data) ) {
                        throw new Exception("failed to json_decode()", CORE_ERR_IO_WRITE);
                    }
                }
                foreach ($data as $k => $v) {
                    $this->data[$k] = $v;
                }
            }

		}

		protected function flush_normally(){
            echo json_encode($this->data);
			$this->data = array();
		}
	}
?>
