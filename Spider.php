<?php
namespace App\Lib;

class Spider
{

    /* 
     * 当前页面的 referer
     */
    public $referer = null;

    /*
     * 当前页面的 url
     */
    public $url = null;

    /*
     * 当前页面的 method;
     */
    public $method = null;

    /*
     * 错误处理函数, 需要配合 run() 使用
     * function on_error(Spider $s), 函数无需返回
     */
    public $on_error = null;

    /*
     * 页面处理函数, 需要配合 run() 使用
     * function on_succ(Spider $s), 函数无需返回
     */
    public $on_succ = null;

    /*
     * 上报redis的情况
     */
    public $health_notify = null;

    /*
     * 紧急中断开关
     */
    public $halt = false;

    /*
     * 每爬完一个页面休息的毫秒数
     */
    public $sleep_mill_seconds = 0;

    /*
     * url 获取/存储 函数
     * function url_queue($op, array $hrefs),  函数要返回需要的url
     *
     * @param   $op = [POP | PUSH]
     * @param   $hrefs = 
     * [
     *      ["url" => url, "referer" => referer]
     * ]
     *
     * 所传url可以是两种格式之一:
     * @return 
     *  示例:
     *      ["url" => url, "referer" => referer]
     *      或者
     *      url,
     *
     */
    public $url_queue = null;

	/*
	 * $request, 当前请求
	 * [
	 * 	"url" => "http://www.google.com",
	 *
	 * 	"method" => "GET",
	 *
	 * 	"header" => [
	 *		"Host: play.google.com",
	 *		"User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:49.0) Gecko/20100101 Firefox/49.0",
	 *		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9;q=0.8",
	 *		"Accept-Language: zh-TW,zh;q=0.8,en-US;q=0.5,en;q=0.3",
	 *		"Upgrade-Insecure-Requests: 1",
	 * 	],
	 *
	 * 	"post_fields" => [
	 * 		"name" => "awen",
	 *		"age" => 12
	 * 	],
	 *
	 * 	"post_fields" => "name=awen&age=12",
	 *
	 * 	"curl_options" => [
	 * 		CURLOPT_RETURNTRANSFER => 1
	 * 	],
	 *
	 * 	"cookies" => [
	 * 		name => value
	 * 	]
	 * ]
	 */
	public $request = [];


	/*
	 * $response, 存储当前请求的返回结果
	 * [
	 * 	"code" => 200,
	 *
	 * 	"header" => [
	 * 		"...",
	 * 		"..."
	 * 	],
	 *
	 * 	"cookies" => [
	 * 		"PID" => [
	 * 			"Name" => "PID",
	 * 			"Value" => "1afaa13123",
	 * 			"Expires" => "2016/09/01 10:00:00",
	 * 			"Domain" => ".google.com",
	 * 			"Max-Age" => 0,
	 * 			"HttpOnly" => 1
	 * 		],
	 * 		......
	 * 	],
	 *
	 * 	"body" => "<html><title> hello! </title></html>"
	 * ]
	 */
	public $response = [];


	/*
	 * $cookies 
     * "浏览器"所有的cookies;
	 * [
 * 	 	 [
 * 	 		"Name" => "PID",
 * 	 		"Value" => "09fa9af",
 * 	 		"Expires" => "2016/09/10 10:00",
 * 	 		"Domain" => ".google.com"
 * 	 	 ]
	 * ]
	 */
	private $cookies = [];


    /*
     * 如果开启debug_request = true, 将会把每个请求的 request 和 response 详细 echo 到标准输出
     */
    public $debug_request = false;

    /*
     * 如果开启debug_request_verbose, 将会把每个请求的 request 和 response 详细 echo 到标准输出
     */
    public $debug_request_verbose = false;

    /*
     * 如果开启debug_health = true, 将会在每次请求页面之后, 详细打印爬虫的状态
     */
    public $debug_health = false;

    /*
     * 错误信息
     */
	public $error = null;

    /*
     * 爬虫入口
     * [
     *    [
     *       "refere" => "",
     *       "url" => "",
     *    ]
     * ]
     */
    private $entries = [];

    /*
     * 爬过的url(取md5)
     */
    private $history = [];

