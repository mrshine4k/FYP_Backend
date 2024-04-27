package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func AuthRoute(route *gin.Engine) {
	//User authentication
	route.POST("/login", controller.LoginHandler())
	route.GET("/logout", controller.LogOutHandler())
	route.GET("/isAuthorized", controller.IsAuthorized())
	//route.GET("/get-my-role-name", controller.GetMyRoleName())

	//Authorization
	route.POST("/authorization-add", controller.AuthorizationAdd())
	route.GET("/authorization-get-all", controller.AuthorizationGetAll())
	route.DELETE("/authorization/:id", controller.AuthorizationDelete())
	route.GET("/authorization/:id", controller.GetAuthorizationById())
	route.PUT("/authorization/:id", controller.UpdateAuthorization())
}
