<?php
	class	IoLinefile	extends _CoreCmdline{
		protected function flush_normally(){
			foreach ($this->data as $line) {
				echo $line . "\n";
			}
			$this->data = array();
		}
	}
?>
