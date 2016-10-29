<?php
/* 
 * ES, ElasticSearch 
 *
 * a easy tool for es operation
 *
 * 使用示例：
 *  查询 商店中作者 "king" 和 "bible" 两人价格为 50~100的
 *  含有"Elasticsearch" 的书
 *
 * $es = new ES;
 * $es->index("store")
 *    ->type("books")
 *    ->must(function(ES $es) {
 *      $es->where("author", "in", ["king", "bible"]);
 *      $es->must(function(ES $es) {
 *          $es->where("price", "<=", 100);
 *          $es->where("price", ">=", 50);
 *      });
 *    })
 *    ->should(function(ES $es) {
 *      $es->match("title", "Elasticsearch", 2);    //标题最重要
 *      $es->match("subtitle", "Elasticsearch");    //子标题其次
 *      $es->should(function(ES $es) {
 *          $es->match("desc", "Elasticsearch");    //描述 和 内容没那么重要
 *          $es->match("content", "Elasticsearch");
 *      });
 *    })
 *    ->sort_score()             //按照内容重要性排序
 *    ->sort('date', 'desc')    //再按发布时间排序
 *    ->search();               //查询数据
 *
 *
 */

class ES
{
    public $host = '127.0.0.1';
    public $port = 9200;

	private $error = null;

    private $index = null;
    private $type = null;

    private $query = [];
    private $sort = [];
    private $filter = [];
    private $cache = [];

    //标记当前的 must/should/must_not 所处的上下文
    //null 为初始化, "*" 为未知, "filter", "query" 分别对应过滤和查询
    private $context = null;   
    private $context_level = 0;


    /*
     * index() 选择索引
     */
    public function index($index_name = null)
    {
        $this->clean_query();
        $this->index = $index_name;
        return $this;
    }

    /*
     * type() 选择type
     */
    public function type($type_name = null)
    {
        $this->type = $type_name;
        return $this;
    }

	/*
	 * error() 查看错误信息
	 */
	public function error()
	{
		return $this->error;
	}

    /* 
     * search() 根据$query结构体进行查询
     *
     * @return false, 查询语句或者数据访问有错误, 调用 error() 查看； 成功返回数组, 如果没有被匹配到，返回 []
     */
    public function search($query = null)
    {
		$path = '';
		if ($this->index) {
			$path .= $this->index . "/";
		}
		if ($this->type) {
			$path .= $this->type . "/";
		}
        $url = "http://" . $this->host . ":" . $this->port . "/" . $path . "_search";
        if ($query) {
            $post_fields = json_encode($query);
        } else {
            $post_fields = json_encode($this->to_query());
        }
        $this->clean_query();
        $data = $this->curl($url, "GET", $post_fields);
		if ($data === false) {
			return false;
		}
        $data = json_decode($data, true);
        return $this->format_output($data);
    }

	private function format_output($data)
	{
		$return = [];
		if (isset($data['hits']['hits'])) {
			foreach ($data['hits']['hits'] as $d) {
				$return[] = $d['_source'];
			}
		}
		return $return;
	}

    /*
     * must() 对应 ES 的must查询语句, 相当于逻辑 AND
     *
     * @param   $func, 原型为 function(Es $es)
     * @return  null;
     */
    public function must(Closure $func)
    {
        $this->before("must");
        $func($this);
        $this->after();
        return $this;
    }

    /*
     * should() 对应 ES 的should查询语句, 相当于逻辑 OR
     *
     * @param   $func, 原型为 function(Es $es)
     * @return  null;
     */
    public function should(Closure $func)
    {
        $this->before("should");
        $func($this);
        $this->after();
        return $this;
    }

    /*
     * must_not() 对应 ES 的must_not查询语句, 相当于逻辑 AND NOT
     *
     * @param   $func, 原型为 function(Es $es)
     * @return  null;
     */
    public function must_not(Closure $func)
    {
        $this->before("must_not");
        $func($this);
        $this->after();
        return $this;
    }

