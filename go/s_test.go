package main
import "./chat"
import "time"

func main() {
	cli := chat.NewClient("127.0.0.1", 9999)
	err := cli.UserLogin(1234)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go func() {
		for {
			cli.SendToUser("abcdef", 1222)
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		var buff [1024]byte
		for {
			cli.Read(buff)
			fmt.Println(string(buff))
			time.Sleep(time.Second * 1)
		}
	}

	fmt.Scanln("press any key to stop :")

}

