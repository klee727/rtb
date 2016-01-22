package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
)

type WatcherEvent struct {
	Name string
	Uri  string
}

type EventListener struct {
	Name      string
	Bootstrap *BootstrapConfigure
	Service   *ServiceBase
	Quit      chan bool

	PostAuctionConn     *ZmqRouter
	PostAuctionNotifier chan *WatcherEvent

	InitOnce sync.Once
}

func NewEventListener(name string, service *ServiceBase, bootstrap *BootstrapConfigure) *EventListener {
	return &EventListener{
		Name:                name,
		Bootstrap:           bootstrap,
		Service:             service,
		Quit:                make(chan bool),
		PostAuctionNotifier: make(chan *WatcherEvent),
	}
}

func (this *EventListener) Init(bootstrapPath string) error {
	this.InitOnce.Do(func() {
		this.PostAuctionConn = NewZmqRouter(this.Name)
		ToPostAuction.OnRecvMessage = handlePostAuctionMessage
		// [AsyncZmqConnection]
		go func() {
			for event := range this.PostAuctionNotifier {
				if err := this.PostAuctionNotifier.NewConnection(event.Name, event.Uri, zmq.SUB); err != nil {
					log.Printf("fail to handle zmq event(%v, %v): %s", event.Name, event.Uri, err.Error())
				} else {
					log.Printf("connect to %s@%s", event.Name, event.Uri)
				}
			}
		}()
		this.Service.NewWatcher("rtbPostAuctionService", "logger").OnNew(func(name string, uri string) {
			PostAuctionNotifier <- &WatcherEvent{name, uri}
		}).Watch()
	})
}

func (this *EventListener) Servce() {
	PostAuctionConn.MessageLoop()
}

func (this *EventListener) Stop() {
	this.PostAuctionConn.Close()
	this.Quit <- true
}

func (this *EventListener) handleEvent(name string, msg []string) {
	switch msg[0] {
	case "BMATCHEDWIN":
		fmt.Println("event: win")
		params := &UrlParam{}
		if err := proto.Unmarshal([]byte(msg[1]), params); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("params:", params)
		}
	default:
		event := msg[1]
		switch event {
		case "IMPRESSION":
		case "CLICK":
		case "CONVERSION":
		default:
			return
		}
		fmt.Println("event: ", event)
		params := &UrlParam{}
		if err := proto.Unmarshal([]byte(msg[2]), params); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("params:", params)
		}
	}
}

func (this *EventListener) handlePostAuctionMessage(name string, sock *zmq.Socket) {
	msg, err := sock.RecvMessage(0)
	if err != nil {
		return
	}
	if len(msg) == 0 {
		return
	}
	handleEvent(name, msg)
}
