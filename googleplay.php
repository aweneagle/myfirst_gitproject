<?php
require "./simple_html_dom.php";

$c = fetch_categories(file_get_contents("./category.html"));
var_dump($c);
//var_dump(run());
die;

//$info = fetch_packageinfo("@./category.html", $finished);
//$info = fetch_packageinfo("@./topselling.html", $finished);
//var_dump($info);
//var_dump($finished);

function run()
{
	$b = new Browser("./RequestChain.headers");
	$b->view("store");
	print_res($b);
	$b->view("SCookie");
	print_res($b);

	$c = [
		"name" => "GAME_CASUAL",
		];

	$b->view("Category", 
		[
		"CATEGORY" => $c["name"],
		]);
	print_res($b);

	$start = 0;
	$info = fetch_packageinfo($b->response['body']);
	$start += count($info);
	$NUM = 40;
	$finished = false;
	while (!$finished) {
		$b->view("topselling_free", [
			"CATEGORY" => $c["name"],
			"start" => $start,
			"num" => $NUM,
			]);
		print_res($b);
		$new = fetch_packageinfo($b->response['body']);
		/*
		 * google play 翻页到尽头时，无论 $start 如何增长， 都会返回最后 一批$NUM 数据
		 */
		$diff = array_diff(array_keys($new), array_keys($info));
		$finished = empty($diff);
		$info = array_merge($info, $new);
		$start += count($new);
		sleep(2);
	}
	return $info;
}


//$b = new Browser("./RequestChain.headers");
//$b->view("store");
//$b->view("SCookie");
//$b->view("Category", [
//	"CATEGORY" => "GAME_CASUAL"
//	]);
//file_put_contents("./category.html", $b->response["body"]);
//print_res($b);
//
//$b->view("topselling_free", [
//	"CATEGORY" => "GAME_CASUAL",
//	"start" => 40,
//	"num" => 40,
//	]);
//file_put_contents("./topselling.html", $b->response["body"]);
//print_res($b);
		
function fetch_categories($page)
{
	$html = str_get_html($page);
	if (!$html) {
		return false;
	}
	if (!$links = $html->find("a.child-submenu-link")) {
		return false;
	}

	$categories = [];
	foreach ($links as $a) {
		$href = $a->getAttribute("href");
		$categories[] = [
			'href' =>  $href,
			'category' => substr($href, strrpos($href, "/") + 1),
			];
	}
	return $categories;
}

function fetch_packageinfo($page)
{
	$html = str_get_html($page);
	if (!$html) {
		return false;
	}
	$container = $html->find("div.id-card-list div.apps");
	if (!$container) {
		return false;
	}
	$info = [];
	foreach ($container as $d) {
		$pkname = $d->getAttribute("data-docid");
		$info[$pkname]["href"] = $d->find("a.card-click-target")[0]->href;
	}
	return $info;
}

function print_res($b)
{
	$response = $b->response;
	unset($response['body']);
	print_r(["request" => $b->request, "response" => $response]);
}

class Browser
{
	public $cookies = [];
	public $request = [];
	public $response = [];


	/*
	 * view
	 * [
	 * 	 $name => [
	 * 	 	"header" => [
	 * 	 		'...'
	 * 	 	],
	 * 	 	"method" => 'POST',
	 * 	 	"post_fields" => 'a=1&b=2'
	 * 	 ]
	 * ]
	 */
	public $views = [];

	private function load_mods($conf_path)
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

	public function __construct($conf_path)
	{
		$mod_list = $this->load_mods($conf_path);
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
				break;

			case "http":
				$req[$mod_name]['http']['url'] = $values[0];
				$req[$mod_name]['http']['method'] = isset($values[1]) ? $values[1] : 'GET';
				$req[$mod_name]['http']['version'] = isset($values[2]) ? $values[2] : 'HTTP/1.0';
				break;

			case "headers":
				$req[$mod_name]['header'] = $values;
				break;
			}
		}
		$this->views = $req;
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

	public function view($view_name, $params = array())
	{
		$this->request = $this->response = [];
		if (!isset($this->views[$view_name])) {
			return;
		}
		$url = $this->parse($this->views[$view_name]['http']['url'], $params);
		$method = $this->parse($this->views[$view_name]['http']['method'], $params);
		$header = $this->parse($this->views[$view_name]['header'], $params);
		$post_fields = $this->parse(@$this->views[$view_name]['post_fields'], $params);
		$this->request['url'] = $url;
		$this->request['method'] = $method;
		$this->request['header'] = $header;
		$this->request['post_fields'] = $post_fields;

		$c = new Curl;
		$c->method = $method;
		$c->header = $header;
		if ($post_fields) {
			$c->post_fields = $post_fields;
		}
		$c->exec($url);
		foreach ($c->response_cookie() as $name => $info) {
			$this->cookies[$name] = $info['Value'];
		}
		$this->response['cookie'] = $c->response_cookie();
		$this->response['code'] = $c->response_code();
		$this->response['header'] = $c->response_header();
		$this->response['body'] = $c->response_body();
	}
}

