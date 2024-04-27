package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func MessageRoute(route *gin.Engine) {
	route.POST("/create-message", controller.CreateMessage())
	route.GET("/get-message-by-id/:id", controller.GetMessageById())
	route.GET("/get-message-by-project/:id", controller.GetMessageByProject())
}
