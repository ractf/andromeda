package routes

import (
	"github.com/capnm/sysinfo"
	"github.com/emicklei/go-restful/v3"
	"github.com/ractf/andromeda/pkg/node"
)

type statusRoutes struct {
	node *node.Node
}

func AddStatusRoutes(node *node.Node, ws *restful.WebService) {
	a := statusRoutes{
		node: node,
	}

	ws.Route(ws.GET("/sysinfo").To(Authenticated(a.sysinfo, node.Config.ApiKey)))
	ws.Route(ws.GET("/instances").To(Authenticated(a.instances, node.Config.ApiKey)))
	ws.Route(ws.GET("/jobs").To(Authenticated(a.jobs, node.Config.ApiKey)))
}

func (s *statusRoutes) sysinfo(request *restful.Request, response *restful.Response) {
	si := sysinfo.Get()
	response.WriteAsJson(si)
}

func (s *statusRoutes) instances(request *restful.Request, response *restful.Response) {
	response.WriteAsJson(s.node.InstanceController.GetLocalInstances())
}

func (s *statusRoutes) jobs(request *restful.Request, response *restful.Response) {
	response.WriteAsJson(s.node.JobSpecList)
}
