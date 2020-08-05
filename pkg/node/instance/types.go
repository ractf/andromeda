package instance

import "github.com/docker/docker/api/types"

type InstanceController interface {
	StartInstance(jobSpec *JobSpec)
	StopInstance(instance *Instance)
	GetLocalInstances() []*Instance
	GetLocalInstancesOf(jobSpec *JobSpec) []*Instance
}

type JobSpec struct {
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
	Container string
	Port      int
	Dead      bool
}
