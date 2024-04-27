/*
Controller for handling data with Task model in DB

1. CreateTask: Create one or many Tasks

2. GetTasks: Get all Tasks from the DB

3. GetTaskById: Get the Task by specified ID

4. SearchTask: Search for Tasks by title

5. UpdateTask: Update one or many Tasks by specified ID(s)

6. DeleteTask: Delete one or many Tasks by specified ID(s)

7. ConvertTasksToInterface: Convert the tasks array to an interface array
*/
package controller

import (
	"backend/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Create one or many Tasks

params: None

return: gin.HandlerFunc Handler function to create one or many tasks
*/
func CreateTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Register the custom required validation function
		_ = validate.RegisterValidation("customrequired", CheckCustomRequired)

		// Create an array of the Task model
		var tasks model.Task

		// Bind the request body to the task model
		bindingErr := c.BindJSON(&tasks)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		fmt.Println("tasks:", tasks)

		// Check if validation failed for any task in the array
		var validationErrFlg = false
		// Validation result array
		var validationErrResult []gin.H

		// If validation failed for any task in the array
		if validationErrFlg {
			// Return the validation errors to the client
			c.JSON(http.StatusBadRequest, gin.H{
				"validationError": validationErrResult,
			})
			return
		}

		// Insert the specified document to DB
		result, insertErr := taskCollection.InsertOne(ctx, tasks)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error inserting task: "+insertErr.Error())
			return
		}

		// Send response to client
		if result.InsertedID != primitive.NilObjectID {
			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Task created",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Error creating task. Please try again later.",
			})
		}
	}
}

/*
Get all Task from the DB

params: None

return: gin.HandlerFunc Handler function to get all tasks
*/
func GetTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Task model
		//var tasks []model.Task // Use for the FindOne population method
		var tasks []gin.H

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "epics"},
					{Key: "localField", Value: "epic"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "epic"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "manager"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "manager"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "manager.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
		}

		// Use the $lookup stage to aggregate data from the Epics and Employee collections
		result, aggregateErr := taskCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating tasks: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the tasks array
		decodeErr := result.All(ctx, &tasks)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding tasks: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"count": len(tasks),
			"tasks": tasks,
		})
	}
}

/*
Get the Task by specified ID

params: None

return: gin.HandlerFunc Handler function to get the task by ID
*/
func GetTaskById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instance of the Task model
		var task []gin.H

		// Convert the hex string to an ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid task ID: "+convertErr.Error())
			return
		}

		// Define pipeline to filter the data by ID and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "_id", Value: queryId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "epics"},
					{Key: "localField", Value: "epic"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "epic"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "manager"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "manager"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "manager.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Epics and Employee collections
		result, aggregateErr := taskCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating task: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the task array
		decodeErr := result.All(ctx, &task)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding task: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"task": task[0],
		})
	}
}

/*
Search for Tasks by title

params: None

return: gin.HandlerFunc Handler function to search for tasks
*/
func SearchTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Task model
		var tasks []gin.H

		// Get the search data from request query
		query := c.Query("q")

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "title", Value: bson.D{
						{Key: "$regex", Value: query},
						{Key: "$options", Value: "i"},
					}},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "epics"},
					{Key: "localField", Value: "epic"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "epic"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "manager"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "manager"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "manager.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Epics and Employee collections
		result, aggregateErr := taskCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating task: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the tasks array
		decodeErr := result.All(ctx, &tasks)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding tasks: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"count": len(tasks),
			"tasks": tasks,
		})
	}
}

