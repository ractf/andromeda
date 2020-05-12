package api

import (
	"encoding/json"
	"fmt"
	"github.com/capnm/sysinfo"
	"github.com/emicklei/go-restful/v3"
	"net/http"
	"ractf.co.uk/andromeda/challenge"
)

type Server struct {
	Instances *challenge.Instances
}

func (s *Server) StartServer() error {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/status").To(s.status))
	ws.Route(ws.GET("/user/{user_id}").To(s.user))
	ws.Route(ws.POST("/disconnect/{user_id}").To(s.disconnect))
	ws.Route(ws.POST("/reset/{user_id}").To(s.reset))
	ws.Route(ws.POST("/").To(s.getInstance))
	restful.Add(ws)
	return http.ListenAndServe(":6000", nil)
}

func isAuthenticated(request *restful.Request, response *restful.Response) bool {
	/*apiKey, exists := os.LookupEnv("API_KEY")
	if !exists || request.HeaderParameter("Authorization") != apiKey {
		_ = response.WriteErrorString(http.StatusForbidden, "Incorrect api key")
		return false
	}*/
	return true
}

func (s *Server) status(request *restful.Request, response *restful.Response) {
	if !isAuthenticated(request, response) {
		return
	}

	si := sysinfo.Get()
	_ = response.WriteAsJson(si)
}

type instanceDetails struct {
	Port      string `json:"port"`
	Challenge string `json:"challenge"`
}

func (s *Server) user(request *restful.Request, response *restful.Response) {
	if !isAuthenticated(request, response) {
		return
	}

	instance := s.Instances.GetCurrentUserInstance(request.PathParameter("user_id"))
	if instance.Users == nil {
		_ = response.WriteErrorString(http.StatusNotFound, "User has no instance")
		return
	}
	_ = response.WriteAsJson(instanceDetails{Port: instance.Port, Challenge: instance.Challenge.Name})
}

func (s *Server) disconnect(request *restful.Request, response *restful.Response) {
	if !isAuthenticated(request, response) {
		return
	}

	s.Instances.Disconnect(request.PathParameter("user_id"))
}

func (s *Server) reset(request *restful.Request, response *restful.Response) {
	if !isAuthenticated(request, response) {
		return
	}

	user := request.PathParameter("user_id")
	instance := s.Instances.GetCurrentUserInstance(request.PathParameter("user_id"))

	s.Instances.AvoidInstance(user, instance)
	s.Instances.Disconnect(user)
	i, err := s.Instances.GetInstanceForUser(user, instance.Challenge)
	if err != nil {
		return
	}

	_ = response.WriteAsJson(instanceDetails{Port: i.Port, Challenge: i.Challenge.Name})
}

type instanceRequest struct {
	User      string `json:"user"`
	Challenge string `json:"challenge"`
}

func (s *Server) getInstance(request *restful.Request, response *restful.Response) {
	if !isAuthenticated(request, response) {
		return
	}

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
