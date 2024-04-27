package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func AccountRoute(route *gin.Engine) {
	route.GET("/accounts-get-all", controller.AccountGetAll())
	//route.GET("/get-account-to-update/:id", controller.GetAccountToUpdate())
	//route.GET("/my-account", controller.MyAccount())
	route.GET("/account-delete-all", controller.AccountDeleteAll())

	route.DELETE("/account-delete-one/:id", controller.AccountDeleteOne())
	route.POST("/account-update/:id", controller.AccountUpdate())

	route.POST("/account-add", controller.AccountAdd())
	route.GET("/reset-password/:id", controller.ResetPassword())
	route.POST("/change-password/:id", controller.ChangePassword())
}
