package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	NewConfigure    chan *CreativeConfig
	RemoveConfigure chan *CreativeConfig

	availableAgentConfigures *map[string]*CreativeConfig
	enabledAgentConfigures   *map[string]*CreativeConfig
	configureLock            sync.Mutex
)

func init() {
	NewConfigure = make(chan *CreativeConfig)
	RemoveConfigure = make(chan *CreativeConfig)

	go AutoApplyConfigure()
}

type CreativeConfig struct {
	Name      string
	Configure map[string]interface{}
	Raw       []byte
	Md5       string
}

func NewCreativeConfig(name, filePath string) (*CreativeConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	h := md5.New()
	h.Write(c)
	md5 := hex.EncodeToString(h.Sum(nil))
	creativeJson := make(map[string]interface{})
	if err := json.Unmarshal(c, &creativeJson); err != nil {
		return nil, err
	}
	config := &CreativeConfig{
		Name:      name,
		Configure: creativeJson,
		Raw:       c,
		Md5:       md5,
	}
	return config, nil
}

func LoadCreatives(dir string) {
	configures := make(map[string]*CreativeConfig)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".json") && !strings.HasPrefix(info.Name(), ".") {
			config, err := NewCreativeConfig(info.Name(), path)
			if err != nil {
				log.Println("load agent configure", path, "failed:", err.Error())
			} else {
				configures[info.Name()] = config
			}
		}
		return nil
	})
	log.Println("scan configure done,", len(configures), "available")
	availableAgentConfigures = &configures
}

func CreativeDiff(cur, old *map[string]*CreativeConfig) ([]*CreativeConfig, []*CreativeConfig) {
	added := make([]*CreativeConfig, 0)
	removed := make([]*CreativeConfig, 0)
	if cur != nil {
		for k, v := range *cur {
			// [NewOrChanged]
			if old != nil {
				if cc, ok := (*old)[k]; !ok || v.Md5 != cc.Md5 {
					added = append(added, v)
				}
			} else {
				added = append(added, v)
			}
		}
	}
	if old != nil {
		for k, v := range *old {
			if cur != nil {
				if _, ok := (*cur)[k]; !ok {
					removed = append(removed, v)
				}
			} else {
				removed = append(removed, v)
			}
		}
	}
	return added, removed
}

func AutoApplyConfigure() {
	configureTimer := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-configureTimer.C:
			LoadCreatives("./agents")
			added, removed := CreativeDiff(availableAgentConfigures, enabledAgentConfigures)
			for _, config := range added {
				NewConfigure <- config
			}
			for _, config := range removed {
				RemoveConfigure <- config
			}
			SetEnableConfigures(availableAgentConfigures)
			configureTimer.Reset(5 * time.Second)
		}
	}
}

func SetEnableConfigures(newValue *map[string]*CreativeConfig) {
	configureLock.Lock()
	defer configureLock.Unlock()
	enabledAgentConfigures = newValue
}

func GetEnableConfigures() *map[string]*CreativeConfig {
	configureLock.Lock()
	defer configureLock.Unlock()
	return enabledAgentConfigures
}
