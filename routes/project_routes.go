package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func ProjectRoute(route *gin.Engine) {
	route.POST("/project", controller.CreateProject())
	route.GET("/projects", controller.GetProjects())
	route.GET("/project/:id", controller.GetProjectById())
	route.GET("/project/search/:query", controller.SearchProject())
	route.PUT("/project/:id", controller.UpdateProject())
	route.DELETE("/project/:id", controller.DeleteProject())

	route.GET("/view-projects-for-manager/:id", controller.GetProjectsForManager())
}
