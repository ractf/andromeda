package api

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/ractf/andromeda/pkg/api/routes"
	"github.com/ractf/andromeda/pkg/node"
	"net/http"
	"strconv"
)

type Server struct {
	Node *node.Node
}

func (s *Server) StartAPIServer(config *node.Config) error {
	ws := new(restful.WebService)

	routes.AddUserRoutes(s.Node, ws)
	routes.AddJobRoutes(s.Node, ws)
	routes.AddStatusRoutes(s.Node, ws)

	restful.Add(ws)

	address := config.ApiIp + ":" + strconv.Itoa(config.ApiPort)
	fmt.Println("Listening on", address)
	return http.ListenAndServe(address, nil)
}