/*
Update one or many Tasks by specified ID(s)

params: None

return: gin.HandlerFunc Handler function to update one or many tasks
*/
func UpdateTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Register the custom required validation function
		_ = validate.RegisterValidation("customrequired", DummyValidation)

		// Create an array of the Task
		var tasks []model.Task

		// Bind the request body to the task model
		bindingErr := c.BindJSON(&tasks)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Check if validation failed for any task in the array
		var validationErrFlg = false
		// The result of the validation process, will be used to return the validation error to the client
		var validationErrResult []gin.H
		// Process the specified tasks
		for i, task := range tasks {
			// Check if validation failed for the current task
			var taskErr = false
			// The result of the validation process for the current task
			var singleValidationErr gin.H

			// Check if the Id is empty or not
			if task.Id.IsZero() || task.Id.Hex() == "" {
				// Create a single error object for the current task
				singleValidationErr = gin.H{
					"element": i + 1,
					"error":   []gin.H{},
				}

				// Add the field and tag to the error array
				singleValidationErr["error"] = append(singleValidationErr["error"].([]gin.H), gin.H{
					"field": "Id",
					"tag":   "required",
				})

				// Set the task error flag to true
				taskErr = true
			}

			// Temp variable to decode the FindOne result
			var result bson.M
			// Validate the ID existence in DB
			decodeErr := taskCollection.FindOne(ctx, bson.M{"_id": task.Id}).Decode(&result)
			if decodeErr != nil {
				if !taskErr {
					singleValidationErr = gin.H{
						"element": i + 1,
						"error":   []gin.H{},
					}
				}

				// Add the field and tag to the error array
				singleValidationErr["error"] = append(singleValidationErr["error"].([]gin.H), gin.H{
					"field": "Id",
					"tag":   "not found",
				})

				// Set the task error flag to true
				taskErr = true
			}

			// Validate each task
			validationErr := validate.Struct(task)
			if validationErr != nil {
				if !taskErr {
					singleValidationErr = gin.H{
						"element": i + 1,
						"error":   []gin.H{},
					}
				}

				// Add the field and tag to the error array
				for _, ve := range validationErr.(validator.ValidationErrors) {
					singleValidationErr["error"] = append(singleValidationErr["error"].([]gin.H), gin.H{
						"field": ve.Field(),
						"tag":   ve.Tag(),
					})
				}
			}

			// If validation failed for the current task
			if singleValidationErr != nil {
				// Add the single validation error to the validation error result array
				validationErrResult = append(validationErrResult, singleValidationErr)

				// Set validation error flag to true to return the error to client later
				validationErrFlg = true
			}
		}

		// If validation failed for any task in the array
		if validationErrFlg {
			// Return the validation error to the client
			c.JSON(http.StatusBadRequest, gin.H{
				"validationError": validationErrResult,
			})
			return
		}

		// Check the length of tasks array to update appropriately
		if len(tasks) == 1 {
			// Update the fields of an task in DB
			update := bson.M{
				"$set": bson.M{
					"title":       tasks[0].Title,
					"description": tasks[0].Description,
					"updatedAt":   time.Now(),
				},
			}

			// Find and update the task in DB
			result := taskCollection.FindOneAndUpdate(ctx, bson.M{"_id": tasks[0].Id}, update)
			if result.Err() != nil {
				c.JSON(http.StatusInternalServerError, "Error updating task: "+result.Err().Error())
				return
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": "1 task updated",
			})
		} else if len(tasks) > 1 {
			// Count the number of documents updated
			var modifyCount int

			for i, task := range tasks {
				// Update the fields of each task in DB
				update := bson.M{
					"$set": bson.M{
						"title":       task.Title,
						"description": task.Description,
						"updatedAt":   time.Now(),
					},
				}

				// Find and update each task in DB
				result := taskCollection.FindOneAndUpdate(ctx, bson.M{"_id": task.Id}, update)
				if result.Err() != nil {
					c.JSON(http.StatusInternalServerError, "Error updating task "+strconv.Itoa(i+1)+": "+result.Err().Error())
					return
				}

				// After each successful update, increment the modifyCount
				modifyCount++
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": strconv.Itoa(modifyCount) + " tasks updated",
			})
		} else {
			// If the tasks array is empty return an error
			c.JSON(http.StatusBadRequest, "Empty JSON array")
		}
	}
}

/*
Delete one or many Tasks by specified ID(s)

params: None

return: gin.HandlerFunc Handler function to delete one or many tasks
*/
func DeleteTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Read the raw body from the request
		body, readingErr := io.ReadAll(c.Request.Body)
		if readingErr != nil {
			c.JSON(http.StatusInternalServerError, "Error reading request body: "+readingErr.Error())
			return
		}

		// Unmarshal the raw body into a slice of string
		var objectIdStrings []string
		unmarshalErr := json.Unmarshal(body, &objectIdStrings)
		if unmarshalErr != nil {
			c.JSON(http.StatusBadRequest, "Error unmarshalling request body: "+unmarshalErr.Error())
			return
		}

		// Convert each element in the slice to ObjectID
		var deleteArr []primitive.ObjectID
		for _, objectIdString := range objectIdStrings {
			id, convertErr := primitive.ObjectIDFromHex(objectIdString)
			if convertErr != nil {
				c.JSON(http.StatusBadRequest, "Invalid task ID: "+convertErr.Error())
				return
			}
			deleteArr = append(deleteArr, id)
		}

		// Check the length of the delete array to delete appropriately
		if len(deleteArr) == 1 {
			// Delete the specified document from DB
			result, deleteErr := taskCollection.DeleteOne(ctx, bson.M{"_id": deleteArr[0]})
			if deleteErr != nil {
				c.JSON(http.StatusInternalServerError, "Error deleting task: "+deleteErr.Error())
				return
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": strconv.Itoa(int(result.DeletedCount)) + " task deleted",
			})
		} else if len(deleteArr) > 1 {
			// Delete all specified documents from DB
			result, deleteErr := taskCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": deleteArr}})
			if deleteErr != nil {
				c.JSON(http.StatusInternalServerError, "Error deleting task: "+deleteErr.Error())
				return
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": strconv.Itoa(int(result.DeletedCount)) + " tasks deleted",
			})
		} else {
			// If the tasks array is empty return an error
			c.JSON(http.StatusBadRequest, "Empty JSON array")
		}
	}
}

/*
Convert the tasks array to an interface array

params: tasks []model.Task Array of tasks to convert to interface array

return: []interface{} The result array of the conversion
*/
func ConvertTasksToInterface(tasks []model.Task) []interface{} {
	// Create an interface array with the length of the tasks array
	taskInterface := make([]interface{}, len(tasks))

	// Bring each Task to the interface array
	for i, task := range tasks {
		taskInterface[i] = task
	}

	// Return the interface array
	return taskInterface
}
