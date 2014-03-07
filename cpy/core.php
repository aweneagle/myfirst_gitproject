<?php
    class Core {

        public static function set_ioloader(IClassLoader $loader){}

        public static function open($io_path){}

        public static function close($io_id){}

        public static function reopen($io_id){}

        public static function set_filter($io_id, $channel, $pos, IMap $filter){}

        public static function del_filter($io_id, $channel, $pos){}

        public static function add_filter_option($io_id, $channel, $filter_pos, IMap $option){}

        public static function set_filter_option($io_id, $channel, $filter_pos, $option_pos, IMap $option){}

        public static function del_filter_option($io_id, $channel, $filter_pos, $option_pos){}

        public static function set_classloader(IClassLoader $loader){}


        public static function set_jobloader(IClassLoader $loader){}

    }



