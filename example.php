<?php
    include "include.php";
    core_reg_autoload(array("CoreAutoLoadExample", "autoload")); 
    core_add_classroot("./");
    core_errtype(CORE_ERRTYPE_ERRNO_ERRMSG);
	core_set_jobrouter("CoreJobRouter");
	core_set_iorouter("CoreIoRouter");
    core_run("job.example");
