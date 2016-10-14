<?php
/*
 *  .env        环境配置
 *  .consts     常量配置   
 *  config/     配置(该配置不应该纳入git管理)
 *      database.php    数据库
 *      cache.php       缓存
 *      queue.php       队列
 *      log.php         日志
 *      api.php         第三方接口
 *  lib/
 *      database/       数据库类
 *      cache/          缓存类
 *      queue/          队列类
 *      log/            日志类
 *      api/            第三方接口类
 *
 *  app.php     核心调度文件
 *
 *  modules/    web业务模块类
 *
 *  console/    控制台业务模块类
 *
 *  middleWare/    中间件类
 *
 */

/*
 * 访问配置
 */
Env("key");
conf()->get("option.key");
conf()->set("option.key", "value"); //php config "option.key" "value"
conf("127.0.0.1")->set("option.key", "value");


/*
 * io访问
 */
Db("db0")->table("user")->select("...");
Cache("cache0")->get($key)->remember($key, "...");
Queue("redis0")->push($obj)->pop();
Log("log_queue0")->log("message here");
Api("payment_api")->request($params);

/*
 * web路由
 */
//中间件全局生效
app()->before(['User@auth']);
//中间件在该请求路径上生效
app()->get('/user/info/{uid}/{format=html}', 'App\Modules\User@info')
    ->before(['User@auth'])
    ->after(['Static@report'])
    ->viewBy('Debug({format})@render')  //使用自定义函数 和 请求参数进行渲染
    ->viewBy('Format({format})@render') //调用多次viewBy() 或者 view() 时，最后一个会生效，它会覆盖前面的
    ->view('/user/info.tpl', 'SMARTY')
    ->view('/user/info.php', 'PHP');    //原生php

app()->post('/font/download', 'App\Modules\Resource@download')->withParam("resourceType", "font");
app()->web();

/*
 * console使用
 */
app()->cmd('user info {uid:用户ID}', 'App\Console\User@info')
    ->withCookie('utoken', '****')
    ->before(['User@auth'])
    ->after(['Static@report'])
    ->viewBy('print_r');    //使用php原生函数进行渲染
app()->exec();


/*
 * 监控
 */

//开发环境监控DB的所有query查询
app()->monitor('App\DB@query()', 'App\Debug@print');
//生产环境监控DB的 支付表更新，并调用相应模块进行上报
app()->monitor('App\DB@table("payment") -> App\DB@update()', 'App\Modules\Monitor\Static@report');
//开发环境监控所有请求的 request 和 response
app()->monitor('App\Modules\*@*', 'Debug@print');
//生产环境监控用户登录模块的 request 和 response, 并记录到日志中
app()->monitor('App\Modules\User@Login', 'App\Modules\Monitor\Static@Log');
//开发/生产环境监控所有的输入输出，并记录成测试集
//所有输入输出包括： DB(), Cache(), Queue(), Log(), Api()
app()->monitor('*@*', 'App\Tester@record()'); 








/*
 * C style
 */
    include "include.php";

    core_errtype(CORE_ERRTYPE_ERRNO_ERRMSG);
    core_default_init();

    if (core_errno() == CORE_ERR_NONE) {

        core_errtype(CORE_ERRTYPE_ERRNO_ERRMSG);
        //core_errtype(CORE_ERRTYPE_EXCEPTION);
        include "tmp_cfg.php";
        core_run("job.example");
        if (core_errno() == CORE_ERR_NONE) {

            Core_flush(CORE_STDOUT);

        } else {

            Core_flush(CORE_STDERR);
            Core_flush(CORE_STDOUT, CORE_IO_FLUSH_NULL);
        }

    } else {
        core_flush(CORE_STDERR);
    }
