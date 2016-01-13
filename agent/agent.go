package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	zmq "github.com/yangzhao28/zmq3"

	"github.com/yangzhao28/rtb/agent/servicebase"
)

var (
	jsConfig string = `{"account":["760547441","pace"],"bidProbability":1,"augmentations":{"random":null},"bidControl":{"fixedBidCpmInMicros":0,"type":"RELAY"},
	"creatives":[{"exchangeFilter":{"include":["youku"]},"format":"640x480","id":0,"name":"0","providerConfig":
	{"youku":{"crid":"crid1","nurl":"http://relay.bigtree.mobi:13898/win_notice?price=${AUCTION_PRICE}&params=${AUCTION_ID}",
	"adm":"http://bttest.qiniudn.com/1/1107185038-552008732/552008732_1431337249.jpg",
	"ldp":"https://itunes.apple.com/cn/app/chao-shen-zhan-ji-guo-min/id967694144?l=zh&ls=1&mt=8",
	"cm":["http://nbb-toutiao.bigtree.mobi:13898/click?params=${AUCTION_ID}"],
	"pm":["http://nbb-toutiao.bigtree.mobi:13898/impression?params=${AUCTION_ID}"],
	"image_width":640,"image_height":480}}}],
	"errorFormat":"lightweight","external":false,"externalId":0,"lossFormat":"lightweight","maxInFlight":20000,"minTimeAvailableMs":5.0,"test":false,"winFormat":"full"}`
)

func main() {
	rand.Seed(time.Now().Unix())
	agentName := "faker_" + strconv.Itoa(rand.Int())

	toRouter := servicebase.NewZmqRouter(agentName)
	toConfigure := servicebase.NewZmqRouter(agentName)

	sb := servicebase.NewServiceBase([]string{"0.0.0.0:2181"}, "rtb-test", "mtl")
	go sb.AutoRecovery()

	// agent config
	sb.NewWatcher("rtbAgentConfiguration", "agents").OnNew(func(name string, uri string) {
		log.Println(">>", name, uri)
		if err := toConfigure.NewConnection(name, uri, zmq.DEALER); err != nil {
			log.Println(err.Error())
		} else {
			toConfigure.SendMessage(name, "CONFIG", agentName, jsConfig)
		}
	}).Watch()

	// router
	sb.NewWatcher("rtbRequestRouter", "agents").OnNew(func(name string, uri string) {
		log.Println(">>", name, uri)
		if err := toRouter.NewConnection(name, uri, zmq.DEALER); err != nil {
			log.Println(err.Error())
		} else {
			toRouter.SendMessage(name, uri, "CONFIG", agentName)
		}
	}).Watch()

	toRouter.OnRecvMessage = func(sock *zmq.Socket) {
		msg, err := sock.RecvMessage(0)
		if err != nil {
			return
		}
		if len(msg) == 0 {
			return
		}
		switch msg[0] {
		case "GOTCONFIG":
			fmt.Println("receive router config response")
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
	toRouter.MessageLoop()
}
