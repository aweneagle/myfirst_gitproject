<?php
    include CORE_ROOT . '/core/err.php';
    include CORE_ROOT . '/core/autoload.php';
    include CORE_ROOT . '/core/routers.php';

    class Core{

        const ERR_TYPE_EXCEPTION = 1;
        const ERR_TYPE_ERRNO_ERRMSG = 2;

        private static $errtype = self::ERR_TYPE_EXCEPTION;
        private static $errmsg = null;
        private static $errno = null;

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
           Core::_assert($assert, $errno, $errmsg, $errmsg1, $errmsg2, ...) 
                    it is use for inner modules , like CORE_ROOT/io/xxx.php, CORE_ROOT/io/xxxx.php ...
             */
        public function _assert(){
            $args = func_get_args();
            if (count($args) <= 2) {
                throw new Exception("failed to call Core::_assert", CORE_ERR_ASSERT);
            }
            $assert = array_shift($args);
			if ($assert === false) {
				$errno = array_shift($args);
				$errmsg = '';

				while (!empty($args)){
					if (is_string(($msg = array_shift($args)))) {
						$errmsg .= ','.$msg;
					}
				}
				$errmsg = trim($errmsg, ",");
				if (!is_numeric($errno)) {
					throw new Exception(  "wrong errno".",errno:".$errno.','.$errmsg, CORE_ERR_ASSERT);
				}
				throw new Exception( $errmsg , $errno );
			}
        }

        /* 
           Core::assert()
                it is use for outer modules, like include.php , somewhere else in application level ...
                */

        public function assert($assert, $errno, $errmsg){
            if ($assert !== true) {
                self::$errmsg = $errmsg;
                self::$errno = $errno;
                
                switch (self::$errtype){
                    case self::ERR_TYPE_EXCEPTION:
                        throw new Exception($errmsg, $errno);

                    case self::ERR_TYPE_ERRNO_ERRMSG:
                    default:
                        _Core::write(_Core::STDERR, $errmsg);
                        _Core::flush(_Core::STDERR);
                        return false;
                }
            }
            return true;

        }

        public static function __callStatic($name, $params){
            switch (self::$errtype) {
                case self::ERR_TYPE_ERRNO_ERRMSG:
                    try {
                        return self::_call($name, $params);
                    } catch (Exception $e) {
                        self::$errno = $e->getCode();
                        self::$errmsg = $e->getMessage();
                        _Core::write(_Core::STDERR, self::$errmsg);
                        _Core::flush(_Core::STDERR);
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
		protected $data = array();

        public function __construct(array $params){}
        abstract public function read($options=null);   /* options , string or array,  usually json string*/
        abstract public function write($data);          /* data,    string or array,    usually json string*/
        abstract protected function flush_return();     /* return string */
        protected function flush_null(){$this->data = array();}
        abstract protected function flush_normally();

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
            self::$src[ self::$available_srcid ] = self::$io_router->fetch($io_path);
            $curr = self::$available_srcid ;

            for ($i = self::$available_srcid ; self::$src[$i] != null ; $i ++) {
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
            Core::_assert(isset(self::$src[$io_id]) && self::$src[$io_id]!=null, 
                    CORE_ERR_IO_READ,
                    "wrong io_id for reading", "id:".$io_id);

            return  self::$src[$io_id]->read($options);
        }

        public static function write($io_id, $data){
            Core::_assert(isset(self::$src[$io_id]) && self::$src[$io_id]!=null, 
                    CORE_ERR_IO_WRITE,
                    "wrong io_id for writing", "id:".$io_id);

            return self::$src[$io_id]->write($data);
        }

        public static function flush($io_id, $option=_CoreIo::FLUSH_NORMALLY){
            Core::_assert(isset(self::$src[$io_id]) && self::$src[$io_id]!=null, 
                    CORE_ERR_IO_FLUSH,
                    "wrong io_id for flushing", "id:".$io_id);

            return self::$src[$io_id]->flush($option);
        }

        public static function redirect($io_from, $io_to){
            Core::_assert(isset(self::$src[$io_to]) && self::$src[$io_to]!=null, 
                    CORE_ERR_IO_REDIRECT,
                    "wrong io_to for redirection", "id:".$io_to);

            if (isset(self::$src[$io_from])) {
                $buffer = self::$src[$io_from]->flush(_CoreIo::FLUSH_RETURN);
                self::$src[$io_to]->write($buffer);
            }
            self::$src[$io_from] = self::$src[$io_to];
        }

        public static function run($job_path, $params){
            Core::_assert(self::$job_router != null, 
						CORE_ERR_JOB_RUNING,
						"no job router to work");
            $job = self::$job_router->fetch($job_path);
            return $job->run($params);
        }

        public static function set_jobrouter(_CoreRouter $router) {
            self::$job_router = $router;
        }

        public static function set_iorouter(_CoreRouter $router){
            self::$io_router = $router;
        }

        public static function add_classroot($dir){
            self::$class_roots[] = $dir;
        }
		
		public static function get_classroots(){
			return self::$class_roots;
		}

    }
