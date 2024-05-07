package controller

import (
	"backend/model"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instance of the Account model to store username and password
		var loginCredentials model.Account

		// Bind the username and password to the Account model
		bindingErr := c.BindJSON(&loginCredentials)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Trim the received username and password
		loginCredentials.Username = strings.TrimSpace(loginCredentials.Username)
		loginCredentials.Password = strings.TrimSpace(loginCredentials.Password)

		// Find account by username and join with authorizations collection
		var account gin.H
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "username", Value: loginCredentials.Username},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "authorizations"},
					{Key: "localField", Value: "account_authorization_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "levelName"},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "username", Value: 1},
					{Key: "password", Value: 1},
					{Key: "account_name", Value: 1},
					{Key: "levelName", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$levelName.levelName", 0}},
					}},
				}},
			},
		}
		usernameQueryResult, usernameQueryErr := accountCollection.Aggregate(ctx, pipeline)
		if usernameQueryErr != nil {
			c.JSON(http.StatusInternalServerError, "Error querying username: "+usernameQueryErr.Error())
			return
		}
		if usernameQueryResult.Next(ctx) {
			decodeErr := usernameQueryResult.Decode(&account)
			if decodeErr != nil {
				c.JSON(http.StatusInternalServerError, "Error decoding username: "+decodeErr.Error())
				return
			}
		}

		// If username does not exist
		if account == nil {
			// Send response to the client
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User account does not exist",
				"field":   "email",
				"success": false,
			})
		} else {
			// Compare entered password with hashed password in database
			comparePasswordSuccess := VerifyPassword(loginCredentials.Password, account["password"].(string))

			// If passwords match
			if comparePasswordSuccess {
				// Create token with user ID and level name in payload
				token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
					"account_id": account["_id"].(primitive.ObjectID).Hex(),
					"level":      account["levelName"].(string),
					"exp":        time.Now().Add(time.Hour * 24).Unix(),
				})

				// Create token with secret key
				signedToken, signingErr := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
				if signingErr != nil {
					c.JSON(http.StatusInternalServerError, "Error signing token: "+signingErr.Error())
					return
				}

				// Encrypt the token with encryption key
				var encryptedToken, encryptErr = Encrypt([]byte(signedToken), []byte(os.Getenv("ENCRYPTION_KEY")))
				if encryptErr != nil {
					c.JSON(http.StatusInternalServerError, "Error encrypting token: "+encryptErr.Error())
					return
				}

				// Encode the encrypted token to base64 and to URL safe format
				encodedBase64 := base64.StdEncoding.EncodeToString([]byte(encryptedToken))
				encodedToken := url.QueryEscape(encodedBase64)

				// Create an HTTP-only cookie
				cookie := &http.Cookie{
					Name:     "access_token",
					Value:    encodedToken,
					Expires:  time.Now().Add(24 * time.Hour), // Adjust the expiration time as needed
					HttpOnly: true,
					Secure:   false,                  // Set to true to ensure the cookie is sent only over HTTPS
					SameSite: http.SameSiteLaxMode,   // Set the SameSite attribute for CSRF protection
					Path:     "/",                    // Set the path attribute to restrict the cookie to a specific path
					Domain:   "trietandfriends.site", // Set the domain attribute to restrict the cookie to a specific domain
					//removed cause unecessary..?
				}

				// Send the cookie to client
				http.SetCookie(c.Writer, cookie)

				// Send the response to client
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Login successful",
					"level":   account["levelName"].(string),
				})
				return
			} else {
				// Send response to the client for incrorrect username or password
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Incorrect password",
					"field":   "password",
				})
			}
		}
	}
}

func LogOutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create an HTTP-only cookie
		cookie := &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour), // Adjust the expiration time as needed
			HttpOnly: true,
			Secure:   false,                   // Set to true to ensure the cookie is sent only over HTTPS
			SameSite: http.SameSiteStrictMode, // Set the SameSite attribute for CSRF protection
			Path:     "/",                     // Set the path attribute to restrict the cookie to a specific path
			Domain:   "",                      // Set the domain attribute to restrict the cookie to a specific domain
		}

		// Send the cookie to client
		http.SetCookie(c.Writer, cookie)

		// Send the response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logout successful",
		})
	}
}

