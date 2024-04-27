package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func EpicRoute(route *gin.Engine) {
	route.POST("/epic", controller.CreateEpic())
	route.GET("/epics", controller.GetEpics())
	route.GET("/epic/:id", controller.GetEpicById())
	route.GET("/epic/search", controller.SearchEpic())
	route.GET("/epic-for-project/:id", controller.GetEpicForProject())
	route.PUT("/epic", controller.UpdateEpic())
	route.DELETE("/epic", controller.DeleteEpic())
	route.GET("/get-leader-for-epic/:id", controller.GetLeaderForEpic())
}