    /*
     * @param  $urls = [
     *      [
     *          "url" => null,
     *          "method" => "GET",
     *          "header" => [
     *              ...
     *          ],
     *          "post_fields" => [
     *              ...
     *          ],
     *          "curl_options" => [
     *              ...
     *          ]
     *      ],
     *      ...
     *  ]
     */
    public function request($urls = []) {
        $curl_pool = [];
        foreach ($urls as $u) {
            $h = curl_init();
            if ($h) {
                $this->init_curl($h, $u);
                $curl_pool[$u] = $h;
            }
        }

        $mh = curl_multi_init();
        $this->multi_curl($mh, $curl_pool);
        foreach ($curl_pool as $u => $h) {
            $content = curl_multi_getcontent($h);
            $this->handle_content($u, $content, $h);
        }

        foreach ($curl_pool as $u => $h) {
            curl_close($h);
        }
        curl_multi_close($mh);
    }

    private function init_curl($ch, $url_info)
    {
        // 根据method设置 请求方法
        // 根据encoding 设置 curl
        // 设置postfield
        
        /* 请求头 */
        // 拼接cookie
        // 设置content length
        $options = [
            CURLOPT_URL => $url_info["url"],
            CURLOPT_CUSTOMREQUEST => "DELETE"
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_HEADER => 1,
        ];

        curl_setopt_array($ch, $options);
    }

    private function handle_content($url, $content, $curl_handle)
    {
        //如果content === false, 打error日志
        //从curl_handle取请求头的size
        //  根据size, 取出header
        //  根据size, 取出body
        //
        //  根据header, 取出code, httpversion
    }

    /*
     * run() 函数以一定数量的url为入口, 自动爬取页面
     *
     */
    public function run(array $urls, array $header = [])
    {
        $this->entries = [];
        $this->put_entries($urls);
        while ($entry = $this->pop_url()) {

            /* 
             * 去除 $url_info["fragment"], 即 #
             */
            $url_info = parse_url($entry["url"]);
            $host = $url_info["host"];
            $scheme = $url_info["scheme"];

            if (isset($url_info['query'])) {
                $query = "?" . $url_info['query'];
            } else {
                $query = "";
            }
            $url = $scheme . "://" . $host . $url_info['path'] . $query;

            // 用于相对路径
            if (isset($url_info['path'])) {
                $path_root = substr($url_info["path"], strrpos($url_info["path"], "/") + 1);
            } else {
                $path_root = '/';
            }

            $referer = $this->referer = $entry["referer"];


            $entry_header = $header;
            if ($referer != null) {
                $entry_header[] = "Referer: " . $referer;
            }
            $res = $this->view($url, "GET", $entry_header);

            if ($res === false) {
                if ($this->on_error) {
                    $func = $this->on_error;
                    $func($this);
                } else {
                    echo "[ERROR] $url : " . $this->error . "\n";
                }
            } else {
                if ($this->on_succ) {
                    $func = $this->on_succ;
                    $func($this);
                } else {
                    echo "[SUCC] $url\n";
                    $dir = "./spider";
                    if (is_dir($dir) || mkdir($dir)) {
                        file_put_contents("./spider/" . urlencode($url) . ".page", $this->response["body"]);
                    }
                }
            }

            $new_urls = [];
            /* 从页面中获取 url */
            $new_urls = array_merge($new_urls, $this->hrefs_from_page($this->response['body']));
            /* 从响应请求头中获取 url */
            $new_urls = array_merge($new_urls, $this->hrefs_from_header($this->response['header']));

            foreach ($new_urls as $i => $u) {
                $u = htmlspecialchars_decode($u);
                $info = parse_url($u);
                /* 过滤掉例如 # 一类的url */
                if (!isset($info["path"])) {
                    unset($new_urls[$i]);
                    continue;
                }
                /* 相对路径: 以当前url为参考, 补全 sheme */
                if (!isset($info["scheme"])) {
                    $info["scheme"] = $scheme;
                }
                /* 相对路径: 以当前url为参考, 补全 host */
                if (!isset($info["host"])) {
                    $info["host"] = $host;
                }
                /* 相对路径: 不以 / 为开头， 补全 path */
                if ($info["path"][0] != '/') {
                    $info["path"] = $path_root . $info["path"];
                }
                $new_urls[$i] = $info["scheme"] 
                    . "://" 
                    . $info["host"] 
                    . $info["path"] 
                    . (isset($info["query"]) ? ("?" . $info["query"]) : "");
            }
            $this->push_urls($new_urls, $url);
            $this->debug_health();

            if ($this->sleep_mill_seconds && is_numeric($this->sleep_mill_seconds)) {
                usleep(intval($this->sleep_mill_seconds) * 1000);
            }

            if ($this->halt) {
                break;
            }
        }
    }

