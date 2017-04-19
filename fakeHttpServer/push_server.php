<?php
require __DIR__ . "/FakeHttpServer.php";
use Tools\FakeHttpServer;
$server = new FakeHttpServer("localhost", 9909);
$server->post("/message/send",function($req, $resp){
    $resp->body = ['info' => "OK", 'result' => "00"];
});
$server->get("/check/online",function($req, $resp){
    $resp->body = ['errcode' => 0, 'isonline' => 1];
});
$server->start(FakeHttpServer::DEAMO);
//$server->stop();
