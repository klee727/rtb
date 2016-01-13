package main

import (
	"fmt"
	"time"

	zmq "github.com/pebbe/zmq3"
)

func main() {
	//  Socket to talk to clients
	responder, _ := zmq.NewSocket(zmq.REP)
	defer responder.Close()
	responder.Bind("tcp://*:5555")

	poller := zmq.NewPoller()
	poller.Add(responder, zmq.POLLIN)

	for {
		fmt.Println("wait")
		polled, err := poller.Poll(time.Second)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		for _, p := range polled {
			msg, _ := p.Socket.RecvMessage(0)
			fmt.Println("Received ", msg)
			p.Socket.Send("1", 0)
		}
	}
}
