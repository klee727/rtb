package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

var (
	JsConfig string = `{
		"account":["faker","pace"],
		"bidProbability":1,
		"bidControl":{"fixedBidCpmInMicros":0,"type":"RELAY"},
		"creatives":[
		{
			"exchangeFilter":{"include":["youku"]},
			"format":"640x480",
			"id":0,
			"name":"0",
			"providerConfig":
			{
				"youku":
				{
					"crid":"crid1",
					"nurl":"http://relay.bigtree.mobi:13898/win_notice?price=${AUCTION_PRICE}&params=${AUCTION_ID}",
					"adm":"http://bttest.qiniudn.com/1/1107185038-552008732/552008732_1431337249.jpg",
					"ldp":"https://itunes.apple.com/cn/app/chao-shen-zhan-ji-guo-min/id967694144?l=zh&ls=1&mt=8",
					"cm":["http://nbb-toutiao.bigtree.mobi:13898/click?params=${AUCTION_ID}"],
					"pm":["http://nbb-toutiao.bigtree.mobi:13898/impression?params=${AUCTION_ID}"],
					"image_width":640,"image_height":480
				}
			}
		}
		],
		"errorFormat":"lightweight",
		"external":false,
		"externalId":0,
		"lossFormat":"lightweight",
		"maxInFlight":20000,
		"minTimeAvailableMs":5.0,
		"test":false,
		"winFormat":"full"}`
)

var (
	BootstrapPath string
	AgentName     string
	JsConfigure   string
)

func init() {
	flag.StringVar(&AgentName, "N", "faker", "name prefix of agent")
	flag.StringVar(&BootstrapPath, "B", "conf/sample.bootstrap.json", "path of rtbkit bootstap file")
	flag.StringVar(&JsConfigure, "f", "", "rtbkit agent json configure file to start with")

	Quit = make(chan bool)
}

type WatcherEvent struct {
	Name string
	Uri  string
}

type BiddingAgent struct {
	Name               string
	AgentConfigureConn *ZmqRouter
	RouterConn         *ZmqRouter
	PostAuctionConn    *ZmqRouter
	Bootstrap          *BootstrapConfigure
	Service            *ServiceBase
	Quit               chan bool

	AgentConfigureNotifier chan *WatcherEvent
	RouterNotifer          chan *WatcherEvent
	PostAuctionNotifier    chan *WatcherEvent

	InitOnce sync.Once
}

func NewBiddingAgent(name string, service *ServiceBase, bootstrap *BootstrapConfigure) *BiddingAgent {
	return &BiddingAgent{
		Name:      name,
		Bootstrap: bootstrap,
		Service:   service,
		Quit:      make(chan bool),
		AgentConfigureNotifier: make(chan *WatcherEvent),
		RouterNotifer:          make(chan *WatcherEvent),
		PostAuctionNotifier:    make(chan *WatcherEvent),
	}
}

func (this *BiddingAgent) Init(bootstrapPath string) error {
	this.InitOnce.Do(func() {
		// this.Bootstrap, err = ParseBootstrap(bootstrapPath)
		// if err != nil {
		// 	return err
		// }
		// this.Service = NewServiceBase([]string{this.Bootstrap.ZookeeprUri}, this.Bootstrap.Installation, this.Bootstrap.Location)
		this.AgentConfigureConn = NewZmqRouter(this.Name)
		this.PostAuctionConn = NewZmqRouter(this.Name)
		this.RouterConn = NewZmqRouter(this.Name)
		// [SetHanlder]
		this.RouterConn.OnRecvMessage = handleRouterMessage
		ToPostAuction.OnRecvMessage = handlePostAuctionMessage
		// [AsyncZmqConnection]
		go func() {
			for {
				var err error
				select {
				case event := <-this.AgentConfigureNotifier:
					err = this.AgentConfigureConn.NewConnection(event.Name, event.Uri, zmq.DEALER)
				case event := <-this.PostAuctionNotifier:
					// [NOTICE] post auction is 'sub'
					err = this.PostAuctionNotifier.NewConnection(event.Name, event.Uri, zmq.SUB)
				case event := <-this.RouterNotifer:
					err = this.RouterNotifer.NewConnection(event.Name, event.Uri, zmq.DEALER)
				}
				if err != nil {
					log.Printf("fail to handle zmq event(%v, %v): %s", event.Name, event.Uri, err.Error())
				} else {
					log.Printf("connect to %s@%s", event.Name, event.Uri)
				}
			}
		}()
		this.Service.NewWatcher("rtbAgentConfiguration", "agents").OnNew(func(name string, uri string) {
			AgentConfigureNotifier <- &WatcherEvent{name, uri}
		}).Watch()
		this.Service.NewWatcher("rtbPostAuctionService", "logger").OnNew(func(name string, uri string) {
			PostAuctionNotifier <- &WatcherEvent{name, uri}
		}).Watch()
		this.Service.NewWatcher("rtbRequestRouter", "agents").OnNew(func(name string, uri string) {
			RouterNotifer <- &WatcherEvent{name, uri}
		}).Watch()
	})
}

