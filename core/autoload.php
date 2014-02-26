<?php
    class CoreAutoLoadExample {
        public static function autoload($name){
            $namelen = strlen($name);
            $subpath = '';
            for ($i = 0; $i < $namelen; $i++) {
                if (ctype_upper($name[$i])) {
                    $subpath .= "/".strtolower($name[$i]);

                } else {
                    $subpath .= $name[$i];
                }
            }
            $subpath = trim($subpath, "/");
            foreach (_Core::get_classroots() as $dir) {
                $file = $dir."/".$subpath.".php";
                if (file_exists($file)) {
                    include $file;
                    if (!class_exists($name) && !interface_exists($name)) {
                        continue;
                    }
                }
            }
        }
    }
