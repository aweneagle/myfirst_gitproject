<?php
$begin = time();
$url[] = 'https://play.google.com/store/apps/details?id=com.progimax.moto.free';
$url[] = 'https://play.google.com/store/apps/details?id=com.moto.acceleration';
$url[] = 'https://play.google.com/store/apps/details?id=me.pou.app';

// Setando opção padrão para todas url e adicionando a fila para processamento
$ch = curl_init();
foreach($url as $key => $value) {
    curl_setopt($ch, CURLOPT_URL, $value);
    curl_setopt($ch, CURLOPT_HEADER, true);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $content = curl_exec($ch);
    $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $header_size = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
    $header = substr($content, 0, $header_size);

    echo "$value\n";
    echo "$code\n";
    echo "$header\n\n\n";
}

curl_close($ch);
