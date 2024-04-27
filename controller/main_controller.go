package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// imagePath := "./uploads/"

		fmt.Println("test")
		file, err := c.FormFile("image")
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{

				"message": "Bad request",
			})
			// return
		}
		var filePath string
		fmt.Println("File name", file.Filename)

		profile_image := uuid.New()
		// Define the desired path where you want to save the file
		filePath = "../frontend/src/lib/images/profile/" + profile_image.String() + ".jpg"

		// Create a new file at the desired path
		newFile, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error create new image": err.Error()})
			return
		}
		defer newFile.Close()

		// Save the uploaded file to the newly created file
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error save upload image": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "path": filePath})
		// defer file.Close()

		// // Tạo tệp mới trong thư mục uploads
		// out, err := os.Create(imagePath + file)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{

		// 		"message": "Internal server error",
		// 	})
		// 	return
		// }
		// defer out.Close()

		// // Copy nội dung của file ảnh vào tệp mới
		// _, err = io.Copy(out, file)
		// if err != nil {

		// 	c.JSON(http.StatusInternalServerError, gin.H{

		// 		"message": "Internal server error",
		// 	})
		// 	return
		// }

	}
}

// func HomeController() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if !global_authorization {
// 			c.JSON(http.StatusNonAuthoritativeInfo, gin.H{
// 				"authorized": global_authorization,
// 				"message":    "You can't access this page with your authorization",
// 			})
// 			return
// 		}

// 		// account := c.MustGet("account_logged").(jwt.MapClaims)

// 		// username := account["username"].(string)

// 		// c.JSON(http.StatusOK, gin.H{
// 		// 	"username": username,
// 		// 	// "username": username,
// 		// })

// 		var cookie_token = global_account_token
// 		if cookie_token == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		token, err := jwt.Parse(cookie_token, func(token *jwt.Token) (interface{}, error) {
// 			return []byte("secret_key"), nil
// 		})
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 			c.Abort()
// 			return
// 		}

// 		//Get username
// 		var username string
// 		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 			// Access user information from claims
// 			username = claims["username"].(string)
// 		} else {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
// 			c.Abort()
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"username":   username,
// 			"authorized": global_authorization,
// 			// "account": account,
// 		})
// 	}
// }

// type login_token_struct struct {
// 	LoginToken string `json:"login_token"`
// }

// func TestController() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		// account := c.MustGet("account_logged").(jwt.MapClaims)
// 		// username := account["username"].(string)

// 		// // Data, err := io.ReadAll(c.Request.Body)
// 		// // if err != nil {
// 		// // 	fmt.Println("request data error!!!")
// 		// // }
// 		// // // fmt.Printf("ctx.Request.body: %v", string(Data))

// 		// // var jsonData login_token_struct // map[string]interface{}
// 		// // e := json.Unmarshal(Data, &jsonData)
// 		// // if e != nil {
// 		// // 	log.Printf("Error: %v \n", err)
// 		// // }

// 		// c.JSON(http.StatusOK, gin.H{
// 		// 	// "login_loken": jsonData.LoginToken,
// 		// 	"username": username,
// 		// })
// 		store := cookie.NewStore([]byte("your-secret-key"))
// 		sessions.Sessions("mysession", store)
// 		session := sessions.Default(c)
// 		session.Set("account_infor", "test")
// 		session.Save()
// 		accountinfor := session.Get("account_infor")
// 		fmt.Println("account infor:", accountinfor)
// 	}
// }
