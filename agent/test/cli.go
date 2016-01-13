package main

import (
	"fmt"

	zmq "github.com/pebbe/zmq3"
)

func main() {
	//  Socket to talk to server
	fmt.Println("Connecting to hello world server...")
	requester, _ := zmq.NewSocket(zmq.REQ)
	defer requester.Close()
	requester.Connect("tcp://ubuntu:15000")
	// send hello
	requester.SetIdentity("agent")
	requester.SendMessage("CONFIG", "agent", "{ \"a\" :  1}")
}
