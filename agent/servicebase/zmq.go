package servicebase

import (
	"sync"
	"time"

	zmq "github.com/yangzhao28/zmq3"
)

type ZmqEventHandler func(socket *zmq.Socket)

type ZmqRouter struct {
	Name        string
	Connections map[string]*zmq.Socket
	Poller      *zmq.Poller
	lock        sync.Mutex

	OnRecvMessage ZmqEventHandler
}

func NewZmqRouter(name string) *ZmqRouter {
	return &ZmqRouter{
		Name:        name,
		Connections: make(map[string]*zmq.Socket),
		Poller:      zmq.NewPoller(),
	}
}

func (this *ZmqRouter) NewConnection(name string, host string, socketType zmq.Type) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	sock, err := zmq.NewSocket(socketType)
	if err != nil {
		return err
	}
	sock.SetIdentity(this.Name)
	if err := sock.Connect(host); err != nil {
		return err
	}
	this.Poller.Add(sock, zmq.POLLIN)
	this.Connections[name] = sock
	return nil
}

func (this *ZmqRouter) SendMessage(name string, msg ...interface{}) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if conn, ok := this.Connections[name]; ok {
		if _, err := conn.SendMessage(msg...); err != nil {
			return err
		}
	}
	return nil
}

func (this *ZmqRouter) SendMessageToAll(msg ...interface{}) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, sock := range this.Connections {
		if _, err := sock.SendMessage(msg...); err != nil {
			return err
		}
	}
	return nil
}

func (this *ZmqRouter) MessageLoop() {
	for {
		polled, err := this.Poller.Poll(time.Second)
		if err != nil {
			continue
		}
		if this.OnRecvMessage != nil && len(polled) > 0 {
			for _, p := range polled {
				this.OnRecvMessage(p.Socket)
			}
		}
	}
}

/*
func main() {
	zk := NewServiceBase([]string{"0.0.0.0:2181"}, "mtl")
	// agent config
	configService, _ := zk.GetEndpointAddress("rtbAgentConfiguration", "agents")

	rand.Seed(time.Now().Unix())
	name := "faker_" + strconv.Itoa(rand.Int())

	r := NewZmqRouter(name)
	r.NewConnection(configService, zmq.DEALER)
	r.SendMessage("CONFIG", name, jsConfig)

	routerService, _ := zk.GetEndpointAddress("rtbRequestRouter", "agents")
	s := NewZmqRouter(name)
	s.OnConnect = func(connectTo string, sock *zmq.Socket) {
		sock.SendMessage(connectTo, "CONFIG", name)
	}
	s.NewConnection(routerService, zmq.DEALER)
	s.OnRecvMessage = func(sock *zmq.Socket) {
		msg, err := sock.RecvMessage(0)
		if err != nil {
			return
		}
		fmt.Println(msg)
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
		}
	}
	s.MessageLoop()
}
*/
