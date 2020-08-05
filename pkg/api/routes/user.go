package routes

import (
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/ractf/andromeda/pkg/node"
	"github.com/ractf/andromeda/pkg/node/instance"
	"math/rand"
	"net/http"
	"sync"
)

type userRoutes struct {
	node  *node.Node
	users map[string]*instance.Instance
	mutex *sync.Mutex
}

func AddUserRoutes(node *node.Node, ws *restful.WebService) {
	u := userRoutes{
		node:  node,
		users: make(map[string]*instance.Instance),
		mutex: &sync.Mutex{},
	}

	ws.Route(ws.POST("/").To(Authenticated(u.getInstance, node.Config.ApiKey)))
}

type instanceRequest struct {
	Job  string `json:"job"`
	User string `json:"user"`
}

type instanceDetails struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

func (u *userRoutes) getInstance(request *restful.Request, response *restful.Response) {
	request.Request.Body = http.MaxBytesReader(response.ResponseWriter, request.Request.Body, 10240)
	dec := json.NewDecoder(request.Request.Body)
	var r instanceRequest
	err := dec.Decode(&r)
	if err != nil {
		fmt.Println(err)
		response.WriteError(500, err)
		return
	}

	job := r.Job
	jobSpec := u.node.GetJobSpecByName(job)

	u.mutex.Lock()
	defer u.mutex.Unlock()

	if instance, ok := u.users[r.User]; ok {
		if instance.Job == jobSpec && !instance.Dead {
			_ = response.WriteAsJson(instanceDetails{Port: instance.Port, Ip: u.node.Config.PublicIp})
			return
		}
	}

	instances := u.node.InstanceController.GetLocalInstancesOf(jobSpec)
	rand.Shuffle(len(instances), func(i, j int) { instances[i], instances[j] = instances[j], instances[i] })

	u.users[r.User] = instances[0]

	_ = response.WriteAsJson(instanceDetails{Port: instances[0].Port, Ip: u.node.Config.PublicIp})
}

func (u *userRoutes) resetInstance(request *restful.Request, response *restful.Response) {
	request.Request.Body = http.MaxBytesReader(response.ResponseWriter, request.Request.Body, 10240)
	dec := json.NewDecoder(request.Request.Body)
	var r instanceRequest
	err := dec.Decode(&r)
	if err != nil {
		fmt.Println(err)
		response.WriteError(500, err)
		return
	}

	job := r.Job
	jobSpec := u.node.GetJobSpecByName(job)

	u.mutex.Lock()
	defer u.mutex.Unlock()

	var avoid *instance.Instance = nil

	if instance, ok := u.users[r.User]; ok {
		if instance.Job == jobSpec {
			avoid = instance
		}
	}

	instances := u.node.InstanceController.GetLocalInstancesOf(jobSpec)
	rand.Shuffle(len(instances), func(i, j int) { instances[i], instances[j] = instances[j], instances[i] })
	index := 0
	if instances[0] == avoid && len(instances) > 1 {
		index = 1
	}

	u.users[r.User] = instances[index]

	_ = response.WriteAsJson(instanceDetails{Port: instances[index].Port, Ip: u.node.Config.PublicIp})
}
