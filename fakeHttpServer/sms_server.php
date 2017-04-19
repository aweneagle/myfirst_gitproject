<?php
require __DIR__ . "/FakeHttpServer.php";
use Tools\FakeHttpServer;
$server = new FakeHttpServer("localhost", 8800);
$server->post("/sms",function($req, $resp){
    $resp->body = ['errcode' => 0];
});
$server->start(FakeHttpServer::DEAMO);
//$server->stop();
