package instance

import "github.com/docker/docker/api/types"

type InstanceController interface {
	StartInstance(jobSpec *JobSpec)
	StopInstance(instance *Instance)
	RestartInstance(instance *Instance)
	GetLocalInstances() []*Instance
	GetLocalInstancesOf(jobSpec *JobSpec) []*Instance
	LoadInstance(instance *Instance)
}

type JobSpec struct {
	Uuid         string           `json:"uuid,omitempty"`
	Name         string           `json:"name,omitempty"`
	Port         int              `json:"port,omitempty"`
	Replicas     int              `json:"replicas,omitempty"`
	Resources    ResourceLimit    `json:"resources,omitempty"`
	ImageName    string           `json:"image,omitempty"`
	RegistryAuth types.AuthConfig `json:"registryAuth,omitempty"`
	Env          []string         `json:"env,omitempty"`
}

type ResourceLimit struct {
	MemLimit int64  `json:"memory,omitempty"`
	CPUs     string `json:"cpus,omitempty"`
}

type Instance struct {
	Job       *JobSpec
	JobId     string `json:"jobid"`
	Container string `json:"container"`
	Port      int    `json:"Port"`
	Dead      bool
}
