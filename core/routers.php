<?php
    class CoreJobRouter extends _CoreRouter{
        public function fetch($path){
            $path = str_replace(".","/",$path);
            $_classname = explode("/",$path);
            $classname = '';
            foreach ($_classname as $n) {
               $classname .= ucfirst($n);
            }
            Core::_assert(class_exists($classname, true), CORE_ERR_JOB_ROUTING, "failed to load job", "class:".$classname);
            return new $classname();
        }
    }

    class CoreIoRouter extends _CoreRouter{
        public function fetch($path){
			if (strpos(":", $path)) {
				$classname = trim(substr($path, 0, strpos(":",$path)));
				$params = substr($path, strpos(":",$path));
				$params = explode(",", $params);
				foreach ($params as &$p) {
					$p = trim($p);
				}
			} else {
				$classname = $path;
				$params = array();
			}
            Core::_assert(class_exists($classname, true), CORE_ERR_IO_ROUTING, "failed to load io", "io:".$classname);
            return new $classname($params);
        }
    }
