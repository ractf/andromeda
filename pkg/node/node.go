package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ractf/andromeda/pkg/node/instance"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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

	os.Mkdir("/opt/andromeda/", 0644)
	os.Mkdir("/opt/andromeda/jobs/", 0644)
	os.Mkdir("/opt/andromeda/instances/", 0644)

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
		n.InstanceController.LoadInstance(inst)
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
	n.refreshConfig()
	for _, spec := range n.JobSpecList {
		instances := n.InstanceController.GetLocalInstancesOf(spec)

		if len(instances) != spec.Replicas {
			for i := 0; i < spec.Replicas-len(instances); i++ {
				go n.InstanceController.StartInstance(spec)
			}
		}

		for _, i := range instances {
			//TODO: Make this configurable
			healthy := n.isPortOpen(i.Port, 3)
			if !healthy && i.Healthy {
				n.postWebhook(i)
			}
			i.Healthy = healthy
		}

		if !n.InstanceController.IsJobUpToDate(spec) {
			for _, i := range n.InstanceController.GetLocalInstancesOf(spec) {
				go n.InstanceController.RestartInstance(i)
			}
		}
	}
}

func (n *Node) isPortOpen(port int, timeout int) bool {
	connectTimeout := time.Duration(timeout) * time.Second
	conn, _ := net.DialTimeout("tcp", net.JoinHostPort(n.Config.PublicIp, strconv.Itoa(port)), connectTimeout)

	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func (n *Node) postWebhook(i *instance.Instance) {
	data := "{\"embeds\": [{\"title\": \"Container Failed Healthcheck\",\"description\": \"Job: " + i.Job.Name + "\nIp:" + n.Config.PublicIp + "\nPort: " + strconv.Itoa(i.Port) + "\nContainer: " + i.Container + "\",\"color\": 9833227}]}"
	_, err := http.Post(n.Config.DiscordWebhookUrl, "application/json", bytes.NewBufferString(data))
	if err != nil {
		fmt.Println(err)
	}
}

func (n *Node) refreshConfig() {
	if !n.Config.RefreshConfig {
		return
	}
	configFile, err := os.Open(n.Config.ConfigPath)
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
	config.ConfigPath = n.Config.ConfigPath
	n.Config = &config
}
