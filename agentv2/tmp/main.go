package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	zmq "github.com/yangzhao28/zmq3"

	"github.com/yangzhao28/rtb/agent/servicebase"
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

	Conf      *BootstrapConfigure
	ZkService *servicebase.ServiceBase
	// := servicebase.NewServiceBase([]string{"0.0.0.0:2181"}, "rtb-test", "mtl")
	ToRouter         *servicebase.ZmqRouter
	ToAgentConfigure *servicebase.ZmqRouter
	ToPostAuction    *servicebase.ZmqRouter

	Quit chan bool
)

func init() {
	flag.StringVar(&BootstrapPath, "B", "conf/sample.bootstrap.json", "path of rtbkit bootstap file")
	flag.StringVar(&AgentName, "N", "faker", "name prefix of agent")
	flag.StringVar(&JsConfigure, "f", "", "rtbkit agent json configure file to start with")

	Quit = make(chan bool)
}

func OnAuction(name string, msg []string) {
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

func OnEvent(name string, msg []string) {
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

func Loop() {
	for {
		select {
		case <-Quit:
			return
		}
	}
}

func UpdateAgentConfigure() {
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

func main() {
	flag.Parse()
	var err error
	if Conf, err = ParseBootstrap(BootstrapPath); err != nil {
		log.Fatalln(err.Error())
	}
	AgentName += strconv.Itoa(os.Getpid())
	AgentName = "test.json"
	ToRouter = servicebase.NewZmqRouter(AgentName)
	ToAgentConfigure = servicebase.NewZmqRouter(AgentName)
	ToPostAuction = servicebase.NewZmqRouter(AgentName)

	ZkService = servicebase.NewServiceBase([]string{"0.0.0.0:2181"}, "rtb-test", "mtl")
	go ZkService.AutoRecovery()
	ZkService.Connect()

	// agent config
	ZkService.NewWatcher("rtbAgentConfiguration", "agents").OnNew(func(name string, uri string) {
		log.Println(">>", name, uri)
		if err := ToAgentConfigure.NewConnection(name, uri, zmq.DEALER); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("connected to", name)
			/*
				enabled := GetEnableConfigures()
				if enabled != nil {
					for _, config := range *enabled {
						ToAgentConfigure.SendMessage(name, "CONFIG", config.Name, config.Raw)
					}
				}
			*/
		}
	}).Watch()

	// post auction
	ZkService.NewWatcher("rtbPostAuctionService", "logger").OnNew(func(name string, uri string) {
		log.Println(">>", name, uri)
		if err := ToPostAuction.NewConnection(name, uri, zmq.SUB); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("connected to", name)
		}
	}).Watch()

	// router
	ZkService.NewWatcher("rtbRequestRouter", "agents").OnNew(func(name string, uri string) {
		log.Println(">>", name, uri)
		if err := ToRouter.NewConnection(name, uri, zmq.DEALER); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("connected to", name)
		}
	}).Watch()

	ToRouter.OnRecvMessage = func(name string, sock *zmq.Socket) {
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
			OnAuction(name, msg)
		case "NEEDCONFIG":
			enabled := GetEnableConfigures()
			if enabled != nil {
				for _, config := range *enabled {
					ToAgentConfigure.SendMessageToAll(name, "CONFIG", config.Name, config.Raw)
				}
			}
		case "GOTCONFIG":
			fmt.Println("router confirmed configure request")
		case "PING0":
			if len(msg) < 2 {
				return
			}
			received := msg[1]
			now := fmt.Sprintf("%.5f", float64(time.Now().UnixNano())/1e9)
			sock.SendMessage("PONG0", received, now, msg)
		case "PING1":
			if len(msg) < 2 {
				return
			}
			// payload := msg[2:]
		case "BYEBYE":
			fmt.Println("rejected by router, need reconnect")
		default:
			fmt.Println(msg)
		}
	}
	go ToRouter.MessageLoop()
	ToPostAuction.OnRecvMessage = func(name string, sock *zmq.Socket) {
		msg, err := sock.RecvMessage(0)
		if err != nil {
			return
		}
		if len(msg) == 0 {
			return
		}
		OnEvent(name, msg)
	}
	go ToPostAuction.MessageLoop()

	go UpdateAgentConfigure()
	Loop()
}