    private function push_urls($urls, $url)
    {
        $new_urls = array_unique($urls);
        /* 添加 referer */
        foreach ($new_urls as $i => $u) {
            $new_urls[$i] = ["url" => $u, "referer" => $url];
        }
        if ($this->url_queue) {
            $func = $this->url_queue;
            $func("PUSH", $new_urls);
        } else {
            $this->entries = array_merge($this->entries, $new_urls);
        }
    }

    public function hrefs_from_page($page)
    {
        $hrefs = [];
        $m1 = preg_match_all('/href\s*=\s*\'([^\']+)\'/', $page, $match1);
        $m2 = preg_match_all('/href\s*=\s*"([^"]+)"/', $page, $match2);
        if ($m1) {
            $hrefs = array_merge($match1[1], $hrefs);
        }
        if ($m2) {
            $hrefs = array_merge($match2[1], $hrefs);
        }
        return array_map("trim", $hrefs);
    }

    public function hrefs_from_header(array $header)
    {
        foreach ($header as $h) {
            if (preg_match('/^location\s*:\s*(.*)$/', $h, $match)) {
                return [ trim($match[1]) ];
            }
        }
        return [];
    }

	public function view($url, $method = "GET", $header = [], $post_fields = null, $curl_options = [])
	{
		$this->request = $this->ini_request();
        $this->response = $this->ini_reponse();
        $this->error = null;
        $this->url = $url;
        $this->method = $method;
        $url_info = parse_url($url);
        if (!isset($url_info['scheme'])) {
            $this->error = '[ERROR] no scheme found(https or http): ' . $url;
            return false;
        }
        if (!isset($url_info['host'])) {
            $this->error = '[ERROR] no host found: ' . $url;
            return false;
        }
        $this->host = $url_info['host'];
        $this->scheme = $url_info['scheme'];
        $cookies = $this->cookies($this->host);
        $options = $curl_options;
        $options[CURLOPT_RETURNTRANSFER] = 1;
		$options[CURLOPT_HEADER] = 1;
        if (is_array($post_fields)) {
            $post_fields_str = [];
            foreach ($post_fields as $key => $val) {
                $post_fields_str[] = $key . '=' . $val;
            }
            $post_fields = implode("&", $post_fields_str);
        }

        if ($cookies) {
            /* 先查找header中是否有设置 cookie */
            $ck_header_found = false;
            foreach ($header as $i => $h) {
                if (preg_match('/^Cookie\s*:\s*.*$/i', $h)) {
                    $header[$i] .= "; " . $this->cookies_header($cookies);
                    $ck_header_found = true;
                }
            }
            if (!$ck_header_found) {
                $header[] = "Cookie: " . $this->cookies_header($cookies);
            }
        }

		$this->request['url'] = $url;
		$this->request['method'] = $method;
		$this->request['header'] = $header;
		$this->request['post_fields'] = $post_fields;
		$this->request['cookies'] = $cookies;
		$this->request['curl_options'] = $options;

        $this->debug_request();

        if (!$ch = curl_init($url)) {
            $this->error = "[ERROR] failed curl_init(), url=$url";
            return false;
        }
        if ($header) {
            foreach ($header as $h) {
                $options[CURLOPT_HTTPHEADER][] = $h;
                if (preg_match('/^Accept-Encoding\s*:\s*(.+)$/', $h, $match)) {
                    $options[CURLOPT_ENCODING] = $match[1];
                }
            }
        }
        $method = strtoupper($method);
        switch ($method) {
        case "POST":
            $options[CURLOPT_POST] = 1;
            break;

        default:
            $options[CURLOPT_CUSTOMREQUEST] = $method;
            break;
        }

        if ($post_fields) {
            $options[CURLOPT_POSTFIELDS] = $post_fields;
            if (is_array($post_fields)) {
                $str = [];
                foreach ($post_fields as $key => $val) {
                    $str[] = $key . '=' . $val;
                }
                $str = implode("&", $str);
            } else {
                $str = $post_fields;
            }
            $options[CURLOPT_HTTPHEADER][] = "Content-Length: " . strlen($str);
        }

        curl_setopt_array($ch, $options);

        $res = false;
        for ($i = 0; $i < 3 && $res === false; $i ++) {
            $res = curl_exec($ch);
        }

        $return = $res;
        if (false === $res) {
            $this->error = $this->_error($ch);

        } else {
            $this->response['code'] = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            if (!isset($options[CURLOPT_HEADER]) || !$options[CURLOPT_HEADER]) {
                $this->response['body'] = $res;
            } else {
                $header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
                $this->response['header'] = array_filter(explode("\r\n", substr($res, 0, $header_size)));

                $first_header = array_shift($this->response['header']);
                if (!preg_match('/^(HTTP\/[\d\.]+)\s/', $first_header, $http_version)) {
                    $this->error = "[ERROR] wrong header:" . $first_header;
                    $return = false;
                } else {
                    foreach ($this->response['header']as $header) {
                        if (preg_match('/^set-cookie\s*:\s*(.*)$/i', $header, $match)) {
                            $ck_agv = $this->cookie_str2arr($match[1], $this->host);
                            if (isset($ck_agv['Name'])) {
                                $this->response['cookies'][$ck_agv['Name']] = $ck_agv;
                            }
                        }
                    }
                    $this->response['version']= $http_version[1];
                    $this->response['body']= substr($res, $header_size);
                }
            }
        }
        curl_close($ch);
        $this->debug_response();
        $this->add_cookies($this->response['cookies']);
        return $return;
    }

