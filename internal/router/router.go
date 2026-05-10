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
	schoolServices ...handler.SchoolServiceHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.MetricsMiddleware())

	// Trust all proxies so that X-Forwarded-* headers work correctly
	// behind reverse proxies (nginx, traefik, etc.) in Docker deployments.
	if err := r.SetTrustedProxies(nil); err != nil {
		panic(err)
	}
	r.GET("/metrics", middleware.MetricsHandler())
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	v1 := r.Group("/api/v1")
	{
		v1.GET("/schools", schoolHandler.ListSchools)
		v1.GET("/schools/:code", schoolHandler.GetSchool)
		v1.GET("/App", appHandler.GetTag)
		v1.GET("/App/GetTag", appHandler.GetTagModel)

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
