<?php

class Spider
{

	/* 
	 * run() 自动爬去数据时，一次multi_curl的url数
	 */
	public $batch_num = 6;


	public $nobody = false;


    /*
     * 错误处理函数
     * function on_error($request, $response), 函数无需返回
     */
    public $on_error = null;

    /*
     * 页面处理函数
     * function on_succ($request, $response), 函数无需返回
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
	 * 使用 {% 和 %} 将  能表示Date()格式的字符串括起来，可以自动切割日志
	 * {% %} 中能接受的字符有： 0-9, a-z, A-Z, - : _
	 * $log_auto_splited = [
	 * 		log_path => real_path_of_log_after_splited,
	 * 		...
	 * ]
	 */
	private $log_auto_splited = [];

	/*
	 * 示例: $debug_log = "./error_{%Ymd%}.log";
	 */
	public $debug_log = null;

	/*
	 * 示例: $error_log = "./error_{%Ymd%}.log";
	 */
	public $error_log = null;

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
	 * view() 请求一个url, 并将结果以 response 的格式返回 
	 *
	 * @param	url,  可以是 string, 也可以是 url 数组
	 *
	 * @return	url为string时，返回 response
	 * 			url为数组时，  返回 ［url => response]  
	 */
	public function view($url, $method = "GET", $header = [], $post_fields = null, $referer = null, $curl_options = [])
	{
		if (is_string($url)) {
			$request = [$url];
		} else {
			$request = $url;
		}
		$response = [];
		foreach ($request as &$u) {
			$u = [
					"url" => $u,
					"method" => $method,
					"header" => $header,
					"post_fields" => $post_fields,
					"referer" => $referer,
					"curl_options" => $curl_options
				];
		}
		$this->request($request, $response);
		foreach ($response as $resp) {
			if ($resp) {
				$this->add_cookies($resp["cookies"]);
			}
		}
		if (is_string($url)) {
			return $response[$url];
		} else {
			return $response;
		}
	}

    /*
     * cookies() 根据所访问的域名来取对应的cookies
     */
    public function cookies($host = null)
    {
        if ($host == null) {
            return $this->cookies;
        } else {
			$domain1 = $host;
			$domain2 = substr($host, strpos($host, "."));
			$ck = [];
			foreach ([$domain1, $domain2] as $d) {
				if (isset($this->cookies[$d])) {
					$ck = array_merge($ck, array_values($this->cookies[$d]));
				}
			}
			return $ck;
        }   
    }

    /*
     * @param  $request = [
     *      [
     *          "url" => null,
     *          "method" => "GET",
     *          "header" => [...],
     *          "post_fields" => [...],
	 *			"referer" => "..",
     *          "curl_options" => [...],
     *      ],
     *      ...
     *  ]
	 *
	 *  @return  &$response = [
	 *  	url  =>  [
	 *  		"code" => 200,
	 *  		"header" => [...],
	 *  		"body" => "..",
	 *  		"version" => "HTTP/1.1",
	 *  		"cookies" => [...]
	 *  	],
	 *  	...
	 *  ]
     */
    public function request(array $request, array &$response = []) {
		foreach ($request as $i => $u) {
			if (!$this->check_and_fix_urlinfo($u)) {
				return;
			}
			$request[$i] = $u;
		}

		/* 初始化curl */
        $curl_pool = [];
		$request_pool = [];
        foreach ($request as $u) {
            $h = curl_init();
            if ($h) {
                $this->init_curl($h, $u);
                $curl_pool[$u["url"]] = $h;
				/* 仅仅做url映射 */
				$request_pool[$u["url"]] = &$u;
			} else {
				$this->error($u["url"] . " failed to curl_init()");
			}
        }

		/* 批量请求 & 处理结果 */
        $mh = curl_multi_init();
        $this->multi_curl($mh, $curl_pool);
        foreach ($curl_pool as $u => $h) {
            $content = curl_multi_getcontent($h);
            $resp = $this->parse_response($u, $content, $h);
			$this->debug_response($resp);

			/* curl 失败处理 */
			if ($content === false) {
				$response[$u] = false;
				if ($this->on_error) {
					$func = $this->on_error;
					$func($request_pool[$u], $resp);
				}

			/* curl 成功*/
			} else {
				$response[$u] = $resp;
				if ($this->on_succ) {
					$func = $this->on_succ;
					$func($request_pool[$u], $resp);
				}
			}
        }

		/* 回收资源 */
		$this->clean_multi_curl($mh, $curl_pool);
        foreach ($curl_pool as $h) {
            curl_close($h);
        }
        curl_multi_close($mh);
    }

	private function curl_error($ch)
	{
		return "[ERRNO: " . curl_errno($ch) . "][ERRMSG]:" . curl_error($ch);
	}

    private function parse_response($url, $content, $curl)
	{
		$this->parse_url($url, $scheme, $host, $path, $query);
		//如果content === false, 打error日志
		if ($content === false) {
			return $this->error($this->curl_error($curl));
		}
		//  根据size, 取出header
		$response = ['url' => $url];
        $response['code'] = curl_getinfo($curl, CURLINFO_HTTP_CODE);

		//从curl_handle取请求头的size
		$header_size = curl_getinfo($curl, CURLINFO_HEADER_SIZE);
		$response['header'] = array_filter(explode("\r\n", substr($content, 0, $header_size)));

		//  根据size, 取出body
		if (!$this->nobody) {
			$response['body'] = substr($content, $header_size);
		}

		//  根据header, 取出code, httpversion
		$first_header = array_shift($response['header']);
		if (!preg_match('/^(HTTP\/[\d\.]+)\s/', $first_header, $http_version)) {
			return $this->error("wrong response header:" . $first_header);
		}

		foreach ($response['header'] as $header) {
			if (preg_match('/^set-cookie\s*:\s*(.*)$/i', $header, $match)) {
				$ck_agv = $this->cookie_str2arr($match[1], $host);
				if (isset($ck_agv['Name'])) {
					$response['cookies'][$ck_agv['Name']] = $ck_agv;
				}
			}
		}
		$response['version'] = $http_version[1];

		if (!isset($response['cookies'])) {
			$response['cookies'] = [];
		}
		if (!isset($response['body'])) {
			$response['body'] = null;
		}
		return $response;
	}

	private function clean_multi_curl($multi_curl, array $curl_pool)
	{
		foreach ($curl_pool as $h) {
			curl_multi_remove_handle($multi_curl, $h);
		}
	}

	private function multi_curl($multi_curl, array $curl_pool)
	{
		foreach ($curl_pool as $h) {
			curl_multi_add_handle($multi_curl, $h);
		}
		do {
			curl_multi_exec($multi_curl, $running);
			curl_multi_select($multi_curl);
		} while ($running > 0);
	}

	private function log($file_path, $msg)
	{
		$time = time();
		if (!isset($this->log_auto_splited[$file_path])) {
			if (preg_match_all('/\{%([\w-_:]+)%\}/', $file_path, $match)) {
				$from = $to = [];
				foreach ($match[1] as $m) {
					$from[] = "{%" . $m . "%}";
					$to[] = @Date($m, $time);
				}
				$real_file_path = str_replace($from, $to, $file_path);

			} else {
				$real_file_path = $file_path;
			}
			$this->log_auto_splited[$file_path] = $real_file_path;
		}
		$file_path = $this->log_auto_splited[$file_path];
		file_put_contents($file_path, "[" . @Date("Y-m-d H:i:s", $time) . "]" . $msg . "\n", FILE_APPEND);
	}

	private function debug($msg)
	{
		if ($this->debug_log) {
			$output = $this->debug_log;
		} else {
			$output = "php://stdout";
		}
		$this->log($output, $msg);
	}

	private function error($errmsg)
	{
		$this->error = $errmsg;
		if ($this->error_log) {
			$this->log($this->error_log, $errmsg);
		}
		return false;
	}

	private function check_and_fix_urlinfo(array &$url_info)
	{
		if (!isset($url_info["url"])) {
			return $this->error("no url found in urlinfo:" . json_encode($url_info));
		}
		if (isset($url_info["header"]) && !is_array($url_info["header"])) {
			return $this->error("no header found in urlinfo:" . json_encode($url_info));
		}
		if (isset($url_info["curl_options"]) && !is_array($url_info["curl_options"])) {
			return $this->error("no curl_options found in urlinfo:" . json_encode($url_info));
		}
		if (isset($url_info["referer"]) && !is_string($url_info["referer"])) {
			return $this->error("no curl_options found in urlinfo:" . json_encode($url_info));
		}
		if (isset($url_info["post_fields"]) && !is_array($url_info["post_fields"]) && !is_string($url_info["post_fields"])) {
			return $this->error("no post_fields found in urlinfo:" . json_encode($url_info));
		}

		if (!isset($url_info["method"])) {
			$url_info["method"] = "GET";
		}
		if (!isset($url_info["header"])) {
			$url_info["header"] = [];
		}
		if (!isset($url_info["curl_options"])) {
			$url_info["curl_options"] = [];
		}
		if (!isset($url_info["referer"])) {
			$url_info["referer"] = "";
		}
		if (isset($url_info["post_fields"]) && is_array($url_info["post_fields"])) {
			$url_info["post_fields"] = http_build_query($url_info["post_fields"]);
		}
		if (!isset($url_info["post_fields"])) {
			$url_info["post_fields"] = null;
		}
		return true;
	}

	private function set_method(&$curl_options, $method)
	{
        $method = strtoupper($method);
        switch ($method) {
        case "POST":
            $curl_options[CURLOPT_POST] = 1;
            break;

        default:
            $curl_options[CURLOPT_CUSTOMREQUEST] = $method;
            break;
        }
	}

    private function cookies_header($cookies)
    {
        $pairs = [];
        foreach ($cookies as $value) {
            $pairs[] = $value['Name'] . "=" . $value["Value"];
        }
        return implode("; ", $pairs);
    }

	private function set_header(&$curl_options, array &$header, array $cookies, $content_len, $referer)
	{
		/* cookie 头下标 */
		$ref_header_index = $contentlen_header_index = $ck_header_index = -1;

		foreach ($header as $i => $h) {
			/* set encoding */
			if (preg_match('/^Accept-Encoding\s*:\s*(.+)$/i', $h, $match)) {
				$curl_options[CURLOPT_ENCODING] = $match[1];
			}

			/* set cookie */
			if (preg_match('/^Cookie\s*:\s*.*$/i', $h)) {
				$ck_header_index = $i;
			}

			/* set content length */
			if (preg_match('/^Content-Length\s*:$/i', $h)) {
				$contentlen_header_index = $i;
			}

			/* set referer */
			if (preg_match('/^Referer\s*:$/i', $h)) {
				$ref_header_index = $i;
			}
		}
		if ($cookies) {
			if ($ck_header_index != -1) {
				$header[$ck_header_index] .= "; " . $this->cookies_header($cookies);
			} else {
				$header[] .= "Cookie: " . $this->cookies_header($cookies);
			}
		}
		if ($referer && $ref_header_index == -1) {
			$header[] = "Referer: " . $referer;
		}
		if ($contentlen_header_index != -1) {
			$header[$contentlen_header_index] = "Content-Length: " . $content_len;
		} else {
			$header[] = "Content-Length: " . $content_len;
		}
		$curl_options[CURLOPT_HTTPHEADER] = $header;
	}

	private function parse_url($url, &$scheme = null, &$host = null, &$path = null, &$query = null)
	{
		$url_info = parse_url($url);
		$scheme = isset($url_info["scheme"]) ? $url_info["scheme"] : null;
		$host = isset($url_info["host"]) ? $url_info["host"] : null;
		$path = isset($url_info["path"]) ? $url_info["path"] : null;
		$query = isset($url_info["query"]) ? $url_info["query"] : null;
	}

    private function add_cookies(array $cookies)
    {
        foreach ($cookies as $c) {
            $this->cookies[$c['domain']][$c['Name']] = $c;
        }
    }

    private function init_curl($ch, $url_info)
    {
        $options = [
            CURLOPT_URL => $url_info["url"],
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_HEADER => 1,
        ];
		$this->parse_url(
			$url_info["url"], 
			$scheme, 
			$host, 
			$path, 
			$query);

        // 根据method设置 请求方法
		$this->set_method($options, $url_info["method"]);

        // 设置postfield
		$content_len = 0;
		if ($url_info["post_fields"]) {
        	$options[CURLOPT_POSTFIELDS] = $url_info["post_fields"];
			$content_len = strlen($url_info["post_fields"]);
		}
        // 根据Accept-Encoding header 开启 curl的解压缩功能
		// 根据cookie 设置 Cookie header
        // 根据content length 设置 Content-Length header
		// 根据referer 设置 Referer header
		$this->set_header($options, $url_info["header"], $this->cookies($host), $content_len, $url_info['referer']);
		$this->debug_request($url_info);
        curl_setopt_array($ch, $options);
    }


    /*
     * run() 函数以一定数量的url为入口, 自动爬取页面
     *
     */
    public function run(array $urls, array $header = [])
    {
        $this->push_urls($urls, null);
        while ($entry = $this->pop_urls()) {

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


    private function debug_response($response)
    {
        if (!$this->debug_request_verbose && !$this->debug_request) {
            return ;
        }
        if ($response) {
            $this->debug("[response url]\t" . $response['url']);
            $this->debug("[response code]\t" . $response['code']);
            if ($this->debug_request_verbose) {
                foreach ($response['header'] as $h) {
                    $this->debug("[response header]\t" . $h . "");
                }
            }
        } else {
            $this->debug("[response NULL]");
        }
    }

    private function debug_health()
    {
        if (!$this->debug_health) {
            return ;
        }
        $this->debug("[entries num]");
        $this->debug(count($this->entries));
        $this->debug("[php memory usage]");
        $usage = $this->human_readable_memory_usage(memory_get_usage());
        $os_usage = $this->human_readable_memory_usage(memory_get_usage(true));
        $this->debug("[memory]\treal usage: $usage\tos usage: $os_usage");
        if ($this->health_notify) {
            $func = $this->health_notify;
            $info = $func();
            $str = [];
            if (is_array($info)) {
                foreach ($info as $key => $val) {
                    $str[] = "[redis]\t$key:$val";
                }
                $str = implode("\n", $str);
            } else {
                $str = $info;
            }
            $this->debug($str);
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

    private function debug_request($request)
    {
        if (!$this->debug_request_verbose && !$this->debug_request) {
            return ;
        }
        if (!$request) {
            $this->debug("[request NULL]");
        } else {
            $this->debug("[request url]\t" . $request['url']);
            $this->debug("[request method]\t" . $request['method']);
            if ($this->debug_request_verbose) {
                foreach ($request['header'] as $h) {
                    $this->debug("[request header]\t" . $h);
                }
                $this->debug("[request post_fields]\t");
                if (is_array($request['post_fields'])) {
                    $str = [];
                    foreach ($request['post_fields'] as $name => $val) {
                        $str[] = $name . "=" . $val;
                    }
                    $str = implode("&", $str);
                } else {
                    $str = $request['post_fields'];
                }
                $this->debug($str);
            }
        }
    }

    private function pop_urls()
    {
		$url_list = [];
        if ($this->url_queue) {
            $func = $this->url_queue;
			for ($i = 0; $i < $this->batch_num; $i ++) {
				$url = $func("POP");
				if (is_string($url)) {
					$url = ["url" => $url, "referer" => null];
				}
				$url_list[] = $url;
			}
            return $url_list;
        } else {
			for ($i = 0; $i < $this->batch_num && !empty($this->entries); $i ++) {
				$url_list[] = array_pop($this->entries);
			}
        }
        return $url_list;
    }

    private function push_urls($urls, $referer)
    {
        $new_urls = array_unique($urls);
        /* 添加 referer */
        foreach ($new_urls as $i => $u) {
            $new_urls[$i] = ["url" => $u, "referer" => $referer];
        }
        if ($this->url_queue) {
            $func = $this->url_queue;
            $func("PUSH", $new_urls);
        } else {
            $this->entries = array_merge($this->entries, $new_urls);
        }
    }


}