	public function error()
	{
		return $this->error;
	}

    /*
     * cookies() 根据所访问的域名来取对应的cookies
     */
    public function cookies($host = null)
    {
        if ($host == null) {
            return $this->cookies;
        } else {
            $ck = []; 
            foreach ($this->cookies as $c) {
                if ($c['domain'][0] == '.') {
                    $len = strlen($c['domain']);
                    if (substr($host, -$len) == $c['domain']) {
                        $ck[] = $c; 
                    }   
                } elseif ($c['domain'] == $host) {
                    $ck[] = $c; 
                }   
            }   
            return $ck;
        }   

    }

    public function reqinfo()
    {
        return $this->request;
    }


    private function cookie_str2arr($header_str, $domain)
    {
        $ck_agv = [];
        foreach (explode(";", $header_str) as $pair) {
            $pair = trim($pair);
            $pos = strpos($pair, "=");
            if ($pos === false) {
                $key = $pair;
                $val = '';
            } else {
                $key = substr($pair, 0, $pos);
                $val = substr($pair, $pos + 1);
            }
            switch (strtolower($key)) {
            case "path":
            case "expires":
            case "domain":
            case "secure":
            case "httponly":
            case "max-age":
                $ck_agv[strtolower($key)] = $val;
                break;

            default:
                $ck_agv['Name'] = $key;
                $ck_agv['Value'] = $val;
                break;
            }
        }
        if (!isset($ck_agv['domain'])) {
            $ck_agv['domain'] = $domain;
        }
        return $ck_agv;
    }

    private function cookies_header($cookies)
    {
        $pairs = [];
        foreach ($cookies as $value) {
            $pairs[] = $value['Name'] . "=" . $value["Value"];
        }
        return implode("; ", $pairs);
    }

	private function replace($preg, $str, $params)
	{
		if (preg_match_all($preg, $str, $match)) {
			$to_replace = [];
			foreach ($match[1] as $i => $m) {
				if (isset($params[$m])) {
					$to_replace[$match[0][$i]] = $params[$m];
				}
			}
			$str = strtr($str, $to_replace);
		}
		return $str;
	}	


