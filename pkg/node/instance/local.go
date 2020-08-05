package instance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

type LocalInstanceController struct {
	ContainerClient ContainerClient
	Mutex           *sync.Mutex
	Instances       map[*JobSpec][]*Instance
	InstanceList    []*Instance
}

func (i LocalInstanceController) StartInstance(jobSpec *JobSpec) {
	instance, err := i.ContainerClient.StartContainer(jobSpec)
	if err != nil {
		fmt.Println(err)
		return
	}

	file, _ := json.MarshalIndent(instance, "", "")
	err = ioutil.WriteFile("/opt/andromeda/instances/"+instance.Container+".json", file, 0644)
	if err != nil {
		fmt.Println(err)
	}

	i.Mutex.Lock()
	instances, ok := i.Instances[jobSpec]
	if !ok {
		instances = make([]*Instance, 0)
	}
	i.Instances[jobSpec] = append(instances, &instance)
	i.Mutex.Unlock()
}

func (i LocalInstanceController) StopInstance(instance *Instance) {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()
	_ = i.ContainerClient.StopContainer(instance.Container)
	instance.Dead = true
}

func (i LocalInstanceController) GetLocalInstances() []*Instance {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()

	clone := make([]*Instance, len(i.InstanceList))
	copy(clone, i.InstanceList)

	return clone
}

func (i LocalInstanceController) GetLocalInstancesOf(jobSpec *JobSpec) []*Instance {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()

	clone := make([]*Instance, len(i.Instances[jobSpec]))
	copy(clone, i.Instances[jobSpec])

	return clone
}

func (i LocalInstanceController) LoadInstance(instance *Instance, jobSpec *JobSpec) {
	i.Mutex.Lock()
	instances, ok := i.Instances[jobSpec]
	if !ok {
		instances = make([]*Instance, 0)
	}
	i.Instances[jobSpec] = append(instances, instance)
	i.Mutex.Unlock()
}
