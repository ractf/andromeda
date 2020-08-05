package node

import (
	"github.com/ractf/andromeda/pkg/node/instance"
	"sync"
	"time"
)

type Node struct {
	Config             *Config
	jobSpecs           map[string]*instance.JobSpec
	jobSpecList        []*instance.JobSpec
	InstanceController instance.InstanceController
	mutex              *sync.Mutex
}

func StartNode(config *Config) *Node {
	node := Node{
		Config:      config,
		jobSpecs:    make(map[string]*instance.JobSpec),
		jobSpecList: make([]*instance.JobSpec, 0),
		mutex:       &sync.Mutex{},
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
	ticker := time.NewTicker(time.Second * time.Duration(30))
	for {
		select {
		case <-ticker.C:
			go n.HousekeepingTick()
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
