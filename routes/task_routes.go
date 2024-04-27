package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func TaskRoute(route *gin.Engine) {
	route.POST("/task", controller.CreateTask())
	// route.GET("/task", controllers.GetTasks())
	// route.GET("/task/:id", controllers.GetTask())
	// route.PUT("/task/:id", controllers.UpdateTask())
	// route.DELETE("/task/:id", controllers.DeleteTask())
}
