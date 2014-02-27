<?php
    class JobExample extends _CoreJob{
        public function run($params){
            $mysql_id = Tmp::$mysql;
            $data = core_read($mysql_id, "select * from db_pepper_admin.login");
            core_redirect(CORE_STDOUT, CORE_STDERR);
            core_write(CORE_STDOUT, $data);
            $data = core_read($mysql_id, array(
                        "query"=>"select passwd from db_pepper_admin.login where uname = ?",
                        "columns"=>array(array("fanqu"), array("t")))
                    );

            $data = core_write($mysql_id, array(
                        "query"=>"update db_pepper_admin.login set passwd = ? where uname = ?",
                        "columns"=>array(array("aaawen", "fanqu"), array("awen", "t")))
                    );
            core_write(CORE_STDOUT, $data);
        }
    }
