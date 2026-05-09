package handler

import "github.com/gin-gonic/gin"

// SchoolServiceHandler is implemented by each school integration package.
// RegisterRoutes receives a router group scoped to /api/v1/schools/{code}/.
type SchoolServiceHandler interface {
	Code() string
	RegisterRoutes(rg *gin.RouterGroup)
}
