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
            $tmp = $subpath;
            $subpath = trim($tmp, "/");
            $tmp = explode("/",$tmp);
            array_pop($tmp);
            $subpath2 = implode("/",$tmp);
            foreach (_Core::get_classroots() as $dir) {

                $file = $dir."/".$subpath2.".php";          //try path  "a/b.php" for class ABC
                if (file_exists($file)) {
                    include $file;
                    if (class_exists($name) || interface_exists($name)) {
                        return ;
                    }
                }

                $file = $dir."/".$subpath.".php";           //try path  "a/b/c.php" for class ABC
                if (file_exists($file)) {
                    include $file;
                    if (class_exists($name) || interface_exists($name)) {
                        return ;
                    }
                }
            }
        }
    }