class Curl
{
	/* 
	 * response fields 
	 */
	private $response_version = null;
	private $response_code = null;
	private $response_body = null;
	private $response_header = [];
	private $response_cookie = [];
	private $error = null;

	/*
	 * request fields
	 */
	public $post_fields = null;
	public $cookie = [];
	public $method = "GET";
	public $header = [];
	public $options = [
		CURLOPT_RETURNTRANSFER => 1,
		CURLOPT_HEADER => 1,
	];

	public function response_version()
	{
		return $this->response_version;
	}

	public function response_code()
	{
		return $this->response_code;
	}

	public function response_body()
	{
		return $this->response_body;
	}

	public function response_header()
	{
		return $this->response_header;
	}

	public function response_cookie()
	{
		return $this->response_cookie;
	}

	public function error()
	{
		return $this->error;
	}

	private function _error($ch)
	{
		return "[ERRNO:" . curl_errno($ch) . "][ERROR]:" . curl_error($ch);
	}

	public function exec($url)
	{
		$this->error = null;
		$this->response_header = [];
		$this->response_cookie = [];
		$this->response_body = [];
		$this->response_code = null;

		$options = $this->options;

		if (!$ch = curl_init($url)) {
			$this->error = "[ERROR] failed curl_init(), url=$url";
			return false;
		}
		if ($this->header) {
			foreach ($this->header as $h) {
				$options[CURLOPT_HTTPHEADER][] = $h;
				if (preg_match('/^Accept-Encoding\s*:\s*(.+)$/', $h, $match)) {
					$options[CURLOPT_ENCODING] = $match[1];
				}
			}
		}
		$method = strtoupper($this->method);
		switch ($method) {
		case "POST":
			$options[CURLOPT_POST] = 1;
			break;

		default:
			$options[CURLOPT_CUSTOMREQUEST] = $method;
			break;
		}

		if ($this->post_fields) {
			$options[CURLOPT_POSTFIELDS] = $this->post_fields;
			if (is_array($this->post_fields)) {
				$str = [];
				foreach ($this->post_fields as $key => $val) {
					$str[] = $key . '=' . $val;
				}
				$str = implode("&", $str);
			} else {
				$str = $this->post_fields;
			}
			$options[CURLOPT_HTTPHEADER][] = "Content-Length: " . strlen($str);
		}

		if ($this->cookie) {
			$header = 'Set-Cookie: ';
			foreach ($this->cookie as $key => $val) {
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
			$this->response_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
			if (!isset($options[CURLOPT_HEADER]) || !$options[CURLOPT_HEADER]) {
				$this->response_body = $res;
			} else {
				$header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
				$this->response_header = array_filter(explode("\r\n", substr($res, 0, $header_size)));

				$first_header = array_shift($this->response_header);
				if (!preg_match('/^(HTTP\/[\d\.]+)\s/', $first_header, $http_version)) {
					$this->error = "[ERROR] wrong header:" . $first_header;
					/* don't return right now , wait for the following steps to be finished
					 */
					$return = false;
				}
				foreach ($this->response_header as $header) {
					if (preg_match('/^set-cookie\s*:\s*(.*)$/i', $header, $match)) {

						$ck_agv = [];
						foreach (explode(";", $match[1]) as $cookie) {
							$cookie = trim($cookie);
							$pos = strpos($cookie, "=");
							if ($pos === false) {
								$key = $cookie;
								$val = '';
							} else {
								$key = substr($cookie, 0, $pos);
								$val = substr($cookie, $pos + 1);
							}
							switch ($key) {
							case "Path":
							case "Expires":
							case "Domain":
							case "Secure":
							case "HttpOnly":
							case "Max-Age":
								$ck_agv[$key] = $val;
								break;

							default:
								$ck_agv['Name'] = $key;
								$ck_agv['Value'] = $val;
								break;
							}
						}
						if (isset($ck_agv['Name'])) {
							$this->response_cookie[$ck_agv['Name']] = $ck_agv;
						}
					}
				}
				$this->response_version = $http_version[1];
				$this->response_body = substr($res, $header_size);
			}
			curl_close($ch);
			return $return;
		}
	}
}
