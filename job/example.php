<?php
    class JobExample extends _CoreJob{
        public function run($params){
            $io = core_open("IoJson()");
            $io_err = core_open("IoLinefile()");
            core_redirect(CORE_STDIN, $io);
            core_redirect(CORE_STDOUT, $io);
            core_redirect(CORE_STDERR, $io_err);

            core_write(CORE_STDIN, array("name"=>"here is my name"));
            $myname = core_read(CORE_STDOUT, "name");
            core_assert(false, 1, "this is for test", "test:".$myname);
        }
    }
