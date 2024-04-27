/*
Controller for handling data with Project model in DB

1. CreateProject: Create a Project

2. GetProjects: Get all Projects

3. GetProjectById: Get a Project by ID

4. SearchProject: Search for Projects by title

5. UpdateProject: Update a Project by ID

6. DeleteProject: Delete a Project by ID
*/
package controller

import (
	"backend/model"
	"context"
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
Create a Project

params: None

return: gin.HandlerFunc Handler function to create a project
*/
func CreateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Create an instace of the Project model
		var project model.Project

		// Bind the request body to the project model
		bindingErr := c.BindJSON(&project)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Process the specified project
		validationErr := validate.Struct(&project)
		if validationErr != nil {
			// Validation error array to store validation errors
			var projectValidationErr []gin.H

			// For each validation error
			for _, ve := range validationErr.(validator.ValidationErrors) {
				// Append the validation error to the projectValidationErr array
				projectValidationErr = append(projectValidationErr, gin.H{
					"field": ve.Field(),
					"tag":   ve.Tag(),
				})
			}

			// Return the validation errors to the client
			c.JSON(http.StatusBadRequest, gin.H{
				"success":         false,
				"validationError": projectValidationErr,
			})
			return
		}

		// Set the Id and timestamps for the project
		project.Id = primitive.NewObjectID()
		project.CreatedAt = time.Now()
		project.UpdatedAt = time.Now()

		// Insert the specified project to DB
		result, insertErr := projectCollection.InsertOne(ctx, project)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error inserting project: "+insertErr.Error())
			return
		}

		// Send response to client
		if result.InsertedID != nil {
			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Project created",
				"project": result.InsertedID,
			})
		}
	}
}

/*
Get all Projects from DB

params: None

return: gin.HandlerFunc Handler function to get all projects
*/
func GetProjects() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Project mdoel
		var projects []gin.H

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "leader"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "leader"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "leader.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "epics"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "project"},
					{Key: "as", Value: "epic"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "tasks"},
					{Key: "localField", Value: "epic._id"},
					{Key: "foreignField", Value: "epic"},
					{Key: "as", Value: "tasks"},
				}},
			},
		}

		// Use the $lookup stage to aggregate data from the Epic and Project collections
		result, aggregateErr := projectCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &projects)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"projects": projects,
		})
	}
}

/*
Get the Project by specified ID

params: None

return: gin.HandlerFunc Handler function to get a project by ID
*/
func GetProjectById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Project not found",
			})
			return
		}

		// Create an array for the employees
		var project []gin.H

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
					{Key: "localField", Value: "leader"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "leader"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "leader.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "leader_userinfor"},
				}},
			},
		}

		// Use the defined stages to aggregate data from User_Infor collection
		result, aggregateErr := projectCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating project: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &project)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding project: "+decodeErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"project": project[0],
		})
	}
}
func GetProjectsForManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Project not found",
			})
			return
		}

		// Create an array for the employees
		var project []gin.H

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "leader"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "leader"},
				}},
			},
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "leader._id", Value: queryId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "leader.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "epics"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "project"},
					{Key: "as", Value: "epic"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "tasks"},
					{Key: "localField", Value: "epic._id"},
					{Key: "foreignField", Value: "epic"},
					{Key: "as", Value: "tasks"},
				}},
			},
		}

		// Use the defined stages to aggregate data from User_Infor collection
		result, aggregateErr := projectCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating project: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &project)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding project: "+decodeErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"project": project,
		})
	}
}

/*
Search for Projects by title

params: None

return: gin.HandlerFunc Handler function to search for projects by title
*/
func SearchProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Project model
		var projects []model.Project

		// Get the search data from request body
		query := c.Query("q")

		// Get all data from DB where the title is similar to the search data
		//result, queryErr := projectCollection.Find(ctx, bson.M{"title": query})
		result, queryErr := projectCollection.Find(ctx, bson.M{"title": bson.M{"$regex": query, "$options": "i"}})
		if queryErr != nil {
			c.JSON(http.StatusInternalServerError, "Error querying projects: "+queryErr.Error())
			return
		}

		// Decode the data from DB to the projects array
		decodeErr := result.All(ctx, &projects)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding projects: "+decodeErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"count":    len(projects),
			"projects": projects,
		})
	}
}

/*
Update a Project by ID

params: None

return: gin.HandlerFunc Handler function to update a project by ID
*/
func UpdateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Create an instance of the Project model
		var project model.Project

		// Convert the hex string to ObjectID
		updateId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid project ID: "+convertErr.Error())
			return
		}

		// Bind the request body to the Project model
		bindingErr := c.BindJSON(&project)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Temp variable to decode the FindOne result
		var tempResult bson.M
		// Validation error array for Project model
		var projectValidationErr []gin.H
		// Check if there is any validation error
		var validationErrFlg = false
		// Validate the ID existence in DB
		decodeErr := projectCollection.FindOne(ctx, bson.M{"_id": updateId}).Decode(&tempResult)
		if decodeErr != nil {
			// Append the validation error for not found Id
			projectValidationErr = append(projectValidationErr, gin.H{
				"field": "Id",
				"tag":   "not found",
			})

			// Set validation error flag to true
			validationErrFlg = true
		}

		// Process the specified project
		validationErr := validate.Struct(&project)
		if validationErr != nil {
			// For each validation error
			for _, ve := range validationErr.(validator.ValidationErrors) {
				// Append the validation error to the projectValidationErr array
				projectValidationErr = append(projectValidationErr, gin.H{
					"field": ve.Field(),
					"tag":   ve.Tag(),
				})
			}

			// Set validation error flag to true
			validationErrFlg = true
		}

		// If validation error occured
		if validationErrFlg {
			// Return the validation errors to the client
			c.JSON(http.StatusBadRequest, gin.H{
				"validationError": projectValidationErr,
				"msg":             "Update project failed",
			})
			return
		}

		// Update the fields of the project in DB
		update := bson.M{
			"$set": bson.M{
				"title":       project.Title,
				"description": project.Description,
				"updatedAt":   time.Now(),
			},
		}

		// Find and update the project in DB
		result := projectCollection.FindOneAndUpdate(ctx, bson.M{"_id": updateId}, update)
		if result.Err() != nil {
			c.JSON(http.StatusInternalServerError, "Error updating project: "+result.Err().Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"msg": "Update project successfully",
		})
	}
}

/*
Delete a Project by ID

params: None

return: gin.HandlerFunc Handler function to delete a project by ID
*/
func DeleteProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		deleteId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid project ID: "+convertErr.Error())
			return
		}

		// Delete the specified project from DB
		result, deleteErr := projectCollection.DeleteOne(ctx, bson.M{"_id": deleteId})
		if deleteErr != nil {
			c.JSON(http.StatusInternalServerError, "Error deleting project: "+deleteErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"msg": strconv.Itoa(int(result.DeletedCount)) + " project deleted",
		})
	}
}
