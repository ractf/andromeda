package instance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Client struct {
	docker      *client.Client
	bindIp      string
	defaultAuth types.AuthConfig
}

func CreateDockerClient(defaultAuth types.AuthConfig) ContainerClient {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return &Client{docker: cli, defaultAuth: defaultAuth}
}

func (c *Client) PullImage(spec *JobSpec) error {
	ctx := context.Background()
	encodedJSON, err := json.Marshal(spec.RegistryAuth)
	if spec.RegistryAuth == (types.AuthConfig{}) {
		encodedJSON, err = json.Marshal(c.defaultAuth)
	}
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	readCloser, err := c.docker.ImagePull(ctx, spec.ImageName, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}

	io.Copy(os.Stdout, readCloser)

	defer readCloser.Close()
	return nil
}

func (c *Client) setupNetwork(spec *JobSpec) (map[nat.Port][]nat.PortBinding, map[nat.Port]struct{}, int) {
	portBindings := make(map[nat.Port][]nat.PortBinding)
	portSet := make(map[nat.Port]struct{})
	portNum := 0

	if spec.Port != 0 {
		tcpPort, _ := nat.NewPort("tcp", strconv.Itoa(spec.Port))
		portNum = rand.Intn(55535) + 10000
		assignedPort := strconv.Itoa(portNum)

		portBindings[tcpPort] = []nat.PortBinding{{
			HostIP:   c.bindIp,
			HostPort: assignedPort,
		}}
		portSet[tcpPort] = struct{}{}
	}
	return portBindings, portSet, portNum
}

func (c *Client) cloneNetwork(instance *Instance) (map[nat.Port][]nat.PortBinding, map[nat.Port]struct{}, int) {
	portBindings := make(map[nat.Port][]nat.PortBinding)
	portSet := make(map[nat.Port]struct{})
	portNum := 0

	if instance.Job.Port != 0 {
		tcpPort, _ := nat.NewPort("tcp", strconv.Itoa(instance.Job.Port))
		portNum = instance.Port
		assignedPort := strconv.Itoa(portNum)

		portBindings[tcpPort] = []nat.PortBinding{{
			HostIP:   c.bindIp,
			HostPort: assignedPort,
		}}
		portSet[tcpPort] = struct{}{}
	}
	return portBindings, portSet, portNum
}

func (c *Client) StartContainer(spec *JobSpec) (Instance, error) {
	err := c.PullImage(spec)
	if err != nil {
		return Instance{}, err
	}

	portBindings, portSet, portNum := c.setupNetwork(spec)

	return c.StartContainerWithNetwork(spec, portSet, portBindings, portNum)
}

func (c *Client) StartContainerWithNetwork(spec *JobSpec, portSet map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, portNum int) (Instance, error) {
	ctx := context.Background()

	containerConfig := container.Config{
		Image:        spec.ImageName,
		ExposedPorts: portSet,
		Env:          spec.Env,
	}

	cpus, err := strconv.ParseFloat(spec.Resources.CPUs, 64)
	if err != nil {
		return Instance{}, err
	}

	hostConfig := container.HostConfig{
		PortBindings: portBindings,
		Resources: container.Resources{
			Memory:   spec.Resources.MemLimit,
			NanoCPUs: int64(cpus * 1000000000),
		},
	}

	resp, err := c.docker.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, "")
	if err != nil {
		return Instance{}, err
	}

	err = c.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return Instance{}, err
	}

	return Instance{
		Job:       spec,
		JobId:     spec.Uuid,
		Container: resp.ID,
		Port:      portNum,
	}, nil
}

func (c *Client) StopContainer(id string) error {
	ctx := context.Background()
	timeout := time.Duration(5) * time.Second
	return c.docker.ContainerStop(ctx, id, &timeout)
}

func (c *Client) RestartContainer(instance *Instance) error {
	ctx := context.Background()
	timeout := time.Duration(5) * time.Second
	err := c.docker.ContainerStop(ctx, instance.Container, &timeout)
	if err != nil {
		return err
	}

	c.docker.ContainerRemove(ctx, instance.Container, types.ContainerRemoveOptions{})

	portBindings, portSet, portNum := c.cloneNetwork(instance)

	spec := instance.Job
	err = c.PullImage(spec)
	if err != nil {
		return err
	}

	newInstance, err := c.StartContainerWithNetwork(spec, portSet, portBindings, portNum)
	if err != nil {
		return err
	}

	instance.Container = newInstance.Container
	return nil
}

func (c *Client) IsImageUpToDate(spec *JobSpec) bool {
	ctx := context.Background()
	imageInspect, _, err := c.docker.ImageInspectWithRaw(ctx, spec.ImageName)
	if err != nil {
		return true
	}
	c.PullImage(spec)
	imageInspect2, _, err := c.docker.ImageInspectWithRaw(ctx, spec.ImageName)
	if err != nil {
		return true
	}
	if imageInspect2.ID != imageInspect.ID {
		return false
	}
	return true
}
