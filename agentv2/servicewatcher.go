package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// ============ misc
var (
	Id     int64 = 0
	IdLock sync.Mutex
)

func GetId() int64 {
	IdLock.Lock()
	defer IdLock.Unlock()
	Id += 1
	return Id
}

// ============= errors
type WatcherErrorWrapper struct {
	ErrorCode int
	Id        int64
	Message   string
}

func (this WatcherErrorWrapper) Error() string {
	return fmt.Sprintf("error occurred [ watcher = %v, err = %v, message = %v ]", this.Id, this.ErrorCode, this.Message)
}

func NewWatcherError(errorCode int, id int64, message string) WatcherErrorWrapper {
	return WatcherErrorWrapper{errorCode, id, message}
}

const (
	ErrorNilConnection = iota
	ErrorZookeeperAccessFail
)

type NodeStat struct {
	Value   []byte
	Version int32
}

type NodeEvent struct {
	Name  string
	Value []byte
}

type ZkWatcher struct {
	Id           int64
	Service      string
	Path         string
	Location     string
	EndpointType string

	Parent         *ServiceBase
	NewNotifier    chan *NodeEvent
	ChangeNotifier chan *NodeEvent
	DeleteNotifier chan *NodeEvent

	KnownNodes  map[string]*NodeStat
	ErrorReport chan error
	Quit        chan bool
	IsRunning   bool

	OnNewHandler    NotifyHandler
	OnChangeHandler NotifyHandler
	OnDeleteHandler NotifyHandler
	OnCloseHandler  NotifyHandler
}

func NewZkWatcher(parent *ServiceBase, service, location, endpointType string) *ZkWatcher {
	zw := &ZkWatcher{
		Path:         path.Join(parent.Root, "serviceClass", service),
		Service:      service,
		Location:     location,
		EndpointType: endpointType,
		ErrorReport:  parent.ErrorReport,

		Id:             GetId(),
		KnownNodes:     make(map[string]*NodeStat),
		Quit:           make(chan bool),
		IsRunning:      false,
		Parent:         parent,
		NewNotifier:    make(chan *NodeEvent),
		ChangeNotifier: make(chan *NodeEvent),
		DeleteNotifier: make(chan *NodeEvent),
	}
	log.Println(zw.DebugString())
	return zw
}

func (this *ZkWatcher) OnNew(handler NotifyHandler) *ZkWatcher {
	this.OnNewHandler = handler
	return this
}

func (this *ZkWatcher) OnChange(handler NotifyHandler) *ZkWatcher {
	this.OnChangeHandler = handler
	return this
}
func (this *ZkWatcher) OnDelete(handler NotifyHandler) *ZkWatcher {
	this.OnDeleteHandler = handler
	return this
}

func (this *ZkWatcher) OnClose(handler NotifyHandler) *ZkWatcher {
	this.OnCloseHandler = handler
	return this
}

func (this *ZkWatcher) Watch() *ZkWatcher {
	if !this.IsRunning {
		go this.HandleEvent()
		go func() {
			timer := time.NewTimer(time.Second)
			this.IsRunning = true
			for this.IsRunning {
				select {
				case <-timer.C:
					this.Scan()
					timer.Reset(5 * time.Second)
				case <-this.Quit:
					this.IsRunning = false
					return
				}
			}
		}()
	}
	return this
}

func (this *ZkWatcher) Diff(oldNodes map[string]*NodeStat, newNodes map[string]*NodeStat) {
	for k, v := range oldNodes {
		if stat, ok := newNodes[k]; ok {
			if stat.Version != v.Version {
				// exist in both but version changed = modified
				this.ChangeNotifier <- &NodeEvent{k, v.Value}
				log.Println(this, "changed service", k)
			}
		} else {
			// exist in old but new = deleted
			this.DeleteNotifier <- &NodeEvent{k, v.Value}
			log.Println(this, "deleted service", k)
		}
	}
	for k, v := range newNodes {
		if _, ok := oldNodes[k]; !ok {
			// exist in new but old = new
			this.NewNotifier <- &NodeEvent{k, v.Value}
			log.Println(this, "new service", k)
		}
	}
}

func (this *ZkWatcher) String() string {
	return fmt.Sprintf("watcher[%v]", this.Id)
}

func (this *ZkWatcher) DebugString() string {
	return fmt.Sprintf("watcher[id:%v service:%v location:%v path:%v endpointType:%v] ",
		this.Id, this.Service, this.Location, this.Path, this.EndpointType)
}

// 扫描后得到差集
func (this *ZkWatcher) Scan() {
	if this.Parent.Conn == nil {
		this.ErrorReport <- NewWatcherError(ErrorNilConnection, this.Id, "")
		return
	}
	children, _, err := this.Parent.Conn.Children(this.Path)
	if err != nil {
		this.ErrorReport <- NewWatcherError(ErrorZookeeperAccessFail, this.Id, err.Error())
		return
	}
	curNodes := make(map[string]*NodeStat)
	for _, node := range children {
		data, stat, err := this.Parent.Conn.Get(path.Join(this.Path, node))
		if err != nil {
			this.ErrorReport <- NewWatcherError(ErrorZookeeperAccessFail, this.Id, err.Error())
			continue
		}
		curNodes[node] = &NodeStat{
			Value:   data,
			Version: stat.Version,
		}
	}
	this.Diff(this.KnownNodes, curNodes)
	this.KnownNodes = curNodes
}