func AccountAdd() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instace of the Account model
		var account model.Account

		// Bind the request body to the project model
		bindingErr := c.BindJSON(&account)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Check if the authorization level is valid
		var authorization model.Authorization
		authorizationQueryErr := authorizationCollection.FindOne(ctx, bson.M{"_id": account.Account_Authorization_Id}).Decode(&authorization)
		if authorizationQueryErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid authorization level",
			})
			return
		}

		// Trim the received username, password, and account name
		account.Username = strings.TrimSpace(account.Username)
		account.Password = strings.TrimSpace(account.Password)
		account.Account_Name = strings.TrimSpace(account.Account_Name)

		// Check if the username already exists in database
		var existingAccount model.Account
		accountQueryErr := accountCollection.FindOne(ctx, bson.M{"username": account.Username}).Decode(&existingAccount)
		// If username does not exist
		if accountQueryErr != nil {
			// Hash the password
			hashedPassword, hashingErr := bcrypt.GenerateFromPassword([]byte(account.Password), 14)
			if hashingErr != nil {
				c.JSON(http.StatusInternalServerError, "Error hashing password: "+hashingErr.Error())
				return
			}

			// Add Id, password and timestamp for the account object
			account.Id = primitive.NewObjectID()
			account.Password = string(hashedPassword)
			account.CreatedAt = time.Now().Unix()
			account.UpdatedAt = time.Now().Unix()

			// Insert the account into database
			accountInsertResult, accountInsertErr := accountCollection.InsertOne(ctx, account)
			if accountInsertErr != nil {
				c.JSON(http.StatusInternalServerError, "Error inserting account: "+accountInsertErr.Error())
				return
			} else {
				fmt.Println("Insert account Successful into Database with Account ID:", accountInsertResult.InsertedID)

				// Create default user information object
				userInfor := model.UserInfor{
					Id:            primitive.NewObjectID(),
					Profile_Image: "profile_image_default.jpg",
				}

				// Insert the default user information values into database
				userInforInsertResult, userInforInsertErr := userInforCollection.InsertOne(ctx, userInfor)
				if userInforInsertErr != nil {
					// If error during inserting user information, delete the account just created
					accountDeleteResult, accountDeleteErr := accountCollection.DeleteOne(ctx, bson.M{"_id": accountInsertResult.InsertedID})
					if accountDeleteErr != nil {
						c.JSON(http.StatusInternalServerError, "Error deleting account: "+accountDeleteErr.Error())
						return
					} else {
						fmt.Println("Delete account success: ", accountDeleteResult)
					}

					// Send response to the client
					c.JSON(http.StatusInternalServerError, "Error inserting user information: "+userInforInsertErr.Error())
					return
				} else {
					// Create the employee object
					employee := model.Employee{
						Id:          primitive.NewObjectID(),
						State:       0,
						AccountID:   account.Id,
						UserInforId: userInfor.Id,
						CreatedAt:   time.Now().Unix(),
						UpdatedAt:   time.Now().Unix(),
					}

					// Insert the employee object into database
					employeeInsertResult, employeeInsertErr := employeeCollection.InsertOne(ctx, employee)
					if employeeInsertErr != nil {
						// If error during inserting employee, delete the account and user information just created
						accountDeleteResult, accountDeleteErr := accountCollection.DeleteOne(ctx, bson.M{"_id": accountInsertResult.InsertedID})
						if accountDeleteErr != nil {
							c.JSON(http.StatusInternalServerError, "Error deleting account: "+accountDeleteErr.Error())
							return
						} else {
							fmt.Println("Delete account success: ", accountDeleteResult)
						}
						userInforDeleteResult, userInforDeleteErr := userInforCollection.DeleteOne(ctx, bson.M{"_id": userInforInsertResult.InsertedID})
						if userInforDeleteErr != nil {
							c.JSON(http.StatusInternalServerError, "Error deleting account: "+userInforDeleteErr.Error())
							return
						} else {
							fmt.Println("Delete account success: ", userInforDeleteResult)
						}

						// Send response to the client
						c.JSON(http.StatusInternalServerError, "Error inserting employee: "+employeeInsertErr.Error())
						return
					} else {
						fmt.Println("Inserted employee into Database with ID:", employeeInsertResult.InsertedID)
						c.JSON(http.StatusCreated, gin.H{
							"success":   true,
							"account":   accountInsertResult.InsertedID,
							"userInfor": userInforInsertResult.InsertedID,
							"employee":  employeeInsertResult.InsertedID,
							"message":   "Account created",
						})
					}
				}

			}
		} else {
			// Send repsonse to the client if the username already exists
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Username already exists",
			})
		}
	}
}

