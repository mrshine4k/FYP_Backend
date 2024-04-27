package middleware

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CookieAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// If user is trying to login, skip the middleware
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}

		// Get the token from cookie in the request header
		encryptedToken, cookieErr := c.Cookie("access_token")
		if cookieErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Decode the token from base64
		decodedToken, decodingErr := base64.StdEncoding.DecodeString(encryptedToken)
		if decodingErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Decrypt the token with encryption key
		decryptedToken, decryptErr := Decrypt([]byte(decodedToken), []byte(os.Getenv("ENCRYPTION_KEY")))
		if decryptErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decrypting token: "+decryptErr.Error())
			c.Abort()
			return
		}

		// Verify the signature of the token
		token, parsingErr := jwt.Parse(string(decryptedToken), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// If verification fails
		if parsingErr != nil {
			// Return unauthorized status
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Get the payload in the token claims
		claims, ok := token.Claims.(jwt.MapClaims)

		// Return unauthorized status if token is invalid
		if !ok && !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Check if the token has expired
		tokenExpireTime := time.Unix(int64(claims["exp"].(float64)), 0)
		if tokenExpireTime.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		//! Check if account level in database is the same as the level in the token
		// Convert the account ID in token to ObjectID
		accountId, convertErr := primitive.ObjectIDFromHex(claims["account_id"].(string))
		if convertErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Use the ObjectID to get all information about the currently logged in account
		var account gin.H
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "account_id", Value: accountId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "accounts"},
					{Key: "localField", Value: "account_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "account"},
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
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "authorizations"},
					{Key: "localField", Value: "account.account_authorization_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "authorization"},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "account_id", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$account._id", 0}},
					}},
					{Key: "userinfor_id", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor._id", 0}},
					}},
					{Key: "profile_image", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.profile_image", 0}},
					}},
					{Key: "fullname", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.fullname", 0}},
					}},
					{Key: "office", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.office", 0}},
					}},
					{Key: "department", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.department", 0}},
					}},
					{Key: "position", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.position", 0}},
					}},
					{Key: "gender", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor.gender", 0}},
					}},
					{Key: "state", Value: "$state"},
					{Key: "authorization", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$authorization.levelName", 0}},
					}},
					{Key: "username", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$account.username", 0}},
					}},
					{Key: "account_name", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$account.account_name", 0}},
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

		// Check if the account's role in database is the same as the role in the token
		if account == nil || account["authorization"].(string) != claims["level"].(string) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		// Set the current account in request context
		c.Set("currentAccount", account)
		c.Next()
	}
}