    /*
     * where() 对应 ES 的term, range, 和 terms语句
     *      $op = "=",  对应 term, 注意当字段是个多值字段时，term的意思将变成“包含”而不是“等于”
     *      $op = "<", ">", "<=", ">=" 对应 range
     *      $op = "in", 对应 terms
     *
     * @param   $field, 字段名
     * @param   $op,    比较操作，有 "=", ">=", "<=", ">", "<", 和 "in"
     * @param   $value, 字段值
     */
    public function where($field, $op, $value)
    {
        $this->try_to_switch_context("filter", "must");
        $this->try_to_push_context("filter", "where");
        $filter = [];
        switch ($op) {
        case "=":
            $filter["term"] = [$field => $value];
            break;

        case "in":
            if (!is_array($value)) {
                throw new \Exception(" 'in' operation should use an array as 'value'");
            }
            $filter["terms"] = [$field => $value];
            break;

        case ">=":
            if (!is_numeric($value)) {
                throw new \Exception(" '>=' operation should use number as 'value'");
            }
            $filter["range"] = [$field => ["gte" => $value]];
            break;

        case "<=":
            if (!is_numeric($value)) {
                throw new \Exception(" '<=' operation should use number as 'value'");
            }
            $filter["range"] = [$field => ["lte" => $value]];
            break;

        case "<":
            if (!is_numeric($value)) {
                throw new \Exception(" '<' operation should use number as 'value'");
            }
            $filter["range"] = [$field => ["lt" => $value]];
            break;

        case ">":
            if (!is_numeric($value)) {
                throw new \Exception(" '>' operation should use number as 'value'");
            }
            $filter["range"] = [$field => ["gt" => $value]];
            break;

        default :
            throw new \Exception(" unknown operation '$op'");

        }
        $this->cache[] = $filter;
        $this->pop_context();
        return $this;
    }

    /*
     * exists() 对应 ES 的 exists 和 missing
     *
     * @param   $field, 字段名
     * @param   $exists,  true or false
     */
    public function exists($field, $exists = true)
    {
        $this->try_to_switch_context("filter", "must");
        $this->try_to_push_context("filter", "exists");
        if ($exists) {
            $filter = ["exists" => ["field" => $field]];
        } else {
            $filter = ["missing" => ["field" => $field]];
        }
        $this->cache[] = $filter;
        $this->pop_context();
        return $this;
    }

    /*
     * match() 对应 ES 的match语句
     *
     * @param   $field, 字段名
     * @param   $keywords,  查询关键字
     * @param   $boost, 权重, 默认为1
     */
    public function match($field, $keywords, $boost = 1)
    {
        $this->try_to_switch_context("query", "should");
        $this->try_to_push_context("query", "match");
        if ($boost == 1) {
            $condition = $keywords;
        } else {
            $condition = ["query" => $keywords, "boost" => $boost];
        }
        $query = [
            "match" => [$field => $condition]
        ];
        $this->cache[] = $query;
        $this->pop_context();
        return $this;
    }

    /*
     * sort() 根据字段排序
     * 要注意使用多个字段排序时，是会根据 sort() 调用的顺序来确定 字段的优先级的
     * 例如: ->sort("date")->sort("price") 会先按照 "date" 排序，当 date 相等时，再按照 "price" 排序
     *
     * @param   $field, 字段名
     * @param   $order, 排序方式，"desc" 为倒序, "asc" 为升序, 不区分大小写
     */
    public function sort($field, $order = "desc")
    {
        $this->sort[] = [$field => ["order" => $order]];
        return $this;
    }

    /*
     * sort_score() 启用分数排序
     * 相当于 sort("_score", "desc")
     */
    public function sort_score()
    {
        $this->sort("_score", "desc");
        return $this;
    }