type accountWithRoleName_struct struct {
	Id                       primitive.ObjectID `bson:"_id,omitempty"`
	Username                 string             `bson:"username,omitempty"`
	Password                 string             `bson:"password"`
	Account_Name             string             `bson:"account_name"`
	Role_Name                string             `bson:"role_name"`
	Account_Authorization_Id primitive.ObjectID `bson:"account_authorization_id,omitempty"`
	CreatedAt                int64              `bson:"createdAt"`
	UpdatedAt                int64              `bson:"updatedAt"`
}

func AccountGetAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var accounts []accountWithRoleName_struct
		defer cancel()

		//find all authorization
		results, errCollection := accountCollection.Find(ctx, bson.M{})

		if errCollection != nil {
			c.JSON(http.StatusInternalServerError, errCollection.Error())
			return
		}

		//reading from the db more optimally
		defer results.Close(ctx)

		//loop through the results
		for results.Next(ctx) {
			var singleAccount accountWithRoleName_struct
			var singleAuthization model.Authorization

			// var accountWithRole account_struct
			// accountWithRole = account_struct{
			// 	RoleName: "admin",
			// 	Username: results.Username,
			// }
			if err := results.Decode(&singleAccount); err != nil {
				c.JSON(http.StatusInternalServerError, "decode single account"+err.Error())
				return
			}

			filter := bson.M{"_id": singleAccount.Account_Authorization_Id}

			//filter role name with corresponding authorization_id
			if err := authorizationCollection.FindOne(ctx, filter).Decode(&singleAuthization); err != nil {
				c.JSON(http.StatusInternalServerError, "Find authorization error "+err.Error())
				return
			}
			//add role name into singleAccount
			singleAccount.Role_Name = singleAuthization.LevelName

			accounts = append(accounts, singleAccount)
		}

		// //reading from the db more optimally
		// defer results.Close(ctx)

		// //loop through the results
		// for results.Next(ctx) {
		// 	var singleAccount model.Account
		// 	if err := results.Decode(&singleAccount); err != nil {
		// 		c.JSON(http.StatusInternalServerError, err.Error())
		// 	}

		// 	accounts = append(accounts, singleAccount)
		// }

		c.JSON(http.StatusOK, accounts)
	}
}

func AccountDeleteAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := accountCollection.Drop(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		} else {
			c.JSON(http.StatusOK, "Delete Account Collection  successful!!!")
			c.JSON(http.StatusCreated, gin.H{
				"state": "success",
			})
		}
	}
}

func AccountDeleteOne() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.JSON(http.StatusOK, gin.H{
		// 	"id": c.Param("id"),
		// })

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		opts := options.Delete().SetCollation(&options.Collation{
			Locale:    "en_US",
			Strength:  1,
			CaseLevel: false,
		})

		id, _ := primitive.ObjectIDFromHex(c.Param("id"))
		filter := bson.D{{Key: "_id", Value: id}}
		if DeleteAccountResult1, err1 := accountCollection.DeleteOne(ctx, filter, opts); err1 != nil {
			c.JSON(http.StatusInternalServerError, "Error deleting account"+err1.Error())
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "deleted account successfully",
			})
			fmt.Println("Delete account success: ", DeleteAccountResult1)
		}
	}
}

func AccountUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		//must have to connect database
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		var accountUpdate model.Account

		//get id from link
		var GetAccountID = c.Param("id")
		//convert to primary object for ID
		id, _ := primitive.ObjectIDFromHex(GetAccountID)

		var getAccountUpdate model.Account
		//validate the request body
		if err := c.BindJSON(&getAccountUpdate); err != nil {
			c.JSON(http.StatusBadRequest, "Request error"+err.Error())
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&getAccountUpdate); validationErr != nil {
			c.JSON(http.StatusBadRequest, "Error, missing field"+validationErr.Error())
			return
		}
		fmt.Println("Get Username to update:", getAccountUpdate.Username)
		fmt.Println("Get Password to update:", getAccountUpdate.Password)
		fmt.Println("Get Authorization to update:", getAccountUpdate.Account_Authorization_Id)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(getAccountUpdate.Password), bcrypt.DefaultCost)
		if err != nil {
			// Handle error
			return
		}
		filter := bson.D{{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "username", Value: getAccountUpdate.Username},
			{Key: "password", Value: string(hashedPassword)},
			{Key: "account_name", Value: getAccountUpdate.Account_Name},
			{Key: "account_authorization_id", Value: getAccountUpdate.Account_Authorization_Id},
		}}} // "password", "asdfadfafs",

		if err1 := accountCollection.FindOneAndUpdate(ctx, filter, update).Decode(&accountUpdate); err1 != nil {
			c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Updated account successfully",
			})
		}
	}
}

func IsAuthorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentAccount := c.MustGet("currentAccount").(gin.H)
		if currentAccount != nil && currentAccount["authorization"] != nil && currentAccount["authorization"].(string) != "" {
			c.JSON(http.StatusOK, gin.H{
				"success":     true,
				"message":     "Authorized",
				"currentUser": currentAccount,
			})
			return
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}
	}
}

func ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Get id from link
		var accountString = c.Param("id")
		//convert to primary object for ID
		accountId, convertErr := primitive.ObjectIDFromHex(accountString)
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Error converting account ID: "+convertErr.Error())
			return
		}

		var account gin.H
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "account_id", Value: accountId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "userinfor"},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "fullname", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.fullname", 0}},
					}},
					{Key: "email", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.email", 0}},
					}},
				}},
			},
		}
		accountQueryResult, aggregateErr := employeeCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			fmt.Print(aggregateErr.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}
		if accountQueryResult.Next(ctx) {
			decodeErr := accountQueryResult.Decode(&account)
			if decodeErr != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Unauthorized",
				})
				c.Abort()
				return
			}
		}

		var password, unhashedPassword = GenerateAndHashPassword()
		if password == "" {
			c.JSON(http.StatusInternalServerError, "Error generating password")
			return
		}

		accountCollection.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "_id", Value: accountId}},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "password", Value: password},
					{Key: "updatedAt", Value: time.Now().Unix()},
				}},
			},
		)

		sendEmailSuccess := SendEmailResetPassword(account["email"].(string), account["fullname"].(string), unhashedPassword)
		if !sendEmailSuccess {
			c.JSON(http.StatusInternalServerError, "Error sending emails")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Account password reset",
		})
	}
}

func ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Get id from link
		var accountString = c.Param("id")
		// Convert to primary object for ID
		accountId, convertErr := primitive.ObjectIDFromHex(accountString)
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Error converting account ID: "+convertErr.Error())
			return
		}

		var account model.Account

		accountQueryErr := accountCollection.FindOne(ctx, bson.M{"_id": accountId}).Decode(&account)
		if accountQueryErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid account Id",
			})
			return
		}

		var dataMap = make(map[string]interface{})
		var jsonData map[string]interface{}
		// Read all data from request
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Failed to read request body")
			return
		}
		// Unmarshal JSON data
		if err := json.Unmarshal(requestBody, &jsonData); err != nil {
			c.JSON(http.StatusBadRequest, "Failed to parse JSON")
			return
		}
		// Append data to map
		for key, value := range jsonData {
			dataMap[key] = value
		}

		var oldPW = dataMap["old_password"].(string)
		var newPW = dataMap["new_password"].(string)
		var renewPW = dataMap["renew_password"].(string)
		var hashedPasswordHex string
		var updatePasswordResult *mongo.SingleResult

		if (oldPW != "" && newPW != "" && renewPW != "") && (newPW == renewPW) {
			// Check if the old password is correct
			if VerifyPassword(oldPW, account.Password) {
				if CheckPassword(newPW) {
					combined := os.Getenv("PEPPER1") + newPW + os.Getenv("PEPPER2")

					// Hash the generated password
					hashedPassword, errhashedPassword := bcrypt.GenerateFromPassword([]byte(combined), 15)
					if errhashedPassword != nil {
						c.JSON(http.StatusInternalServerError, "Failed to hash password")
						return
					}

					hashedPasswordHex = hex.EncodeToString(hashedPassword)

					updatePasswordResult = accountCollection.FindOneAndUpdate(
						ctx,
						bson.D{{Key: "_id", Value: accountId}},
						bson.D{
							{Key: "$set", Value: bson.D{
								{Key: "password", Value: hashedPasswordHex},
								{Key: "updatedAt", Value: time.Now().Unix()},
							}},
						},
					)
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"message": "*New password does not meet requirements",
					})
					return
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "*Incorrect password",
				})
				return
			}
		}

		if updatePasswordResult.Err() != nil {
			c.JSON(http.StatusInternalServerError, "Failed to update password")
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Password updated",
			})
		}
	}
}