	private function _error($ch)
	{
		return "[ERRNO:" . curl_errno($ch) . "][ERROR]:" . curl_error($ch);
	}

    private function debug_response()
    {
        echo "\n";
        if (!$this->debug_request_verbose && !$this->debug_request) {
            return ;
        }
        $response = $this->response;
        unset($response['body']);
        if ($this->response) {
            echo "[response code] " . $this->response['code'] . "\n";
            if ($this->debug_request_verbose) {
                foreach ($this->response['header'] as $h) {
                    echo "[response header]" . $h . "\n";
                }
                echo "\n";
            }
        } else {
            echo "[response NULL]\n";
        }
    }

    private function debug_health()
    {
        echo "\n";
        if (!$this->debug_health) {
            return ;
        }
        echo "[entries num]\t";
        echo count($this->entries) . "\n";
        echo "[php memory usage]\t";
        $usage = $this->human_readable_memory_usage(memory_get_usage());
        $os_usage = $this->human_readable_memory_usage(memory_get_usage(true));
        echo "real usage: $usage \t";
        echo "os usage: $os_usage \n";
        if ($this->health_notify) {
            $func = $this->health_notify;
            $info = $func();
            $str = [];
            if (is_array($info)) {
                foreach ($info as $key => $val) {
                    $str[] = "[redis]\t$key:$val";
                }
                $str = implode("\n", $str) . "\n";
            } else {
                $str = $info;
            }
            echo $str;
        }
    }

    private function human_readable_memory_usage($usage)
    {
        if ($usage > 1024 && $usage < 1024 * 1024) {
            $usage /= 1024;
            $usage = number_format($usage, 2);
            $usage .= " Kb";
        } elseif ($usage > 1024 * 1024) {
            $usage /= (1024 * 1024);
            $usage = number_format($usage, 2);
            $usage .= " Mb";
        }
        return $usage;
    }

    private function debug_request()
    {
        echo "\n";
        if (!$this->debug_request_verbose && !$this->debug_request) {
            return ;
        }
        if (!$req = $this->reqinfo()) {
            echo "[request NULL]\n";
        } else {
            echo "[request url]" . $req['url'] . "\n";
            echo "[request method]" . $req['method'] . "\n";
            if ($this->debug_request_verbose) {
                echo "\n";
                foreach ($req['header'] as $h) {
                    echo "[request header]" . $h . "\n";
                }
                echo "\n[request post_fields]";
                if (is_array($req['post_fields'])) {
                    $str = [];
                    foreach ($req['post_fields'] as $name => $val) {
                        $str[] = $name . "=" . $val;
                    }
                    $str = implode("&", $str);
                } else {
                    $str = $req['post_fields'];
                }
                echo $str . "\n";
            }
        }
    }

    private function add_cookies(array $cookies)
    {
        $rebuild = [];
        foreach ($this->cookies as $c) {
            $rebuild[$c['domain']][$c['Name']] = $c;
        }
        $to_add = [];
        foreach ($cookies as $c) {
            if (isset($rebuild[$c['domain']][$c['Name']])) {
                continue;
            }
            $to_add[] = $c;
        }
        $this->cookies = array_merge($this->cookies, $cookies);
    }

    private function ini_reponse()
    {
        return [
            "code" => null,
            "header" => [],
            "cookies" => [],
            "body" => null
            ];
    }

    private function ini_request()
    {
        return [
            "url" => null,
            "method" => null,
            "header" => [],
            "post_fields" => null,
            "curl_options" => [],
            "cookies" => []
        ];
    }

    private function pop_url()
    {
        if ($this->entries) {
            return array_pop($this->entries);
        }
        if ($this->url_queue) {
            $func = $this->url_queue;
            $url = $func("POP");
            if (is_string($url)) {
                $url = ["url" => $url, "referer" => null];
            }
            return $url;
        }
        return false;
    }

    private function put_entries(array $urls)
    {
        foreach ($urls as $url) {
            /*
             * 必须是完整的 url 
             */
            $uinfo = parse_url($url);
            if (isset($uinfo['host']) && isset($uinfo['scheme'])) {
                $this->entries[] = ["url" => $url, "referer" => null];
            }
        }
    }

}
