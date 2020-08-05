package node

import (
	"encoding/json"
	"fmt"
	"github.com/ractf/andromeda/pkg/node/instance"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type Node struct {
	Config             *Config
	jobSpecs           map[string]*instance.JobSpec
	JobSpecList        []*instance.JobSpec
	InstanceController instance.InstanceController
	mutex              *sync.Mutex
}

func StartNode(config *Config) *Node {
	node := Node{
		Config:      config,
		jobSpecs:    make(map[string]*instance.JobSpec),
		JobSpecList: make([]*instance.JobSpec, 0),
		mutex:       &sync.Mutex{},
	}

	instanceController := instance.LocalInstanceController{
		ContainerClient: instance.CreateDockerClient(config.DefaultRegistryAuth),
		Mutex:           &sync.Mutex{},
		Instances:       make(map[*instance.JobSpec][]*instance.Instance),
	}
	node.InstanceController = instanceController

	node.loadJobs()
	node.loadInstances()
	go node.HousekeepingLoop()

	return &node
}

func (n *Node) loadJobs() {
	files, err := ioutil.ReadDir("/opt/andromeda/jobs/")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		jsonFile, err := os.Open("/opt/andromeda/jobs/" + file.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		bytes, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var jobSpec *instance.JobSpec
		err = json.Unmarshal(bytes, &jobSpec)
		jsonFile.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}

		n.SubmitJobSpec(jobSpec)
	}
}

func (n *Node) loadInstances() {
	files, err := ioutil.ReadDir("/opt/andromeda/instances/")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		jsonFile, err := os.Open("/opt/andromeda/instances/" + file.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		bytes, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var inst *instance.Instance
		err = json.Unmarshal(bytes, &inst)
		jsonFile.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}

		inst.Job = n.jobSpecs[inst.JobId]
		n.InstanceController.LoadInstance(inst, inst.Job)
	}
}

func (n *Node) GetJobSpecByUuid(uuid string) *instance.JobSpec {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.jobSpecs[uuid]
}

func (n *Node) SubmitJobSpec(spec *instance.JobSpec) {
	n.mutex.Lock()
	n.jobSpecs[spec.Uuid] = spec
	n.JobSpecList = append(n.JobSpecList, spec)
	n.mutex.Unlock()

	file, _ := json.MarshalIndent(spec, "", "")
	err := ioutil.WriteFile("/opt/andromeda/jobs/"+spec.Uuid+".json", file, 0644)
	if err != nil {
		fmt.Println(err)
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
	for _, spec := range n.JobSpecList {
		instances := n.InstanceController.GetLocalInstancesOf(spec)

		if len(instances) != spec.Replicas {
			for i := 0; i < spec.Replicas-len(instances); i++ {
				go n.InstanceController.StartInstance(spec)
			}
		}
	}
}
