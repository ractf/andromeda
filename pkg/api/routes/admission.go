package routes

import (
	"encoding/json"
	"github.com/emicklei/go-restful/v3"
	"github.com/ractf/andromeda/pkg/node"
	"github.com/ractf/andromeda/pkg/node/instance"
	"net/http"
)

type admissionRoutes struct {
	node *node.Node
}

func AddAdmissionRoutes(node *node.Node, ws *restful.WebService) {
	a := admissionRoutes{
		node: node,
	}

	ws.Route(ws.POST("/jobs").To(Authenticated(a.submitJob, node.Config.ApiKey)))
}

func (a *admissionRoutes) submitJob(request *restful.Request, response *restful.Response) {
	request.Request.Body = http.MaxBytesReader(response.ResponseWriter, request.Request.Body, 10240)
	dec := json.NewDecoder(request.Request.Body)
	dec.DisallowUnknownFields()
	var spec instance.JobSpec
	err := dec.Decode(&spec)
	if err != nil {
		_ = response.WriteError(500, err)
	}
	a.node.SubmitJobSpec(&spec)
}
