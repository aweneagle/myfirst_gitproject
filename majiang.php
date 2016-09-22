<?php

$b = new Bird;

//大对碰
$cards = ['1w','1w'];
$b->ck($cards);
$cards = ['df','df'];
$b->ck($cards);
$cards = ['df','df','hz','hz','hz'];
$b->ck($cards);

//普通胡
$cards = ['df','df','7w','8w','9w', '1t','2t','3t'];
$b->ck($cards);

//顺子
$cards = ['df','df','2w','3w','3w','4w','4w','5w'];
$b->ck($cards);

//双豪华七姐
$cards = ['df','df','hz','hz','hz','hz','4w','4w','5w','5w','3w','3w','df','df'];
$b->ck($cards);

class Bird
{

	private $codes = [
		//万子
		"1w" => 1,
		"2w" => 2,
		"3w" => 3,
		"4w" => 4,
		"5w" => 5,
		"6w" => 6,
		"7w" => 7,
		"8w" => 8,
		"9w" => 9,

		//条子
		"1t" => 11,
		"2t" => 12,
		"3t" => 13,
		"4t" => 14,
		"5t" => 15,
		"6t" => 16,
		"7t" => 17,
		"8t" => 18,
		"9t" => 19,

		//筒子
		"1o" => 21,
		"2o" => 22,
		"3o" => 23,
		"4o" => 24,
		"5o" => 25,
		"6o" => 26,
		"7o" => 27,
		"8o" => 28,
		"9o" => 29,

		//东风
		"df" => 51,
		//西风
		"xf" => 53,
		//南风
		"nf" => 55,
		//北风
		"bf" => 57,

		//发财
		"fc" => 59,
		//红中
		"hz" => 61,
		//白板
		"bb" => 63,
		];

	/**
	 * 从可用模式中找出所有“胡牌”的模式
	 */
	public function findSuccPatterns($patterns)
	{
		$SUCC_PATTERNS = [
			["1*2" => 1],	//
			["1*2", "123"],
			["1*2", "1*3"],	//大对碰
			//......   待补
			];
	}

	/**
	 * 从能胡牌的模式中找出“番数”最大的模式
	 */
	public function findBestPatterns($patterns)
	{
	}

	/**
	 * 搜索出所有可能的匹配模式
	 * @return   false, 模式匹配出错，直接返回; true, 模式匹配成功
	 */
	public function pattern($codes, array &$patterns = array(), $curr_chain=array())
	{
		$possible = [
			"1*4" => 4,     //4张牌, 豪华
			"1*3" => 3,     //3张牌, 三张
			"1*2" => 2,     //2张牌, 对子
			"123" => 3,     //3张牌, 顺子
			];

		$found = false;
		foreach ($possible as $type => $num) {
			if ($group = $this->findPattern($codes, $type)) {

				$left_chain = $curr_chain;
				$left_chain[] = ["cards" => $group, "type" => $type];
				//成功匹配到一组模式后，剩下的牌
				$codes_left = $codes;
				foreach ($group as $i => $card) {
					unset($codes_left[$i]);
				}

				$group = array_values($group);
				//全部成功匹配?
				if (empty($codes_left)) {
					$patterns[] = $left_chain;
					//回溯至上一节点
					//(因为后面没有牌的情况下
					//只可能有一种模式, 4张为“豪华”，3张要么“顺子”,要么“三张”，2张为“对子”，
					//所以这里可以直接return了)
					return true;

				} 
				//成功匹配到
				if ($this->pattern(array_values($codes_left), $patterns, $left_chain)) {
					$found = true;
				} 
			}
		}

		return $found;

	}


	/**
	 * getCardsCode 将牌名映射成数字编码
	 */
	public function getCardsCodes($cards)
	{
		$codes = $this->codes;
		$return = array();
		foreach ($cards as $c) {
			$return[] = $codes[$c];
		}
		sort($return);
		return $return;
	}
	/**
	 * findPattern 根据模式类型$type, 看看是否能够找出对应的一组牌，并返回这组牌
	 *
	 * @return  false 查找失败;  array() 一组牌体
	 */
	private function findPattern($codes, $type="1*4")
	{
		switch ($type) {
		case "1*4" : 
			if (count($codes) >= 4) {
				if ($codes[0] == $codes[1] && $codes[0] == $codes[2] && $codes[0] == $codes[3]) {
					return array_slice($codes, 0, 4);
				}
			}
			break;

		case "1*3":
			if (count($codes) >= 3) {
				if ($codes[0] == $codes[1] && $codes[0] == $codes[2]) {
					return array_slice($codes, 0, 3);
				}
			}
			break;

		case "1*2":
			if (count($codes) >= 2) {
				if ($codes[0] == $codes[1]) {
					return array_slice($codes, 0, 2);
				}
			}
			break;

		case "123":
			if (count($codes) >= 3) {
				$card = $codes[0]; 	//第一张牌
				$group = [];
				$group[0] = $card; 		//顺子(只存牌的位置)
				for ($i = 1; $i < count($codes) && count($group) < 3; $i ++) {
					if ($card == $codes[$i]) {
						continue;
					} elseif ($card == ($codes[$i] - 1)) {
						$card = $codes[$i];
						$group[$i] = $card;
						continue;
					} else {
						return false;
					}
				}
				if (count($group) == 3) {
					return $group;
				}
			}
			break;
		}

		return false;
	}


	/**
	 * 测试函数
	 */
	public function ck($cards)
	{
		$patterns = array();
		$this->pattern($this->getCardsCodes($cards), $patterns);
		echo "\n";
		echo "!!!  " . implode(" ",$cards) . " !!!\n";
		if ($patterns) {
			global $argv;
			echo "!!!  此牌能胡的方式有: !!!\n";
			if (!isset($argv[1])) {
				foreach ($patterns as $plist) {
					$list = [];
					foreach ($plist as $p) {
						$list[] = $p['type'];
					}
					echo "|||   " . implode(",", $list) . "    |||\n";
				}
			} else {
				print_r($patterns);
			}
		} else {
			echo "!!!  该牌不能胡 !!!\n";
		}
	}

}

