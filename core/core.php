<?php
    include CORE_ROOT . '/core/err.php';
    include CORE_ROOT . '/core/autoload.php';
    include CORE_ROOT . '/core/routers.php';

    class Core{

        const ERR_TYPE_EXCEPTION = 1;
        const ERR_TYPE_ERRNO_ERRMSG = 2;

        private static $errtype = self::ERR_TYPE_EXCEPTION;
        private static $errmsg = null;
        private static $errno = CORE_ERR_NONE;

        public function get_errtype(){
            return self::$errtype;
        }

        public function get_errmsg(){
            return self::$errmsg;
        }

        public function get_errno(){
            return self::$errno;
        }

        public function errtype($errtype){
            switch ($errtype) {
                case self::ERR_TYPE_EXCEPTION:
                case self::ERR_TYPE_ERRNO_ERRMSG:
                    self::$errtype = $errtype;
                    break;

                default:
                    return false;
            }
            return true;
        }

        /* 
           Core::assert($assert, $errno, $errmsg, $errmsg1, $errmsg2, ...) 
             */
        public function assert(){
            $args = func_get_args();
            if (count($args) <= 2) {
                throw new Exception("failed to call Core::_assert", CORE_ERR_ASSERT);
            }
            $assert = array_shift($args);
############if ($assert !== true) {
################$errno = array_shift($args);
################$errmsg = '';

################while (!empty($args)){
####################if (is_string(($msg = array_shift($args)))) {
########################$errmsg .= ','.$msg;
####################}
################}
################$errmsg = trim($errmsg, ",");
################if (!is_numeric($errno)) {
####################throw new Exception(  "wrong errno".",errno:".$errno.','.$errmsg, CORE_ERR_ASSERT);
################}
################throw new Exception( $errmsg , $errno );
############}
        }

        public static function __callStatic($name, $params){
            switch (self::$errtype) {
                case self::ERR_TYPE_ERRNO_ERRMSG:
                    try {
                        return self::_call($name, $params);
                    } catch (Exception $e) {
                        self::$errno = $e->getCode();
                        self::$errmsg = $e->getMessage();
                        try {
                            _Core::write(_Core::STDERR, self::$errmsg);
                            _Core::flush(_Core::STDERR);
                        } catch (Exception $e) {
                            die( "FAILED TO WRITE INTO [STDERR],MAYBE IT'S REDIRECTED WITH WRONG IO CLASS,err=".$e->getMessage());
                        }
                        return false;
                    }
                    break;

                case self::ERR_TYPE_EXCEPTION:
                default:
                    return self::_call($name, $params);

            }
        }

        public static function _call($name, $params){
            switch (count($params)) {

                case 0:
                    return _Core::$name();

                case 1:
                    return _Core::$name($params[0]);

                case 2:
                    return _Core::$name($params[0], $params[1] );

                case 3:
                    return _Core::$name($params[0], $params[1], $params[2]);

                case 4:
                    return _Core::$name($params[0], $params[1], $params[2], $params[3]);

                case 5:
                    return _Core::$name($params[0], $params[1], $params[2], $params[3], $params[4]);

                case 6:
                    return _Core::$name($params[0], $params[1], $params[2], $params[3], $params[4], $params[5]);

                case 7:
                    return _Core::$name($params[0], $params[1], $params[2], $params[3], $params[4], $params[5], $params[6]);

                default:
                    return self::assert(false, CORE_ERR_CALLMETHOD, "wrong method name for Core class", "func:".$name);
            }
        }
    }

    abstract class _CoreRouter{
        abstract public function fetch($job_path);
    }

    abstract class _CoreJob{
        abstract public function run($params);
    }

    abstract class _CoreIo{
        const FLUSH_NORMALLY = 1;
        const FLUSH_RETURN = 2;
        const FLUSH_NULL = 3;
########protected $data = array();
        protected $options = array();

        public function __construct(array $params){}
        abstract public function read($options=null);   /* options , string or array,  usually json string*/
        abstract public function write($data);          /* data,    string or array,    usually json string*/
        protected function flush_return(){ $tmp = json_encode($this->data); $this->data = array(); return $tmp;}  /* return string */
        protected function flush_null(){$this->data = array();}
        abstract protected function flush_normally();

        public function set_option($option, $value) {
            $this->options[$option] = $value;
        }

        protected function assert(){
            $args = func_get_args();
            if (count($args) <= 2) {
                throw new Exception("failed to call assert()", CORE_ERR_ASSERT);
            }
            $assert = array_shift($args);
############if ($assert !== true) {
################$errno = array_shift($args);
################$errmsg = '';

################while (!empty($args)){
####################if (is_string(($msg = array_shift($args)))) {
########################$errmsg .= ','.$msg;
####################}
################}
################$errmsg = trim($errmsg, ",");
################if (!is_numeric($errno)) {
####################throw new Exception(  "wrong errno".",errno:".$errno.','.$errmsg, CORE_ERR_ASSERT);
################}
################throw new Exception( $errmsg , $errno );
############}
        }

        public function flush($flush_way){
            switch ($flush_way) {

                case self::FLUSH_RETURN:
                    return $this->flush_return();

                case self::FLUSH_NULL:
                    return $this->flush_null();

                case self::FLUSH_NORMALLY:
                default:
                    return $this->flush_normally();
            }
        }
    }