func (this *ZkWatcher) parseServicePath(data []byte) (string, error) {
	info := make(map[string]string)
	if err := json.Unmarshal(data, &info); err != nil {
		return "", err
	} else {
		if this.Location == "*" || info["serviceLocation"] == this.Location {
			return info["servicePath"], nil
		} else {
			return "", errors.New("unacceptable location")
		}
	}
}

func (this *ZkWatcher) parseServiceUri(servicePath string, endpointType string) (string, error) {
	curServicePath := path.Join(this.Parent.Root, servicePath)
	children, _, err := this.Parent.Conn.Children(curServicePath)
	if err != nil {
		return "", err
	}
	// eg. [logger, agents, pacing]
	for _, endpoint := range children {
		if endpoint != endpointType {
			continue
		}
		curEndpointPath := path.Join(curServicePath, endpoint)
		connectionTypes, _, err := this.Parent.Conn.Children(curEndpointPath)
		if err != nil {
			return "", err
		}
		// all connect url, pick one but 127.0.0.1 out
		for _, ct := range connectionTypes {
			if node, _, err := this.Parent.Conn.Get(path.Join(curEndpointPath, ct)); err != nil {
				return "", err
			} else {
				addr, err := parseAddress(node)
				if err != nil {
					return "", err
				}
				return addr, nil
			}
		}
	}
	return "", errors.New("uri not found")
}

func parseAddress(content []byte) (string, error) {
	info := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(content, &info); err != nil {
		return "", err
	} else {
		for _, i := range info {
			if zmqAddr, ok := i["zmqConnectUri"]; ok {
				if addr, ok := zmqAddr.(string); ok && strings.Index(addr, "127.0.0.1") < 0 {
					return addr, nil
				}
			}
		}
		return "", errors.New("not found")
	}
}

func (this *ZkWatcher) Close() {
	this.Quit <- true
	if this.OnCloseHandler != nil {
		this.OnCloseHandler(this.Service, "")
	}
}

func (this *ZkWatcher) HandleEvent() {
	for {
		var handler NotifyHandler = nil
		var event *NodeEvent = nil
		select {
		case event = <-this.NewNotifier:
			handler = this.OnNewHandler
		case event = <-this.ChangeNotifier:
			handler = this.OnChangeHandler
		case event = <-this.DeleteNotifier:
			handler = this.OnDeleteHandler
		}
		if handler != nil && event != nil {
			if p, err := this.parseServicePath(event.Value); err == nil {
				if uri, err := this.parseServiceUri(p, this.EndpointType); err == nil {
					if handler != nil {
						handler(event.Name, uri)
					}
				} else {
					log.Println("parse new service uri error:", err.Error())
				}
			} else {
				log.Println("pase new service path error:", err.Error())
			}
		}
	}
}

type NotifyHandler func(serviceName string, servicePath string)

type ServiceBase struct {
	Hosts    []string
	Root     string
	Location string

	// Connection
	Conn        *zk.Conn
	ConnTimeout time.Duration

	// ZkWatcher
	Watchers     map[string]*ZkWatcher
	watchersLock sync.Mutex

	ErrorReport chan error
}

func NewServiceBase(hosts []string, root, location string) *ServiceBase {
	root = strings.TrimSpace(root)
	if root[0] != '/' {
		root = "/" + root
	}
	return &ServiceBase{
		Hosts:       hosts,
		Root:        root,
		Location:    location,
		ConnTimeout: 5 * time.Second,
		Watchers:    make(map[string]*ZkWatcher),
		ErrorReport: make(chan error),
	}
}

func (this *ServiceBase) Connect() (err error) {
	if this.Conn != nil {
		this.Conn.Close()
	}
	this.Conn, _, err = zk.Connect(this.Hosts, this.ConnTimeout)
	return err
}

func (this *ServiceBase) AutoRecovery() {
	for event := range this.ErrorReport {
		switch t := event.(type) {
		case WatcherErrorWrapper:
			log.Println(t.Error())
			if t.ErrorCode == ErrorNilConnection {
				err := this.Connect()
				if err != nil {
					log.Printf("reconnect, err = %v", err.Error())
				}
			}
		case error:
			log.Println(t.Error())
		default:
			log.Printf("unknown error: %v", t)
		}
	}
}

func (this *ServiceBase) NewWatcher(service string, endpointType string) *ZkWatcher {
	this.watchersLock.Lock()
	defer this.watchersLock.Unlock()
	if w, ok := this.Watchers[service]; ok {
		w.Close()
	}
	w := NewZkWatcher(this, service, this.Location, endpointType)
	this.Watchers[service] = w
	return w
}
