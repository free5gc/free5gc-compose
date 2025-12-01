package sbi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getOamRoutes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/",
			APIFunc: s.apiGetOamIndex,
		},
	}
}

func (s *Server) apiGetOamIndex(gc *gin.Context) {
	s.Processor().GetOamIndex(gc)
}
