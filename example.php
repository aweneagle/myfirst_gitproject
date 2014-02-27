<?php
    include "include.php";

    core_errtype(CORE_ERRTYPE_ERRNO_ERRMSG);
    core_default_init();

    if (core_errno() == CORE_ERR_NONE) {

        core_errtype(CORE_ERRTYPE_ERRNO_ERRMSG);
        //core_errtype(CORE_ERRTYPE_EXCEPTION);
        include "tmp_cfg.php";
        core_run("job.example");
        if (core_errno() == CORE_ERR_NONE) {

            Core_flush(CORE_STDOUT);

        } else {

            Core_flush(CORE_STDERR);
            Core_flush(CORE_STDOUT, CORE_IO_FLUSH_NULL);
        }

    } else {
        core_flush(CORE_STDERR);
    }
