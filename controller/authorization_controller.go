package controller

import (
	"backend/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type data_struct struct {
	LevelName   string `json:"levelName"`
	Description string `json:"description"`
}

func AuthorizationAdd() gin.HandlerFunc {
	return func(c *gin.Context) {
		Data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("request data error!!!")
		}
		// fmt.Printf("ctx.Request.body: %v", string(Data))

		var jsonData data_struct // map[string]interface{}

		// function attempts to unmarshal the JSON data in "Data" into the Go value pointed to by "jsonData"
		e := json.Unmarshal(Data, &jsonData)
		if e != nil {
			log.Printf("Error: %v \n", err)
		}
		fmt.Println("\nAccount Level Name: ", jsonData.LevelName)

		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		newAuthorization := model.Authorization{
			Id:          primitive.NewObjectID(),
			LevelName:   jsonData.LevelName,
			Description: jsonData.Description,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}

		var authorizationModel model.Authorization
		errFindLevelName := authorizationCollection.FindOne(
			ctx, bson.M{"levelName": jsonData.LevelName}).Decode(&authorizationModel)

		if errFindLevelName != nil {
			//insert the album into the database
			result, insertErr := authorizationCollection.InsertOne(ctx, newAuthorization)
			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, "Error inserting album"+insertErr.Error())
				return
			} else {
				fmt.Println("Insert Authorization Successful ID:", result)
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Authorization created",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Level name exists",
			})
		}

	}
}

// func GetMyRoleName() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		fmt.Println("My role name: ", global_role_name)
// 		c.JSON(http.StatusOK, gin.H{
// 			"roleName": global_role_name,
// 		})
// 	}
// }

func AuthorizationGetAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		var authorizations []model.Authorization
		defer cancel()

		//find all authorization
		results, err := authorizationCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		//reading from the db more optimally
		defer results.Close(ctx)

		//loop through the results
		for results.Next(ctx) {
			var singleAuthization model.Authorization
			if err := results.Decode(&singleAuthization); err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
			}

			authorizations = append(authorizations, singleAuthization)
		}

		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"authorizations": authorizations,
		})
	}
}

/*
Get the Authorizatoin by specified ID

params: None

return: gin.HandlerFunc Handler function to get the Authorization by ID
*/
func GetAuthorizationById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instance of the Authorization model
		var authorization model.Authorization

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Authorization not found",
			})
			return
		}

		queryErr := authorizationCollection.FindOne(ctx, bson.M{"_id": queryId}).Decode(&authorization)
		if queryErr != nil {
			c.JSON(http.StatusInternalServerError, queryErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"authorization": authorization,
		})
	}
}

/*
Update an Authorization by ID

params: None

return: gin.HandlerFunc Handler function to update an Authorization by ID
*/
func UpdateAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instance of the Authorization model
		var authorization model.Authorization

		// Convert the hex string to ObjectID
		updateId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Authorization not found",
			})
			return
		}

		// Bind the request body to the Project model
		bindingErr := c.BindJSON(&authorization)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid authorization",
			})
			return
		}

		// Temp variable to decode the FindOne result
		var tempResult bson.M
		// Validate the ID existence in DB
		decodeErr := authorizationCollection.FindOne(ctx, bson.M{"_id": updateId}).Decode(&tempResult)
		if decodeErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Authorization not found",
			})
			return
		}

		// Update the fields of the authorization in DB
		update := bson.M{
			"$set": bson.M{
				"levelName":   authorization.LevelName,
				"description": authorization.Description,
				"updatedAt":   time.Now().Unix(),
			},
		}

		// Find and update the authorization in DB
		result := authorizationCollection.FindOneAndUpdate(ctx, bson.M{"_id": updateId}, update)
		if result.Err() != nil {
			c.JSON(http.StatusInternalServerError, "Error updating authorization: "+result.Err().Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Authorization updated",
		})
	}
}

/*
Delete an Authorization by ID

params: None

return: gin.HandlerFunc Handler function to delete an Authorization by ID
*/
func AuthorizationDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		deleteId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Authorization not found",
			})
			return
		}

		// Delete the specified authorization from DB
		result, deleteErr := authorizationCollection.DeleteOne(ctx, bson.M{"_id": deleteId})
		if deleteErr != nil {
			c.JSON(http.StatusInternalServerError, "Error deleting authorization: "+deleteErr.Error())
			return
		}

		if result.DeletedCount == 1 {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Authorization deleted",
			})
		}
	}
}