    /*
     * to_query() 返回query数据，用于debug
     */
    public function to_query()
    {
        $this->after();
        $query = [];
        if (!empty($this->query)) {
            if (!empty($this->filter)) {
                $query = [
                    "query" => [
                        "filtered" => [
                            "query" => $this->query,
                            "filter" => $this->filter,
                        ],
                    ]
                ];
            } else {
                $query = [
                    "query" => $this->query
                ];
            }
        } else {
            if (!empty($this->filter)) {
                $query = [
                    "filter" => $this->filter,
                ];
            } else {
                $query = [
                    "query" => ["match_all" => []]
                ];
            }
        }
        if (!empty($this->sort)) {
            $query['sort'] = $this->sort;
        }
        return $query;
    }

    private function try_to_push_context($tag, $func_name)
    {
        if ($this->context != "*" && $this->context != null) {
            if ($tag != "*" && $tag != $this->context) {
                throw new \Exception("could not call $func_name() in a $this->context context");
            }
        }
        if ($this->context == "*" || $this->context == null) {
            $this->context = $tag;
        }
        $this->context_level += 1;
    }

    private function build_query_from_cache(array &$cache)
    {
        $query = [];
        while (!empty($cache)) {
            $op = array_shift($cache);
            if (in_array($op, ["must", "must_not", "should"])) {
                $query[] = ["bool" => [$op => $this->build_query_from_cache($cache)]];
            } elseif ($op == "end") {
                return $query;
            } else {
                $query[] = $op;
            } 
        }
        if (empty($query)) {
            return [];
        }
        return $query[0];
    }

	private function merge_bool($merge_from, $merge_into)
	{
		if (isset($merge_from['bool'])) {
			foreach ($merge_from['bool'] as $bool => $info) {
				if (!isset($merge_into['bool'][$bool])) {
					$merge_into['bool'][$bool] = [];
				}
				$merge_into['bool'][$bool] = array_merge($merge_into['bool'][$bool], $info);
			}
		}
		return $merge_into;
	}

    private function pop_context()
    {
        $this->context_level -= 1;
        if ($this->context_level == 0) {
            switch ($this->context) {
            case "query":
				$this->query = $this->merge_bool(
					$this->build_query_from_cache($this->cache), 
					$this->query
				);
                break;

            case "filter":
				$this->filter = $this->merge_bool(
					$this->build_query_from_cache($this->cache, $this->filter),
					$this->filter
				);
                break;
            }

            $this->context = null;
        }
    }

    private function clean_query()
    {
        $this->query = $this->filter = $this->sort = [];
        $this->cache = [];
        $this->context = null;
        $this->context_level = 0;
    }

    private function before($bool)
    {
        $this->try_to_push_context("*", $bool);
        $this->cache[] = $bool;
    }

    private function after()
    {
        if ($this->context_level) {
            $this->cache[] = "end";
            $this->pop_context();
        }
    }

    private function try_to_switch_context($to_context, $bool)
    {
        $curr_ctx = $this->context;
        if ($curr_ctx != "*" && $curr_ctx != $to_context && $curr_ctx != null) {
            $this->after();
			$this->before($bool);
        }
        if ($curr_ctx == null) {
            $this->before($bool);
        }

    }

	private function curl($url, $method, $post_fields)
	{
		$ch = curl_init($url);
		if (!$ch) {
			$this->set_error("failed curl_init('$url')");
			return false;
		}
		$curl_options = [
			CURLOPT_RETURNTRANSFER => 1,
		];
		$this->set_method($curl_options, $method);
		$curl_options[CURLOPT_POSTFIELDS] = $post_fields;
		if (curl_setopt_array($ch, $curl_options) === false) {
			$this->set_error("failed curl_setopt_array(" . json_encode($curl_options) . ")");
			curl_close($ch);
			return false;
		}
		$res = curl_exec($ch);
		if ($res === false) {
			$this->error = "failed curl_exec($url). error:" . curl_error($ch);
			return false;
		}
		return $res;
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

	private function set_error($errmsg)
	{
		$this->error = $errmsg;
		return false;
	}
}
