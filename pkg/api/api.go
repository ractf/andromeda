package api

import (
	"encoding/json"
	"fmt"
	"github.com/capnm/sysinfo"
	"github.com/emicklei/go-restful/v3"
	"github.com/ractf/andromeda/pkg/challenge"
	"net/http"
	"os"
)

type Server struct {
	Instances *challenge.Instances
}

func (s *Server) StartServer(address string) error {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/status").To(Authenticated(s.status)))
	ws.Route(ws.GET("/user/{user_id}").To(Authenticated(s.user)))
	ws.Route(ws.POST("/disconnect/{user_id}").To(Authenticated(s.disconnect)))
	ws.Route(ws.POST("/reset/{user_id}").To(Authenticated(s.reset)))
	ws.Route(ws.POST("/").To(Authenticated(s.getInstance)))
	restful.Add(ws)
	return http.ListenAndServe(address, nil)
}

func Authenticated(function restful.RouteFunction) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		apiKey, exists := os.LookupEnv("API_KEY")
		if !exists || request.HeaderParameter("Authorization") != apiKey {
			_ = response.WriteErrorString(http.StatusForbidden, "Incorrect api key")
			return
		}
		function(request, response)
	}
}

func (s *Server) status(request *restful.Request, response *restful.Response) {
	si := sysinfo.Get()
	_ = response.WriteAsJson(si)
}

type instanceDetails struct {
	Port      string `json:"port"`
	Challenge string `json:"challenge"`
}

func (s *Server) user(request *restful.Request, response *restful.Response) {
	instance := s.Instances.GetCurrentUserInstance(request.PathParameter("user_id"))
	if instance.Users == nil {
		_ = response.WriteErrorString(http.StatusNotFound, "User has no instance")
		return
	}
	_ = response.WriteAsJson(instanceDetails{Port: instance.Port, Challenge: instance.Challenge.Name})
}

func (s *Server) disconnect(request *restful.Request, response *restful.Response) {
	s.Instances.Disconnect(request.PathParameter("user_id"))
}

func (s *Server) reset(request *restful.Request, response *restful.Response) {
	user := request.PathParameter("user_id")
	instance := s.Instances.GetCurrentUserInstance(request.PathParameter("user_id"))

	s.Instances.AvoidInstance(user, instance)
	s.Instances.Disconnect(user)
	i, err := s.Instances.GetInstanceForUser(user, instance.Challenge)
	if err != nil {
		return
	}
	fmt.Println(i)

	_ = response.WriteAsJson(instanceDetails{Port: i.Port, Challenge: i.Challenge.Name})
}

type instanceRequest struct {
	User      string `json:"user"`
	Challenge string `json:"challenge"`
}

func (s *Server) getInstance(request *restful.Request, response *restful.Response) {
	request.Request.Body = http.MaxBytesReader(response.ResponseWriter, request.Request.Body, 10240)
	dec := json.NewDecoder(request.Request.Body)
	dec.DisallowUnknownFields()
	var r instanceRequest
	err := dec.Decode(&r)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := r.User
	chal := r.Challenge

	spec := s.Instances.GetChallengeByName(chal)

	instance, err := s.Instances.GetInstanceForUser(user, spec)
	if err != nil {
		_ = response.WriteErrorString(http.StatusInternalServerError, "Something went wrong")
		fmt.Println(err)
		return
	}

	_ = response.WriteAsJson(instanceDetails{Port: instance.Port, Challenge: instance.Challenge.Name})
}
