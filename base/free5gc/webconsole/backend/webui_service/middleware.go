package webui_service

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Path validation constants
const (
	MinURILength = 2
)

var PublicPath = "public"

func ReturnPublic() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		if method == "GET" {
			destPath := PublicPath + context.Request.RequestURI
			if destPath[len(destPath)-1] == '/' {
				destPath = destPath[:len(destPath)-1]
			}
			destPath = verifyDestPath(destPath)
			context.File(destPath)
		} else {
			context.Next()
		}
	}
}

func verifyDestPath(requestedURI string) string {
	protected_route := []string{
		"status",
		"analysis",
		"subscriber",
		"profile",
		"tenant",
		"charging",
		"login",
	}

	destPath := filepath.Clean(requestedURI)

	// if destPath contains ".." then it is not a valid path
	if strings.Contains(destPath, "..") {
		return PublicPath
	}

	// If it is in ProtectedRoute, we must return to root
	for _, r := range protected_route {
		uri := strings.Split(requestedURI, "/")

		if len(uri) < MinURILength {
			continue
		}
		if uri[1] == r {
			return PublicPath
		}
	}
	return destPath
}