####class####_CoreCmdline####extends _CoreIo{
        public function read($options = null){}
########public function write($data){
            if (is_array($data) ) {
                if (!empty($data)) {
                    $data = json_encode($data) . "\n";
                } else {
                    $data = '';
                }
            } else if (is_string($data)) {

                /* filter the string "[]" */
                $tmp = @json_decode($data, true);
                if (is_array($tmp) && empty($tmp)) {
                    $data = '';
                }else{
                    $data .= "\n";
                }
            } else {
                $data = 'UNKNOWN ERR MSG,type=['.gettype($data).']';
            }
            echo $data ;
########}
########protected function flush_normally(){
########}
####}

    class _Core {

        const MONITOR_CHANNEL_IN = 1;
        const MONITOR_CHANNEL_OUT = 2;
        const MONITOR_CHANNEL_IN_AND_OUT = 3;

        const STDIN = 0;
        const STDOUT = 1;
        const STDERR = 2;

        private static $job_router = null;
        private static $io_router = null;

        private static $class_roots = array();

        private static $src = array( self::STDIN=>null, self::STDOUT=>null, self::STDERR=>null );
        private static $available_srcid = 3;


        public static function open($io_path){
            $newio = self::$io_router->fetch($io_path);
            if ( !$newio instanceof _CoreIo ){
                throw new Exception("class is not instance of _CoreIo,io_path=".$io_path,CORE_ERR_IO_OPEN);
            }
            self::$src[ self::$available_srcid ]  = $newio;
            $curr = self::$available_srcid ;

            for ($i = self::$available_srcid ; @self::$src[$i] != null ; $i ++) {
                ;   
            }

            self::$available_srcid = $i;
            return $curr;
        }

        public static function close($io_id){
            self::$src[$io_id] = null;
            self::$available_srcid  = $io_id;
            return true;
        }

        public static function read($io_id, $options=null){
            if (!isset(self::$src[$io_id]) || self::$src[$io_id]==null) { 
                throw new Exception(
                    "wrong io_id for reading,". "id:".$io_id,
                    CORE_ERR_IO_READ
                    );
            }

            return  self::$src[$io_id]->read($options);
        }

        public static function write($io_id, $data){
            if (!isset(self::$src[$io_id]) || self::$src[$io_id]==null) {
                throw new Exception(
                    "wrong io_id for writing,". "id:".$io_id,
                    CORE_ERR_IO_WRITE
                    );
            }

            return self::$src[$io_id]->write($data);
        }

        public static function flush($io_id, $option=_CoreIo::FLUSH_NORMALLY){
            if (!isset(self::$src[$io_id]) || self::$src[$io_id]==null){
                throw new Exception(
                    "wrong io_id for flushing,". "id:".$io_id,
                    CORE_ERR_IO_FLUSH
                    );
            }

            return self::$src[$io_id]->flush($option);
        }

        public static function set_option($io_id, $option, $val){
            Core::assert(@self::$src[$io_id] != null, CORE_ERR_IO_OPTIONS, "null io_id", "id:".$io_id);
            self::$src[$io_id]->set_option($option, $val);
            return true;
        }

        public static function set_options($io_id, array $options) {
            Core::assert(@self::$src[$io_id] != null, CORE_ERR_IO_OPTIONS, "null io_id", "id:".$io_id);
            foreach ($options as $key => $val) {
                self::$src[$io_id]->set_option($key, $val);
            }
            return true;
        }

        public static function redirect($io_from, $io_to){
            if (self::$src[$io_to]==null) { 
                throw new Exception(
                    "wrong io_to for redirection,". "id:".$io_to,
                    CORE_ERR_IO_REDIRECT
                        );
            }

            if (@self::$src[$io_from] != null) {
                $buffer = self::$src[$io_from]->flush(_CoreIo::FLUSH_RETURN);
                self::$src[$io_to]->write($buffer);
            }
            self::$src[$io_from] = self::$src[$io_to];
        }

        public static function default_init(){
            // 1. set default stderr 
            if (self::$src[self::STDERR] == null){
                self::$src[self::STDERR] = new _CoreCmdline(array());
            }
            // 4. set default autoloader
            self::reg_autoload(array("CoreAutoLoadExample", "autoload")); 
            // 5. add curr dir into class root
            self::add_classroot("./");
            // 6. set default job router
            self::set_jobrouter("CoreJobRouter");
            // 7. set default io router
            self::set_iorouter("CoreIoRouter");
            // 2. set default stdout 
            if (self::$src[self::STDOUT] == null){
                $stdout = self::open("IoJson()");
                self::redirect(self::STDOUT, $stdout);
            }
            // 3. set default stdin 
            if (self::$src[self::STDIN] == null){
                $stdin = self::open("IoWeb()");
                self::redirect(self::STDIN, $stdin);
            }
        }

        public static function run($job_path, $params){
            if (self::$job_router == null) { 
                throw new Exception(
########################"no job router to work",
########################CORE_ERR_JOB_RUNING
                            );
            }
            $job = self::$job_router->fetch($job_path);
            return $job->run($params);
        }

        public static function set_jobrouter($classname) {
            if ( !class_exists($classname)) {
                throw new Exception(
                        "your io router should be placed in core/routers.php,". "class:".$classname,
                        CORE_ERR_SET_IOROUTER
                        );
            }
            self::$job_router = new $classname();
        }

        public static function set_iorouter($classname){
            if ( !class_exists($classname)) {
                throw new Exception(
                        "your io router should be placed in core/routers.php,". "class:".$classname,
                        CORE_ERR_SET_IOROUTER
                        );
            }
            self::$io_router = new $classname();
        }

        public static function add_classroot($dir){
            self::$class_roots[] = $dir;
        }
########
########public static function get_classroots(){
############return self::$class_roots;
########}

        public static function reg_autoload($func){

            if (is_array($func)) {

                $classname = @$func[0];
                $function = @$func[1];

                Core::assert(
                            class_exists($classname),
                            CORE_ERR_AUTOLOAD_REG,
                            "your auto load class should be placed in core/autoload.php ", "class:".$classname) ;

                Core::assert(
                            method_exists($classname, $function),
                            CORE_ERR_AUTOLOAD_REG,
                            "your autoload class has not implemented it's function ", "class:".$classname, "func:".$function); 

                spl_autoload_register(array($classname, $function));

                return true;

            } else if (is_string($func)) {

                Core::assert(
                            function_exists($func),
                            CORE_ERR_AUTOLOAD_REG,
                            "your auto load function should be placed in core/autoload.php ", "func:".$func); 

                spl_autoload_register($func);

                return true;

            } else {
                Core::assert(false, 
                        CORE_ERR_AUTOLOAD_REG,
                        "wrong input for auto load");
            }
        }

    }
