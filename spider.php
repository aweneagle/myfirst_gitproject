<?php
/*
 *	1. 入口:
 *		[url]
 *		header
 *		cookie
 *
 *		响应页面 body  => 
 *			a). 正则匹配 => 二级入口  function($page, Spider $s)  return [$urls]
 *			b). 抓取内容 => 获取内容  function($page, Spider $s)  return $contents
 *
 *	2. foreach (二级入口 as 入口)
 *			重复步骤 1.( 入口 )
 *
 *
 * 1.  
 */
class Spider
{

     private $history_urls = [];
 	 private $entry_urls = [];
 	 private $contents = [];

	 public $url = null;
	 public $host = null;

	 public $handle_fetch_entries;
	 public $handle_fetch_contents;
	 public $handle_store_contents;
 	 public $batch_num = 100;
 
	 public function run($url)
	 {
		 $this->entry_urls[] = $url;
		 do {
			 $url = array_shift($this->entry_urls);
			 if (isset($this->history_urls[$url])) {
				 continue;
			 }

			 $this->url = $url;
			 $this->history_urls[$url] = 1;
			 $this->view();

			 $new_entry_urls = $this->$handle_fetch_entries($this);
			 $new_contents = $this->$handle_fetch_contents($this);

			 $this->contents = array_merge($new_contents, $this->contents);
			 if (count($this->contents) >= $this->batch_num) {
				 $this->$handle_store_contents($this->contents);
				 $this->contents = [];
			 }
			 $this->entry_urls = array_unique(array_merge($new_entry_urls, $this->entry_urls));

		 } while (!empty($this->entry_urls));
	 }
}
 
