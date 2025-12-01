package sbi

import (
	"github.com/gin-gonic/gin"
)

type Route struct {
	Method  string
	Pattern string
	APIFunc gin.HandlerFunc
}

func applyRoutes(group *gin.RouterGroup, endpoints []Route) {
	for _, endpoint := range endpoints {
		switch endpoint.Method {
		case "GET":
			group.GET(endpoint.Pattern, endpoint.APIFunc)
		case "POST":
			group.POST(endpoint.Pattern, endpoint.APIFunc)
		case "PUT":
			group.PUT(endpoint.Pattern, endpoint.APIFunc)
		case "PATCH":
			group.PATCH(endpoint.Pattern, endpoint.APIFunc)
		case "DELETE":
			group.DELETE(endpoint.Pattern, endpoint.APIFunc)
		}
	}
}
