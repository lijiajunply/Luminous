package router

import (
	"net/http"

	"luminous/internal/handler"
	"luminous/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	schoolHandler *handler.SchoolHandler,
	adminHandler *handler.AdminHandler,
	appHandler *handler.AppHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	if err := r.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/schools", schoolHandler.ListSchools)
		v1.GET("/schools/:code", schoolHandler.GetSchool)
		v1.GET("/App", appHandler.GetTagModel)
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/schools", adminHandler.AdminListSchools)
		admin.POST("/schools", adminHandler.CreateSchool)
		admin.PUT("/schools/:code", adminHandler.UpdateSchool)
		admin.DELETE("/schools/:code", adminHandler.DeleteSchool)
	}

	return r
}
