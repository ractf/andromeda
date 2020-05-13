package challenge

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Client struct {
	docker *client.Client
}

func CreateDockerClient() Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return Client{docker: cli}
}

func (c *Client) BuildImage(spec *Spec) error {
	fmt.Println("Building image", spec.Name)
	ctx := context.Background()

	cmd := exec.Command("/bin/bash", "-c", "tar -cvf .ractf.tar *")
	cmd.Dir = spec.Path
	err := cmd.Run()
	if err != nil {
		return err
	}

	reader, err := os.Open(spec.Path + "/.ractf.tar")
	if err != nil {
		return err
	}

	_, err = c.docker.ImageBuild(ctx, reader, types.ImageBuildOptions{
		SuppressOutput: true,
		PullParent:     true,
		Tags:           []string{spec.ImageName},
		Dockerfile:     "Dockerfile",
		Memory:         int64(spec.MemLimit * 1024 * 1024),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) StartContainer(spec Spec, bindIp string) (Instance, error) {
	ctx := context.Background()
	port, _ := nat.NewPort("tcp", strconv.Itoa(spec.Port))
	assignedPort := strconv.Itoa(rand.Intn(55535) + 10000)

	containerConfig := container.Config{
		Image: spec.ImageName,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
	}

	hostConfig := container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   bindIp,
					HostPort: assignedPort,
				},
			},
		},
	}

	resp, err := c.docker.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, spec.ImageName+assignedPort)
	if err != nil {
		return Instance{}, err
	}

	err = c.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return Instance{}, err
	}

	return Instance{
		Port:      assignedPort,
		Challenge: spec,
		Users:     make([]string, 0),
		Container: resp.ID,
	}, nil
}

func (c *Client) StopContainer(id string) error {
	ctx := context.Background()
	timeout := time.Duration(5) * time.Second
	return c.docker.ContainerStop(ctx, id, &timeout)
}
