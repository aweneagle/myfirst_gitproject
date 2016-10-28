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
 *    ->sort_by_score()             //按照内容重要性排序
 *    ->sort('date', 'desc')    //再按发布时间排序
 *    ->search();               //查询数据
 *
 *
 */

class ES
{
    public $host = '127.0.0.1';
    public $port = 9300;

    private $index = "*";
    private $type = "*";

    /*
     * index() 选择索引
     */
    public function index($index_name = "*")
    {
        $this->index = $index_name;
        return $this;
    }

    /*
     * type() 选择type
     */
    public function type($type_name = "*")
    {
        $this->type = $type_name;
        return $this;
    }

    /* 
     * search() 根据$query结构体进行查询
     *
     * @return false, 查询语句或者数据访问有错误, 调用 error() 查看； 成功返回数组, 如果没有被匹配到，返回 []
     */
    public function search($query = null)
    {
    }

    /*
     * must() 对应 ES 的must查询语句, 相当于逻辑 AND
     *
     * @param   $func, 原型为 function(Es $es)
     * @return  null;
     */
    public function must($func)
    {
    }

    /*
     * should() 对应 ES 的should查询语句, 相当于逻辑 OR
     *
     * @param   $func, 原型为 function(Es $es)
     * @return  null;
     */
    public function should($func)
    {
    }

    /*
     * must_not() 对应 ES 的must_not查询语句, 相当于逻辑 AND NOT
     *
     * @param   $func, 原型为 function(Es $es)
     * @return  null;
     */
    public function must_not($func)
    {
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
    }

    /*
     * exists() 对应 ES 的 exists 和 missing
     *
     * @param   $field, 字段名
     * @param   $exists,  true or false
     */
    public function exists($field, $exists = true)
    {
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
    }

    /*
     * sort_score() 启用分数排序
     * 相当于 sort("_score", "desc")
     */
    public function sort_score()
    {
    }


    /*
     * to_query() 返回query数据，用于debug
     */
    public function to_query()
    {
    }

}
