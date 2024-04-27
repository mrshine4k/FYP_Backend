package routes

import (
	"backend/controller"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func UserInforRoute(route *gin.Engine) {
	route.GET("/profile", func(c *gin.Context) {
		var authorized = c.MustGet("authorized")

		if authorized == false {
			c.JSON(http.StatusNonAuthoritativeInfo, gin.H{
				"message": "You can't access this page with your authorization",
			})
			return
		}

		// Access user information from the context
		user := c.MustGet("account_logged").(jwt.MapClaims)
		username := user["username"].(string)
		password := user["password"].(string)
		account_id := user["account_id"].(string)
		c.JSON(http.StatusOK, gin.H{
			"account_id": account_id,
			"username":   username,
			"password":   password,
		})

	})
	route.GET("/get-all-userinfor", controller.UserInforGetAll())
	route.GET("/get-one-userinfor/:id", controller.GetUserInforByID())
	//route.POST("/user-infor-add", controller.AddUserInfor())
	//route.POST("/add-user-infor-just-create", controller.AddUserInforJustCreate())
	route.PUT("/user-infor-update/:id", controller.UpdateUserInfor())
	route.PUT("/update-profile-image/:id", controller.UpdateProfileImage())
}
