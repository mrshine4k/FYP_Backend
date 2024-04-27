package routes

import (
	"backend/controller"

	"github.com/gin-gonic/gin"
)

func EmployeeRoute(route *gin.Engine) {
	route.POST("/get-employee-id", controller.GetEmployeeIDByFullname())
	route.POST("/employee-update-state", controller.EmployeeUpdateState())
	route.POST("/search-employee-with-fullname", controller.EmployeeSearchFullName())
	route.GET("/employee-get-all", controller.EmployeeGetAll())
	route.GET("/get-employee-detail/:id", controller.GetEmployeeDetailWithID())
	route.GET("/get-employee-by-role/:role", controller.GetEmployeeByRole())
	route.GET("/get-employee-by-manager/:manager", controller.GetEmployeeByManager())

	route.POST("/create-employee", controller.CreateEmployee())
	route.PUT("/update-employee", controller.UpdateEmployee())
	route.GET("/get-employee-by-project/:project", controller.GetEmployeeByProject())
}
