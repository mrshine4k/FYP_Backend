package controller

import (
	"backend/model"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instace of the message model
		var message model.Message

		// Bind the request body to the message model
		bindingErr := c.BindJSON(&message)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Set the Id and timestamps for the message
		message.Id = primitive.NewObjectID()
		message.CreatedAt = time.Now().Unix()
		message.UpdatedAt = time.Now().Unix()

		// Insert the specified message to DB
		result, insertErr := messageCollection.InsertOne(ctx, message)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error inserting message: "+insertErr.Error())
			return
		}

		// Send response to client
		if result.InsertedID != nil {
			c.JSON(http.StatusCreated, gin.H{
				"success":   true,
				"messageId": message.Id,
			})
		}
	}
}

func GetMessageById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Message not found",
			})
			return
		}

		// Create an array for the employees
		var message gin.H

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "_id", Value: queryId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "sender"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "sender"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "sender.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "senderUserInfor"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "projects"},
					{Key: "localField", Value: "project"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "project"},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "createdAt", Value: 1},
					{Key: "files", Value: 1},
					{Key: "initialFiles", Value: 1},
					{Key: "message", Value: 1},
					{Key: "project", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$project", 0}},
					}},
					{Key: "sender", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$sender", 0}},
					}},
					{Key: "senderUserInfor", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$senderUserInfor", 0}},
					}},
				}},
			},
		}

		// Use the defined stages to aggregate data from User_Infor collection
		result, aggregateErr := messageCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating message: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		if result.Next(ctx) {
			decodeErr := result.Decode(&message)
			if decodeErr != nil {
				c.JSON(http.StatusInternalServerError, "Error decoding message: "+decodeErr.Error())
				return
			}
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": message,
		})
	}
}

func GetMessageByProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Message not found",
			})
			return
		}

		// Create an array for the employees
		var message []gin.H

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "project", Value: queryId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "sender"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "sender"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "sender.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "senderUserInfor"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "projects"},
					{Key: "localField", Value: "project"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "project"},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "createdAt", Value: 1},
					{Key: "files", Value: 1},
					{Key: "initialFiles", Value: 1},
					{Key: "message", Value: 1},
					{Key: "project", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$project", 0}},
					}},
					{Key: "sender", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$sender", 0}},
					}},
					{Key: "senderUserInfor", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$senderUserInfor", 0}},
					}},
				}},
			},
		}

		// Use the defined stages to aggregate data from User_Infor collection
		result, aggregateErr := messageCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating message: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &message)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding message: "+decodeErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": message,
		})
	}
}
