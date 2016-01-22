package servicebase

import (
	"sync"
	"time"

	zmq "github.com/yangzhao28/zmq3"
)

//
//
//
type ZmqRouter struct {
	Name            string
	Connections     map[string]*zmq.Socket
	ConnectionNames map[*zmq.Socket]string
	Poller          *zmq.Poller
	Quit            bool
	OnRecvMessage   func(name string, socket *zmq.Socket)

	lock sync.Mutex
}

func NewZmqRouter(name string) *ZmqRouter {
	return &ZmqRouter{
		Name:            name,
		Connections:     make(map[string]*zmq.Socket),
		ConnectionNames: make(map[*zmq.Socket]string),
		Poller:          zmq.NewPoller(),
		Quit:            bool,
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
	if socketType == zmq.XSUB || socketType == zmq.SUB {
		sock.SetSubscribe("")
	}
	if err := sock.Connect(host); err != nil {
		return err
	}
	this.Poller.Add(sock, zmq.POLLIN)
	this.Connections[name] = sock
	this.ConnectionNames[sock] = name
	return nil
}

func (this *ZmqRouter) CloseConnection(name string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if conn, ok := this.Connections[name]; ok {
		conn.Close()
		delete(this.Connections, name)
		delete(this.ConnectionNames, conn)
	}
}

func (this *ZmqRouter) SendMessageTo(name string, msg ...interface{}) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if conn, ok := this.Connections[name]; ok {
		if _, err := conn.SendMessage(msg...); err != nil {
			return err
		}
	}
	return nil
}

func (this *ZmqRouter) SendMessage(msg ...interface{}) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, sock := range this.Connections {
		if _, err := sock.SendMessage(msg...); err != nil {
			return err
		}
	}
	return nil
}

func (this *ZmqRouter) Close() {
	this.CloseConnection()
	this.Quit = true
}

func (this *ZmqRouter) MessageLoop() {
	this.Quit = false
	for !this.Quit {
		polled, err := this.Poller.Poll(time.Second)
		if err != nil {
			continue
		}
		if this.OnRecvMessage != nil && len(polled) > 0 {
			for _, p := range polled {
				if name, ok := this.ConnectionNames[p.Socket]; ok {
					this.OnRecvMessage(name, p.Socket)
				}
			}
		}
	}
}
