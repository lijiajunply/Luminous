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
	schoolServices ...handler.SchoolServiceHandler,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.GET("/metrics", middleware.MetricsHandler())
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	v1 := r.Group("/api/v1")
	{
		v1.GET("/schools", schoolHandler.ListSchools)
		v1.GET("/schools/:code", schoolHandler.GetSchool)

		for _, svc := range schoolServices {
			svc.RegisterRoutes(v1.Group("/schools/" + svc.Code()))
		}
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
