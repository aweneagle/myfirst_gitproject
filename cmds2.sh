./awen-els.sh DELETE "/awen2/";
#./awen-els.sh PUT "/awen2" '{ "settings": { "number_of_shards": 1 }}';
./awen-els.sh PUT "/awen2";
./awen-els.sh PUT "/awen2/test/1" '{"text":"therewe is my pretty good dog"}';
./awen-els.sh PUT "/awen2/test/2" '{"text":"a"}';
./awen-els.sh PUT "/awen2/test/3" '{"text":"good dog a here"}';
./awen-els.sh PUT "/awen2/test/4" '{"text":"come come a"}';
./awen-els.sh PUT "/awen2/test/5" '{"text":"a dog and a cat"}';
./awen-els.sh PUT "/awen2/test/6" '{"text":"mini dog we have a"}';
./awen-els.sh PUT "/awen2/test/7" '{"text":"a month"}';
./awen-els.sh PUT "/awen2/test/8" '{"text":"a year"}';
./awen-els.sh PUT "/awen2/test/9" '{"text":"come for a driver"}';
./awen-els.sh PUT "/awen2/test/10" '{"text":"qiku is a company"}';