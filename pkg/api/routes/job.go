package routes

import (
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/emicklei/go-restful/v3"
	"github.com/ractf/andromeda/pkg/node"
	"github.com/ractf/andromeda/pkg/node/instance"
	"net/http"
)

type jobRoutes struct {
	node *node.Node
}

func AddJobRoutes(node *node.Node, ws *restful.WebService) {
	j := jobRoutes{
		node: node,
	}

	ws.Route(ws.POST("/job/submit").To(Authenticated(j.submitJob, node.Config.ApiKey)))
	ws.Route(ws.POST("/job/{id}/restart").To(Authenticated(j.restartJob, node.Config.ApiKey)))
}

func (j *jobRoutes) submitJob(request *restful.Request, response *restful.Response) {
	request.Request.Body = http.MaxBytesReader(response.ResponseWriter, request.Request.Body, 10240)
	dec := json.NewDecoder(request.Request.Body)
	dec.DisallowUnknownFields()
	var spec instance.JobSpec
	err := dec.Decode(&spec)
	if err != nil {
		_ = response.WriteError(500, err)
	}
	spec.Uuid = uuid.Generate().String()
	j.node.SubmitJobSpec(&spec)

	response.Write([]byte(fmt.Sprintf("{\"id\":\"%v\"}", spec.Uuid)))
}

func (j *jobRoutes) restartJob(request *restful.Request, response *restful.Response) {
	jobId := request.PathParameter("id")
	jobSpec := j.node.GetJobSpecByUuid(jobId)
	for _, i := range j.node.InstanceController.GetLocalInstancesOf(jobSpec) {
		go j.node.InstanceController.RestartInstance(i)
	}
}
