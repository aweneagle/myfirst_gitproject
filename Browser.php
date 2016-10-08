<?php

$b = new Browser();
$b->load_actions("./Baidu.header");
if (false === $b->view("baidu")) {
	echo $b->error() . "\n";
} else {
	print_r($b->response);
}



class Browser
{

	private $error;

	public function error()
	{
		return $this->error;
	}

	/*
	 * $cookies 
	 * [
	 * 	 ".google.com"=> [ 
	 * 	 	"PID" => [
	 * 	 		"Name" => "PID",
	 * 	 		"Value" => "09fa9af",
	 * 	 		"Expires" => "2016/09/10 10:00",
	 * 	 	 ]
	 * 	  ]
	 * ]
	 */
	private $cookies = [];

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
	private $request = [];


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
	 * $actions, 请求内容, act_name => $reqeust,  $request 格式与上文一致
	 * [
	 * 	 act_name => $request
	 * ]
	 */
	private $actions = [];

	private function load_actions_file($conf_path)
	{
		$fp = fopen($conf_path, "r");
		if (!$fp) {
			return [];
		}
		$mod_list = [];
		$mod = [];
		while (!feof($fp)) {
			if (!$line = trim(fgets($fp))) {
				continue;
			}
			/*
			 * 过滤掉注释
			 */
			if (strpos($line, "#") === 0) {
				continue;
			}
			if (preg_match('/^\[([\w-_]+)\]$/', $line, $match)) {
				if ($mod) {
					$mod_list[] = $mod;
				}
				$mod['mod'] = $match[1];
				$mod['values'] = [];
			} else {
				$mod['values'][] = $line;
			}
		}
		if ($mod) {
			$mod_list[] = $mod;
		}
		fclose($fp);
		return $mod_list;
	}


	/*
	 * prepare,  准备一个新的请求实例
	 * @param	$url, 例如"http://www.google.com"
	 * @param	$method, 可以是常规的method, POST, GET, 可以是自定义的HTTP METHOD, 例如 DELETE
	 */
	public function prepare($url, $method = "GET", $req_name = null)
	{
		$request = [];
		$request['url'] = $url;
		$request['method'] = $method;
		$request['curl_options'] = [
			CURLOPT_RETURNTRANSFER => 1,
			CURLOPT_HEADER => 1,
			];
		$request['cookies'] = [];
		$request['post_fields'] = null;
		$request['header'] = [];

		if (!$req_name) {
			$this->request = $request;
		} else {
			$this->actions[$req_name] = $request;
		}
	}

	private function _put_req_fields($req_name, $field_name, $values, $append)
	{
		if (!$req_name) {
			$req = &$this->request;
		} elseif (isset($this->actions[$req_name])) {
			$req = &$this->actions[$req_name];
		} else {
			return false;
		}
		if ($append) {
			if (is_array($values)) {
				if (!isset($req[$field_name])) {
					$req[$field_name] = [];
				}
				$req[$field_name] = array_merge($req[$field_name], $values);
			} else {
				$req[$field_name] = $req[$field_name] . $values;
			}
		} else {
			$req[$field_name] = $values;
		}
		return true;
	}

	/*
	 * set_request_header,  设置请求实例的 请求头
	 * @param	$header,	一行一个请求头,例如:
	 * 			[
	 *				"Host: play.google.com",
	 *				"User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:49.0) Gecko/20100101 Firefox/49.0",
	 *				"Accept: text/html,application/xhtml+xml,application/xml;q=0.9;q=0.8",
	 *				"Accept-Language: zh-TW,zh;q=0.8,en-US;q=0.5,en;q=0.3",
	 *				"Upgrade-Insecure-Requests: 1",
	 *			]
	 *
	 * @param	$append,   是否追加
	 */
	public function set_request_header(array $header, $req_name = null, $append = false)
	{
		return $this->_put_req_fields($req_name, 'header', $header, $append);
	}

	/*
	 * set_request_postfields,	设置请求实例的post数据
	 * @param	$post_fields,	可以是"Key-Value" 键值对的数组， 也可以是字符串
	 * 			如果是数组, 请起头 Content-Type 将会被设置为:  multipart/form-data
	 * 			postfields 的设置并不受 Http method 的影响， 即使 $method = "GET", postfields 一样可以设置且生效
	 * @param	$append,   是否追加; 如果post_field为字符串，则直接追加在原字符串上
	 */
	public function set_request_postfields($post_fields, $req_name = null, $append = false)
	{
		return $this->_put_req_fields($req_name, 'post_fields', $post_fields, $append);
	}

	/*
	 * set_request_cookies,  设置请求实例的cookie
	 * @param	$cookies,	"Key-Value"键值对的数据， 即 "cookie名" => "cookie值"
	 */
	public function set_request_cookies(array $cookies, $req_name = null, $append = false)
	{
		return $this->_put_req_fields($req_name, 'cookies', $cookies, $append);
	}

