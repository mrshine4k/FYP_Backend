package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func MainRoute(route *gin.Engine) {
	//route.GET("/home-employee", controller.HomeController())
	//route.GET("/home-manager", controller.HomeController())
	//route.POST("/home", controller.TestController())
	route.POST("/upload-image", controller.UploadFile())
}
