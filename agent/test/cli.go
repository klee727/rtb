package main

import (
	"fmt"
	"os"

	zmq "github.com/pebbe/zmq3"
)

func main() {
	//  Socket to talk to server
	addr := os.Args[1]
	fmt.Println("Connecting to server", addr)
	requester, _ := zmq.NewSocket(zmq.SUB)
	requester.SetSubscribe("")
	defer requester.Close()
	requester.Connect(addr)
	// send hello
	requester.SetIdentity("listener")
	for {
		m, err := requester.RecvMessage(0)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(m)
	}
}
