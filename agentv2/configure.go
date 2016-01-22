package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type BootstrapConfigure struct {
	Installation string                 `json:"installation"`
	Location     string                 `json:"location"`
	ZookeeprUri  string                 `json:"zookeeper-uri"`
	CarbonUri    []string               `json:"carbon-uri"`
	PortRanges   map[string]interface{} `json:"portRanges"`
}

func ParseBootstrap(p string) (*BootstrapConfigure, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	bc := &BootstrapConfigure{}
	err = json.Unmarshal(c, bc)
	return bc, err
}
