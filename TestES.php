<?php
class ESTest
{

    public function testShould()
    {
        $this->testBool("should");
        $this->testFilterAndQueryBool("should");
    }

    public function testDefault()
    {
        /* where 默认是在 must 子句中 */
        $es = new ES;
        $es->where("price", ">=", 10);
        $query1 = $es->to_query();

        $es = new ES;
        $es->must(function($es) {
            $es->where("price", ">=", 10);
        });
        $query2 = $es->to_query();
        $this->assertEquals($query1, $query2);

        /* match 默认是在should子句中 */
        $es = new ES;
        $es->match("name", "es");
        $query1 = $es->to_query();

        $es = new ES;
        $es->should(function($es) {
            $es->match("name", "es");
        });
        $query2 = $es->to_query();
        $this->assertEquals($query1, $query2);

        /* where 和 match 并存*/
        $es = new ES;
        $es->where("price", ">=", 10)
           ->match("name", "es");
        $query1 = $es->to_query();

        $es = new ES;
        $es->must(function($es) {
            $es->where("price", ">=", 10);
        });
        $es->should(function($es) {
            $es->match("name", "es");
        });
        $query2 = $es->to_query();
        $this->assertEquals($query1, $query2);

    }

    public function testMust()
    {
        $this->testBool("must");
        $this->testFilterAndQueryBool("must");

    }

    public function testMustNot()
    {
        $this->testBool("must_not");
        $this->testFilterAndQueryBool("must_not");
    }

    public function testSort()
    {
        $es = new ES;
        $es->must(function($es) {
            $es->where("price", ">=", 10);
        })->sort("date")
          ->sort_score()
          ->sort("age", "asc");
        $query = $es->to_query();
        $this->assertEquals($query['sort'],
            [
                ["date" => ["order" => "desc"]],
                ["_score" => ["order" => "desc"]],
                ["age" => ["order" => "asc"]],
            ]);
    }

    public function testQuery()
    {
        $this->testQueryBool("should", "must", "should");
        $this->testQueryBool("should", "must", "must");
        $this->testQueryBool("should", "must", "must_not");

        $this->testQueryBool("must", "should", "should");
        $this->testQueryBool("must", "should", "must");
        $this->testQueryBool("must", "should", "must_not");

        $this->testQueryBool("must_not", "should", "should");
        $this->testQueryBool("must_not", "should", "must");
        $this->testQueryBool("must_not", "should", "must_not");

        $this->testQueryBool("must_not", "must", "should");
        $this->testQueryBool("must_not", "must", "must");
        $this->testQueryBool("must_not", "must", "must_not");
    }

    public function testFilter()
    {
        $this->testFilterBool("should", "must", "should");
        $this->testFilterBool("should", "must", "must");
        $this->testFilterBool("should", "must", "must_not");

        $this->testFilterBool("must", "should", "should");
        $this->testFilterBool("must", "should", "must");
        $this->testFilterBool("must", "should", "must_not");

        $this->testFilterBool("must_not", "should", "should");
        $this->testFilterBool("must_not", "should", "must");
        $this->testFilterBool("must_not", "should", "must_not");

        $this->testFilterBool("must_not", "must", "should");
        $this->testFilterBool("must_not", "must", "must");
        $this->testFilterBool("must_not", "must", "must_not");
    }

    private function testBool($bool)
    {
        $es = new ES;
        $es->$bool(function($es) {
            $es->match("title", "es");
            $es->match("content", "es");
        });
        $query = $es->to_query();
        $this->assertEquals($query, [
            "query" => [
                "bool" => [
                    "$bool" => [
                        ["match" => ["title" => "es"]],
                        ["match" => ["content" => "es"]]
                    ]
                ]
            ],
        ]);
    }

    private function testQueryBool($bool1, $bool2, $bool3)
    {
        $es = new ES;
        $es->$bool1(function($es) {
            $es->match("title", "es", 2);
            $es->match("content", "es", 1);
        })->$bool2(function($es) {
            $es->match("title", "es", 2);
            $es->match("content", "es", 1);
            $es->$bool3(function($es) {
                $es->match("title", "es", 2);
                $es->match("content", "es", 1);
            });
        });
        $this->assertEquals($es->query,
            [
                "query" => [
                    "bool" => [
                        "$bool1" => [
                            ["match" => ["title" => ["query" => "es", "boost" => 2]]],
                            ["match" => ["content" => ["query" => "es", "boost" => 1]]],
                        ],
                        "$bool2" => [
                            ["match" => ["title" => ["query" => "es", "boost" => 2]]],
                            ["match" => ["content" => ["query" => "es", "boost" => 1]]],
                            ["bool" => [
                                "$bool3" => [
                                    ["match" => ["title" => ["query" => "es", "boost" => 2]]],
                                    ["match" => ["content" => ["query" => "es", "boost" => 1]]],
                                ],
                            ]],
                        ],
                    ]
                ]
            ]
        );
    }

    private function testFilterBool($bool1, $bool2, $bool3)
    {
        $es = new ES;
        $es->$bool1(function($es) {
            $es->where("price", ">=", 2);
            $es->where("price", "<=", 10);
            $es->where("age", ">", 5);
            $es->where("age", "<", 9);
            $es->$bool3(function($es) {
                $es->where("author", "in", ["awen", "king"]);
                $es->where("publisher", "in", ["bbc", "acc"]);
            });
        });
        $es->$bool2(function($es) {
            $es->exists("weight");
            $es->exists("name");
            $es->exists("address", false);
        });
        $this->assertEquals($es->to_query(),
            [
                "filter" => [
                    "bool" => [
                        "$bool1" => [
                            ["range" => ["price" => ["gte" => 2]]],
                            ["range" => ["price" => ["lte" => 10]]],
                            ["range" => ["age" => ["gt" => 5]]],
                            ["range" => ["age" => ["lt" => 9]]],
                            ["bool" => [
                                "$bool3" => [
                                    ["author" => ["terms" => ["awen", "king"]]],
                                    ["publisher" => ["terms" => ["bbc", "acc"]]],
                                ],
                            ]],
                        ],
                        "$bool2" => [
                            ["exists" => ["field" => "weight"]],
                            ["exists" => ["field" => "name"]],
                            ["missing" => ["field" => "address"]],
                        ],
                    ]
                ]
            ]
        );
    }

    private function testFilterAndQueryBool($bool)
    {
        $es = new ES;
        $es->$bool(function($es) {
            $es->match("title", "es");
            $es->match("content", "es");
        })->$bool(function($es) {
            $es->where("price", ">=", 10);
            $es->where("price", "<=", 100);
            $es->where("autho", "in", ["king", "awen"]);
        })->sort("date")
          ->sort("_score", "asc");

        $query = $es->to_query();
        $this->assertEquals($query, [
            "query" => [
                "filtered" => [
                    "query" => [
                        "bool" => [
                            "$bool" => [
                                ["match" => ["title" => "es"]],
                                ["match" => ["content" => "es"]],
                            ],
                        ],
                    ],
                    "filter" => [
                        "bool" => [
                            "$bool" => [
                                ["range" => ["price" => ["gte" => 10, "lte" => 100]]],
                                ["autho" => ["terms" => ["king", "awen"]]],
                            ]
                        ]
                    ],
                ]
            ],
            "sort" => [
                "date" => ["order" => "desc"],
                "_score" => ["order" => "asc"],
            ],
        ]);

    }


}