func (this *BiddingAgent) handleAuction(name string, msg []string) {
	id := msg[2]
	// source := msg[3]

	bid := map[string]interface{}{
		"creative":   0,
		"price":      "3010USD/1M",
		"priority":   299,
		"spotIndex":  0,
		"account":    "",
		"biddingKey": "",
	}
	bids := map[string]interface{}{
		"bids": []map[string]interface{}{
			bid,
		},
	}
	bidsJson, err := json.Marshal(bids)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Println("==> from", string(bidsJson))
	ToRouter.SendMessage(name, "BID", id, bidsJson, "null", "{}")
}

func (this *BiddingAgent) handleRouterMessage(name string, sock *zmq.Socket) {
	msg, err := sock.RecvMessage(0)
	if err != nil {
		log.Println("recv error:", err.Error())
		return
	}
	log.Println("Msg:", msg)
	if len(msg) == 0 {
		return
	}
	switch msg[0] {
	case "AUCTION":
		handleAuction(name, msg)
	case "NEEDCONFIG":
		// [TODO] send available configures
	case "PING0":
		if len(msg) < 2 {
			return
		}
		received := msg[1]
		now := fmt.Sprintf("%.5f", float64(time.Now().UnixNano())/1e9)
		sock.SendMessage("PONG0", received, now, msg)
	case "BYEBYE":
		fmt.Println("rejected by router, need reconnect")
	default:
		fmt.Println(msg)
	}
}

func (this *BiddingAgent) handleEvent(name string, msg []string) {
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

func (this *BiddingAgent) handlePostAuctionMessage(name string, sock *zmq.Socket) {
	msg, err := sock.RecvMessage(0)
	if err != nil {
		return
	}
	if len(msg) == 0 {
		return
	}
	handleEvent(name, msg)
}

func (this *BiddingAgent) Servce() {
	go RouterConn.MessageLoop()
	go PostAuctionConn.MessageLoop()
}

func (this *BiddingAgent) Stop() {
	this.RouterConn.Close()
	this.PostAuctionConn.Close()
	this.Quit <- true
}

/*
func (this *BiddingAgent) UpdateAgentConfigure(configure []byte) {
	for {
		select {
		case config := <-NewConfigure:
			if config != nil {
				log.Println("new configure found: ", config.Name)
				ToAgentConfigure.SendMessageToAll("CONFIG", config.Name, config.Raw)
				ToRouter.SendMessageToAll(config.Name, "CONFIG", config.Name)
			}
		case config := <-RemoveConfigure:
			if config != nil {
				log.Println("configure removed: ", config.Name)
				ToAgentConfigure.SendMessageToAll("CONFIG", config.Name, "")
			}
		}
	}
}
*/

func (this *BiddingAgent) UpdateAgentConfigure(configure []byte) {
	this.AgentConfigureConn.SendMessage("CONFIG", this.Name, configure)
	this.RouterConn.SendMessage("CONFIG", this.Name)
}
