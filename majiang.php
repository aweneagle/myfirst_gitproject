<?php
$codes = [
    "1w" => 1,
    "2w" => 2,
    "3w" => 3,
    "4w" => 4,
    "5w" => 5,
    "6w" => 6,
    "7w" => 7,
    "8w" => 8,
    "9w" => 9,

    "1t" => 11,
    "2t" => 12,
    "3t" => 13,
    "4t" => 14,
    "5t" => 15,
    "6t" => 16,
    "7t" => 17,
    "8t" => 18,
    "9t" => 19,

    "1o" => 21,
    "2o" => 22,
    "3o" => 23,
    "4o" => 24,
    "5o" => 25,
    "6o" => 26,
    "7o" => 27,
    "8o" => 28,
    "9o" => 29,

    "df" => 51,
    "xf" => 53,
    "nf" => 55,
    "bf" => 57,

    "fc" => 59,
    "hz" => 61,
    "bb" => 63,
];

function findPattern($codes, $type="1*4")
{
    switch ($type) {
    case "1*4" : 
        if (count($codes) >= 4) {
            return $codes[0] == $codes[1] && $codes[0] == $codes[2] && $codes[0] == $codes[3];
        }
        break;

    case "1*3":
        if (count($codes) >= 3) {
            return $codes[0] == $codes[1] && $codes[0] == $codes[2];
        }
        break;

    case "1*2":
        if (count($codes) >= 2) {
            return $codes[0] == $codes[1];
        }
        break;

    case "123":
        if (count($codes) >= 3) {
            return $codes[0] == ($codes[1] + 1) && $codes[0] == ($codes[2] + 2);
        }
        break;
    }

    return false;
}

/**
 * 搜索出所有可能的匹配模式
 * @return   false, 模式匹配出错，直接返回; true, 模式匹配成功
 */
function pattern($codes, array &$patterns = array(), $i = 0)
{
    if (!is_array($patterns)) {
        $patterns = array();
    }
    $patterns[$i] = array();
    $posibile = [
        "1*4" => 4,     //4张牌
        "1*3" => 3,     //3张牌
        "1*2" => 2,     //2张牌
        "123" => 3,     //3张牌
    ];
    $found = false;
    foreach ($posibile as $type => $num) {
        if (findPattern($codes, $type)) {
            $patterns[$i][] = $type;
            $found = true;
            $next_codes = array_slice($codes, $num);

            if (!pattern($next_codes, $patterns, $i+1)) {
                continue;
            }
        }
    }

}

/**
 * 从模式中找出所有“胡牌”的模式
 */
function findSuccPatterns($patterns)
{
}

/**
 * 从模式中找出“番数”最大的模式
 */
function findBestPatterns($patterns)
{
}
