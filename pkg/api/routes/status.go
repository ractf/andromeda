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
}

func (s *statusRoutes) sysinfo(request *restful.Request, response *restful.Response) {
	si := sysinfo.Get()
	_ = response.WriteAsJson(si)
}