	/*
	 * set_request_curloptions,	设置请求实例的curl参数;  
	 * 		该方法的curl参数设置优先级最低,不会覆盖其他方法对curl参数的设置
	 */
	public function set_request_curloptions(array $curl_options, $req_name = null)
	{
		return $this->_put_req_fields($req_name, 'curl_options', $cookies, $append);
	}

	public function load_actions($conf_path)
	{
		$mod_list = $this->load_actions_file($conf_path);
		$req = [];
		$mod_name = null;
		foreach ($mod_list as $values) {
			$name = $values['mod'];
			$values = $values['values'];
			switch ($name) {
			case "name":
				$mod_name = $values[0];
				break;

			case "post_fields":
				$req[$mod_name]['post_fields'] = implode("&", $values);
				$this->set_request_postfields(implode("&", $values), $mod_name, false);
				break;

			case "http":
				//$req[$mod_name]['version'] = isset($values[2]) ? $values[2] : 'HTTP/1.0';
				$this->prepare($values[0], isset($values[1]) ? $values[1] : 'GET', $mod_name);
				break;

			case "header":
				$this->set_request_header($values, $mod_name, true);
				break;

			case "cookies":
				$this->set_request_cookies($values, $mod_name, true);
				break;
			}
		}
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

	private function parse($value, array $params)
	{
		if (is_array($value)) {
			foreach ($value as $k => $v) {
				$value[$k] = $this->parse($v, $params);
			}
		} elseif (is_string($value)) {
			$value = $this->replace('/\{%([\w-_]+)%\}/', $value, $params);
			$value = $this->replace('/\{\$cookie\.([\w-_]+)\}/', $value, $this->cookies);
		}
		return $value;
	}

	public function view($req_name, $params = array())
	{
		$this->request = $this->response = [];
		if (!isset($this->actions[$req_name])) {
			$this->error = "no request of name '$req_name' found";
			return false;
		}
		$url = $this->parse($this->actions[$req_name]['url'], $params);
		$method = $this->parse($this->actions[$req_name]['method'], $params);
		$header = $this->parse($this->actions[$req_name]['header'], $params);
		$post_fields = $this->parse($this->actions[$req_name]['post_fields'], $params);
		$cookies = $this->parse($this->actions[$req_name]['cookies'], $params);
		$this->request['url'] = $url;
		$this->request['method'] = $method;
		$this->request['header'] = $header;
		$this->request['post_fields'] = $post_fields;
		$this->request['cookies'] = $cookies;
		$this->request['curl_options'] = $this->actions[$req_name]['curl_options'];

		return $this->request();
	}

	private function _error($ch)
	{
		return "[ERRNO:" . curl_errno($ch) . "][ERROR]:" . curl_error($ch);
	}

	public function request()
	{
		$this->error = null;
		$this->response = [];
		$this->response['version'] = null;
		$this->response['code'] = null;
		$this->response['cookies'] = [];
		$this->response['header'] = [];
		$this->response['body'] = null;

		$url = $this->request['url'];
		$method = $this->request['method'];
		$header = $this->request['header'];
		$post_fields = $this->request['post_fields'];
		$cookies = $this->request['cookies'];
		$options = $this->request['curl_options'];

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

		if ($cookies) {
			$header = 'Set-Cookie: ';
			foreach ($cookies as $key => $val) {
				if ($val) {
					$header .= $key . "=" . $val . "; ";
				} else {
					$header .= $key . "; ";
				}
			}
			$options[CURLOPT_HTTPHEADER][] = $header;
		}

		curl_setopt_array($ch, $options);

		$res = curl_exec($ch);
		if (false === $res) {
			$this->error = $this->_error($ch);
			curl_close($ch);
			return false;
		} else {
			$return = $ret;
			$this->response['code'] = curl_getinfo($ch, CURLINFO_HTTP_CODE);
			if (!isset($options[CURLOPT_HEADER]) || !$options[CURLOPT_HEADER]) {
				$this->response['body'] = $res;
			} else {
				$header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
				$this->response['header'] = array_filter(explode("\r\n", substr($res, 0, $header_size)));

				$first_header = array_shift($this->response['header']);
				if (!preg_match('/^(HTTP\/[\d\.]+)\s/', $first_header, $http_version)) {
					$error = "[ERROR] wrong header:" . $first_header;
					/* don't return right now , wait for the following steps to be finished
					 */
					$return = false;
				}
				foreach ($this->response['header']as $header) {
					if (preg_match('/^set-cookie\s*:\s*(.*)$/i', $header, $match)) {

						$ck_agv = [];
						foreach (explode(";", $match[1]) as $pair) {
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
							case "httpOnly":
							case "max-age":
								$ck_agv[$key] = $val;
								break;

							default:
								$ck_agv['Name'] = $key;
								$ck_agv['Value'] = $val;
								break;
							}
						}
						if (isset($ck_agv['Name'])) {
							$this->response['cookies'][$ck_agv['Name']] = $ck_agv;
						}
					}
				}
				$this->response['version']= $http_version[1];
				$this->response['body']= substr($res, $header_size);
			}
			curl_close($ch);
			return $return;
		}
	}
}
