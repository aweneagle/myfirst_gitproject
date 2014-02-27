<?php

/*
about core:
    core    is an php engine for running simple jobs, it abstracts all inputs/outputs as io, which make job procedure more simple
    
about include.php:

    include.php 
            all functions here is public for applications outside this package, and only these function should be used outside this package
    
autho: awen
date: 2014-02-21

   functions implemented here(function module) should always follow the rules below:

         1. handle err in two ways:
            a) throw exception ,    if core_errtype() set the Core::$errtype = CORE_ERR_TYPE_EXCEPTION
            b) return false ,       if core_errtype() set the Core::$errtype = CORE_ERR_TYPE_ERRNO_ERRMSG

         2. var 'false' should never be used as the returned data when there's no error accuring

         3. core_assert() is the exception, it should always throw exception when failed 

*/

    if (!defined('CORE_ROOT')) {
        define('CORE_ROOT', dirname(realpath(__FILE__)));
    }
    
    include CORE_ROOT . '/core/core.php';

    define ("CORE_STDIN", _Core::STDIN);
    define ("CORE_STDOUT", _Core::STDOUT);
    define ("CORE_STDERR", _Core::STDERR);
    define ("CORE_ERRTYPE_ERRNO_ERRMSG", Core::ERR_TYPE_ERRNO_ERRMSG);        /* no exception will be thrown out, you should fetch errno, errmsg by core_get_errno() and core_get_errmsg() */
    define ("CORE_ERRTYPE_EXCEPTION", Core::ERR_TYPE_EXCEPTION);            /* all error will be thrown out */
    define ("CORE_IO_FLUSH_NORMALLY", _CoreIo::FLUSH_NORMALLY);
    define ("CORE_IO_FLUSH_NULL", _CoreIo::FLUSH_NULL);
    define ("CORE_IO_FLUSH_RETURN", _CoreIo::FLUSH_RETURN);

/*==================  error  module ===============*/
/*
core_assert:    core_assert( $assert, $errno, [[$errmsg,] $errmsg2,] ... )
@param  $assert 
@param  $errno 
@param  $errmsg 
@param  $errmsg2
*/
    function core_assert(){

        $args = func_get_args();
        Core::assert(count($args) >= 2,
                CORE_ERR_ASSERT,
                "more args need for function core_assert()");

        $assert = array_shift($args);
        if ($assert !== true) {
            $errno = array_shift($args);
            $errmsg = '';
            while (!empty($args)){
                if (is_string(($msg = array_shift($args)))) {
                    $errmsg .= ",".$msg;
                }
            }

            $errmsg = trim($errmsg, ",");

            throw new Exception($errmsg, $errno);
        }
    }


/*
core_errtype:   set the errtype for error reporting
@param  $errtype    there're some options here:
CORE_ERR_TYPE_EXCEPTION     when this options is set, all exceptions will be thrown normally
CORE_ERR_TYPE_ERRNO_ERRMSG  when this options is set, all exceptions will be caught, and errno & errmsg can be fetch by core_errno() & core_errmsg();
*/
    function core_errtype($errtype){
        return Core::errtype($errtype);
    }

    function core_get_errtype(){
        return Core::get_errtype();
    }

/*
core_errno: get the errno 
core_errmsg: get the errmsg
*/
    function core_errno(){
        return Core::get_errno();
    }
    function core_errmsg(){
        return Core::get_errmsg();
    }




/*===================  function module =================*/

/*
core_reg_autoload:   register the autoload function, it can be called more than once, and all registered functions will work
@param   $func     array,    $func[0], the autoload classname   
                             $func[1],  the autoload function
                   or func,  'auto_load'
 */
    function core_reg_autoload($func){
        return Core::reg_autoload($func);
    }


/*
core_add_classroot: add class root for all autoload functions
@param  $dir    the root dir
*/
    function core_add_classroot($dir){
        return Core::add_classroot($dir);
    }

/*
core_set_router: set the job router 
@param  $classname   it must be placed in $CORE_ROOT/core/routers.php, and it must extends the super class _CoreRouter
*/
    function core_set_jobrouter($classname){
        return Core::set_jobrouter($classname);
    }
    
    function core_set_iorouter($classname){
        return Core::set_iorouter($classname);
    }

/*
 * set the io option
 *
 *@param $io, int id
 *@param $option, string 
 *@param $val, mix
 *@return void
 */
    function core_set_option($io, $option, $val){
        return Core::set_option($io, $option, $val);
    }
    function core_set_options($io, array $options){
        return Core::set_options($io, $options);
    }


/*
core_run:   run the job
@param  $job_path, string ,  it's formate dependends on the router
*/
    function core_run($job_path, $params=null){
        return Core::run($job_path, $params);
    }


/*
core_open:  open an io(src)
@param  $io_apth        formate like this: "$classname:$params1,$params2,..."
        example:
            core_open("io_linefile:/data/log,0666");

@return $io     the descriptor(io) , it works like the file descriptor on linux     
*/
    function core_open($io_path){
        return Core::open($io_path);
    }


/*
core_read:  read data from io(src)
@param  $io         io descriptor, get by core_open();
@param  $options    some additional information for reading data
@return $data
*/
    function core_read($io, $options=null){
        return Core::read($io, $options);
    }


/*
core_write: write data into io(src)
@param  $io         io descriptor,
@param  $datas      data to be write
@return  $old_data, optional, if there is any
*/
    function core_write($io, $data){
        return Core::write($io, $data);
    }


/*
core_redirect:  redirect io "io_from" to "io_to",  all unflushed data in "io_from" will be write into "io_to"
@param  $io_from    
@param  $io_to
@return void
*/
    function core_redirect($io_from, $io_to){
        return Core::redirect($io_from, $io_to);
    }

/*
core_flush: flush data from buffer into io
@param  $io
@param  $option,    there're some options here:
CORE_IO_FLUSH_NORMAL:   do the flush normally
CORE_IO_FLUSH_RETURN:   not into the real io, but flush out the buffer from this function 
CORE_IO_FLUSH_NULL:     abandon the buffer
*/
    function core_flush($io, $option = CORE_IO_FLUSH_NORMALLY){
        return Core::flush($io, $option);
    }


/*
core_watch: monitor the io , usually used as log
@param  $io         the io descriptor 
@param  $monitor    the monitor descriptor 
@return void
*/
    function core_watch($io, $monitor){
        return Core::watch($io, $monitor);
    }

/*default init
 *
 *@return null
 */
    function core_default_init(){
        return Core::default_init();
    }
