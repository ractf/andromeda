package routes

import (
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

func Authenticated(function restful.RouteFunction, apiKey string) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		if request.HeaderParameter("Authorization") != apiKey {
			_ = response.WriteErrorString(http.StatusForbidden, "Incorrect api key")
			return
		}
		function(request, response)
	}
}
