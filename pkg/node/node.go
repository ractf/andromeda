package node

import (
	"encoding/json"
	"github.com/ractf/andromeda/pkg/node/instance"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Node struct {
	Config             *Config
	jobSpecs           map[string]*instance.JobSpec
	jobSpecList        []*instance.JobSpec
	InstanceController instance.InstanceController
	mutex              *sync.Mutex
	configPath         string
}

func StartNode(config *Config, configPath string) *Node {
	node := Node{
		Config:      config,
		jobSpecs:    make(map[string]*instance.JobSpec),
		jobSpecList: make([]*instance.JobSpec, 0),
		mutex:       &sync.Mutex{},
		configPath:  configPath,
	}

	instanceController := instance.LocalInstanceController{
		ContainerClient: instance.CreateDockerClient(config.DefaultRegistryAuth),
		Mutex:           &sync.Mutex{},
		Instances:       make(map[*instance.JobSpec][]*instance.Instance),
	}
	node.InstanceController = instanceController

	go node.HousekeepingLoop()

	return &node
}

func (n *Node) GetJobSpecByName(name string) *instance.JobSpec {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.jobSpecs[name]
}

func (n *Node) SubmitJobSpec(spec *instance.JobSpec) {
	n.mutex.Lock()
	n.jobSpecs[spec.Name] = spec
	n.jobSpecList = append(n.jobSpecList, spec)
	n.mutex.Unlock()
	for i := 0; i < spec.Replicas; i++ {
		go n.InstanceController.StartInstance(spec)
	}
}

func (n *Node) HousekeepingLoop() {
	go n.HousekeepingTick()
	housekeepingTicker := time.NewTicker(time.Second * time.Duration(30))
	configRefreshTicker := time.NewTicker(time.Second * time.Duration(30))
	for {
		select {
		case <-housekeepingTicker.C:
			go n.HousekeepingTick()
		case <-configRefreshTicker.C:
			go n.refreshConfig()
		}
	}
}

func (n *Node) HousekeepingTick() {
	for _, spec := range n.jobSpecList {
		instances := n.InstanceController.GetLocalInstancesOf(spec)

		for _, instance := range instances {
			_ = instance
		}
	}
}

func (n *Node) refreshConfig() {
	configFile, err := os.Open(n.configPath)
	if err != nil {
		return
	}
	defer configFile.Close()

	bytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	n.Config = &config
}
