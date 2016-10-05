<?php
namespace App;

class Majiang
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
     * $SUCC, 胡牌模式
     *
     * 牌体数目要求：
     * 	COUNT_NONE,  没有
     * 	COUNT_ONE,  有1副
     * 	COUNT_MUST,  有1 ~ n副
     * 	COUNT_MAY,  有0 ~ m副
     */
    const COUNT_NONE = 0;
    const COUNT_ONE = 1;
    const COUNT_MUST = 2;
    const COUNT_MAY = 3;

    private $SUCC = [
        /**
         * 'rule'   牌型规则
         * 'weight' 番数
         * 'name'   牌型名称
         * 'num'    (可选)牌数要求
         * 'must'   (可选)对应牌体的数量, 例如: ['1*4' => 1] 表示有且只有一个豪华
         */

        [
            'rule' => ["1*2" => self::COUNT_ONE, "1*3" => self::COUNT_MAY, "123" => self::COUNT_MUST, "1*4" => self::COUNT_NONE],
            'weight' => 1,
            'name' => "普通胡",
        ],
        [
            'rule' => ["1*2" => self::COUNT_ONE, "1*3" => self::COUNT_MAY, "123" => self::COUNT_NONE, "1*4" => self::COUNT_NONE],
            'weight' => 2,
            'name' => "大对碰",
        ],
        [
            'rule' => ["1*2" => self::COUNT_MUST, "1*3" => self::COUNT_NONE, "123" => self::COUNT_NONE, "1*4" => self::COUNT_NONE],
            'weight' => 2,
            'name' => "七姐",
            'num' => 14
        ],
        [
            'rule' => ["1*2" => self::COUNT_MUST, "1*3" => self::COUNT_NONE, "123" => self::COUNT_NONE, "1*4" => self::COUNT_MUST],
            'weight' => 3,
            'name' => "豪华七姐",
            'num' => 14,
            'must' => ["1*4" => 1]
        ],
        [
            'rule' => ["1*2" => self::COUNT_MUST, "1*3" => self::COUNT_NONE, "123" => self::COUNT_NONE, "1*4" => self::COUNT_MUST],
            'weight' => 4,
            'name' => "双豪华七姐",
            'num' => 14,
            'must' => ["1*4" => 2]
        ],
        [
            'rule' => ["1*2" => self::COUNT_MUST, "1*3" => self::COUNT_NONE, "123" => self::COUNT_NONE, "1*4" => self::COUNT_MUST],
            'weight' => 5,
            'name' => "三豪华七姐",
            'num' => 14,
            'must' => ["1*4" => 3]
        ],
    ];

    /**
     * 从可用模式中找出所有“胡牌”的模式
     */
    public function findSuccPatterns($codes, array &$patterns = array())
    {
        $all_patterns = array();
        $this->pattern($codes, $all_patterns);
        if ($all_patterns) {
            foreach ($all_patterns as $body) {
                $count = [];
                foreach ($body as $group) {
                    @$count[$group['type']] += 1;
                }

                foreach ($this->SUCC as $rule_info) {

                    // 该牌型 $body 是否符合 该规则 $list?
                    $match = true;
                    $must = $num = false;
                    if (isset($rule_info['must'])) {
                        $must = $rule_info['must'];
                    }
                    if (isset($rule_info['num'])) {
                        $num = $rule_info['num'];
                    }

                    if ($num && $num != count($codes)) {
                        continue;
                    }

                    foreach ($rule_info['rule'] as $type => $rule) {
                        switch ($rule) {
                        case self::COUNT_ONE:
                            if (!isset($count[$type]) || $count[$type] != 1) {
                                $match = false; 
                            }
                            break;

                        case self::COUNT_MUST:
                            if (!isset($count[$type])
                                || isset($must[$type]) && $must[$type] != $count[$type]
                            ) {
                                $match = false;
                            }
                            break;

                        case self::COUNT_NONE:
                            if (isset($count[$type])) {
                                $match = false;
                            }
                            break;

                        case self::COUNT_MAY:
                        default:
                            break;
                        }
                    }

                    if ($match) {
                        $patterns[] = [
                            'cards' => $body, 
                            'rule' => $rule_info,
                        ];
                    }
                }
            }
        }
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
    private function pattern($codes, array &$patterns = array(), $curr_chain=array())
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

    private function getCards($codes)
    {
        $map = array_flip($this->codes);
        $cards = [];
        foreach ($codes as $c) {
            $cards[] = $map[$c];
        }
        return $cards;
    }

    /**
     * getCardsCode 将牌名映射成数字编码
     */
    private function getCardsCodes($cards)
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


    private function e($msg)
    {
        echo "[ERROR] $msg" . $this->lk;
        return false;
    }

    /**
     * lk,  line break 
     */
    public $lk = "\n";
    /**
     * 测试函数
     */
    public function ck($cards)
    {
        $ck_nums = [];
        foreach ($cards as $c) {
            if (!isset($this->codes[$c])) {
                return $this->e("未知牌:" . $c);
            }
            @$ck_nums[$c] += 1;
        }
        if (count($cards) > 14 || count($cards) == 0) {
            return $this->e("牌数不正确:" . count($cards));
        }
        foreach ($ck_nums as $c => $n) {
            if ($n > 4) {
                return $this->e("$c牌数不正确:" . $n);
            }
        }


        $patterns = array();
        $this->findSuccPatterns($this->getCardsCodes($cards), $patterns);
        echo $this->lk;
        echo "  " . implode(" ",$cards) . " " . $this->lk;
        if ($patterns) {
            global $argv;
            echo "此牌能胡的方式有: " . $this->lk;
            foreach ($patterns as $plist) {
                $list = [];
                echo "[牌型]" . $plist['rule']['name'] . $this->lk;
                echo "[番数]" . $plist['rule']['weight'] . $this->lk;
                foreach ($plist['cards'] as $p) {
                    $list[] = implode(" ", $this->getCards($p['cards']));
                }
                echo "[牌体]" . implode(", ", $list) . " " . $this->lk;
                echo $this->lk;
            }
        } else {
            echo "!!!  该牌不能胡 !!!" . $this->lk;
        }
    }

}
