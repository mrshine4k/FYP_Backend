/*
Controller for handling data with Epic model in DB

1. CreateEpic: Create one or many Epics

2. GetEpics: Get all Epics from the DB

3. GetEpicById: Get the Epic by specified ID

4. SearchEpic: Search for Epics by title

5. UpdateEpic: Update one or many Epics by specified ID(s)

6. DeleteEpic: Delete one or many Epics by specified ID(s)

7. ConvertEpicsToInterface: Convert the epics array to an interface array
*/
package controller

import (
	"backend/model"
	"context"
	"encoding/json"
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
Create one or many Epics

params: None

return: gin.HandlerFunc Handler function to create one or many epics
*/
func CreateEpic() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Register the custom required validation function
		_ = validate.RegisterValidation("customrequired", CheckCustomRequired)

		// Create an array of the Epic model
		var epic model.Epic

		// Bind the request body to the epic model
		bindingErr := c.BindJSON(&epic)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		epic.Id = primitive.NewObjectID()
		epic.CreatedAt = time.Now()
		epic.UpdatedAt = time.Now()

		// Insert the specified document to DB
		result, insertErr := epicCollection.InsertOne(ctx, epic)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error inserting epic: "+insertErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusCreated, gin.H{
			"inserted ID": result.InsertedID,
		})

		// // Check if validation failed for any epic in the array
		// var validationErrFlg = false
		// // Validation result array
		// var validationErrResult []gin.H
		// // Process the specified epics
		// for i, epic := range epics {
		// 	// Validate each epic
		// 	validationErr := validate.Struct(epic)
		// 	if validationErr != nil {
		// 		// If validation failed, create a single error object for the each epic
		// 		singleValidationErr := gin.H{
		// 			"element": i + 1,
		// 			"error":   []gin.H{},
		// 		}

		// 		// Append the field and tag to the error array in the single error object
		// 		for _, ve := range validationErr.(validator.ValidationErrors) {
		// 			singleValidationErr["error"] = append(singleValidationErr["error"].([]gin.H), gin.H{
		// 				"field": ve.Field(),
		// 				"tag":   ve.Tag(),
		// 			})
		// 		}

		// 		// Set the validation error flag to true
		// 		validationErrFlg = true

		// 		// Append the single error object to the validation error result array
		// 		validationErrResult = append(validationErrResult, singleValidationErr)
		// 		continue
		// 	}

		// 	// If validation succeed, set the ID and timestamp for each epic
		// 	epic.Id = primitive.NewObjectID()
		// 	epic.CreatedAt = time.Now()
		// 	epic.UpdatedAt = time.Now()
		// 	epics[i] = epic
		// }

		// // If validation failed for any epic in the array
		// if validationErrFlg {
		// 	// Return the validation errors to the client
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"validationError": validationErrResult,
		// 	})
		// 	return
		// }

		// // Check the length of epics array to insert appropriately
		// if len(epics) == 1 {
		// 	// Insert the specified document to DB
		// 	result, insertErr := epicCollection.InsertOne(ctx, epics[0])
		// 	if insertErr != nil {
		// 		c.JSON(http.StatusInternalServerError, "Error inserting epic: "+insertErr.Error())
		// 		return
		// 	}

		// 	// Send response to client
		// 	c.JSON(http.StatusCreated, gin.H{
		// 		"inserted ID": result.InsertedID,
		// 	})
		// } else if len(epics) > 1 {
		// 	// Insert all specified documents to DB
		// 	result, insertErr := epicCollection.InsertMany(ctx, ConvertEpicsToInterface(epics))
		// 	if insertErr != nil {
		// 		c.JSON(http.StatusInternalServerError, "Error inserting epics: "+insertErr.Error())
		// 		return
		// 	}

		// 	// Send response to client
		// 	c.JSON(http.StatusCreated, gin.H{
		// 		"inserted IDs": result.InsertedIDs,
		// 	})
		// } else {
		// 	// If the epics array is empty return an error
		// 	c.JSON(http.StatusBadRequest, "Empty JSON array")
		// }
	}
}

func GetLeaderForEpic() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to an ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid epic ID: "+convertErr.Error())
			return
		}

		// Epic for the result
		var epic gin.H

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "_id", Value: queryId},
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
		}

		// Use the $lookup stage to aggregate data from the Epic and Project collections
		result, aggregateErr := epicCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		if result.Next(ctx) {
			decodeErr := result.Decode(&epic)
			if decodeErr != nil {
				c.JSON(http.StatusInternalServerError, "Error decoding epic: "+decodeErr.Error())
				return
			}
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		var project1 = epic["project"].(primitive.A)
		var project = project1[0].(gin.H)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"leader":  project["leader"].(primitive.ObjectID),
		})
	}
}

/*
Get all Epics from the DB

params: None

return: gin.HandlerFunc Handler function to get all epics
*/
func GetEpics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Epic model
		//var epics []model.Epic // Use for the FindOne population method
		var epics []gin.H

		/* Manual application side population, only use when need to manipulate the data after fetching from DB.
		/* Otherwise, it is less efficient than using $lookup stage in aggregation pipeline
		// Get all data from DB
		result, queryErr := epicCollection.Find(ctx, bson.M{})
		if queryErr != nil {
			c.JSON(http.StatusInternalServerError, "Error querying epics: "+queryErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &epics)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}

		// Temp instance of the Project model
		var project model.Project
		// For each epic in the epics array
		for i, epic := range epics {
			// Find the project with the ID specified in the Project collection
			queryErr := projectCollection.FindOne(ctx, bson.M{"_id": epic.Project}).Decode(&project)
			if queryErr != nil {
				c.JSON(http.StatusInternalServerError, "Error querying project: "+queryErr.Error())
				return
			}
			epics[i].ProjectPopulated = project
		}
		*/

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "projects"},
					{Key: "localField", Value: "project"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "project"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "epic", Value: "$title"},
					{Key: "project", Value: "$project.title"},
				}},
			},
		}

		// Use the $lookup stage to aggregate data from the Epic and Project collections
		result, aggregateErr := epicCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &epics)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"count": len(epics),
			"epics": epics,
		})
	}
}

/*
Get the Epic by specified ID

params: None

return: gin.HandlerFunc Handler function to get the epic by ID
*/
func GetEpicById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an instance of the Epic model
		var epic []gin.H

		// Convert the hex string to an ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid epic ID: "+convertErr.Error())
			return
		}

		/* Manual application side search, only use when need to manipulate the data after fetching from DB.
		// Query DB for the epic with the given ID
		queryErr := epicCollection.FindOne(ctx, bson.M{"_id": queryId}).Decode(&epic)
		if queryErr != nil {
			c.JSON(http.StatusInternalServerError, "Error querying epic: "+queryErr.Error())
			return
		}
		*/

		// Define pipeline to filter the data by ID and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "_id", Value: queryId},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "tasks"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "epic"},
					{Key: "as", Value: "tasks"},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Epic and Project collections
		result, aggregateErr := epicCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epic: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &epic)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epic: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"epic":    epic[0],
			"success": true,
		})
	}
}

/*
Search for Epics by title

params: None

return: gin.HandlerFunc Handler function to search for epics
*/
func SearchEpic() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Epic model
		var epics []gin.H

		// Get the search data from request query
		query := c.Query("q")

		/* Manual application side search, only use when need to manipulate the data after fetching from DB.
		// Get all data from DB where the title is similar to the search data
		result, queryErr := epicCollection.Find(ctx, bson.M{"title": bson.M{"$regex": query, "$options": "i"}})
		if queryErr != nil {
			c.JSON(http.StatusInternalServerError, "Error querying epics: "+queryErr.Error())
			return
		}
		*/

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
					{Key: "from", Value: "projects"},
					{Key: "localField", Value: "project"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "project"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Epic and Project collections
		result, aggregateErr := epicCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epic: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &epics)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"count": len(epics),
			"epics": epics,
		})
	}
}

/*
Get all epics for a specified project

params: None

return: gin.HandlerFunc Handler function to get all epics for a specified project
*/
func GetEpicForProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Epic model
		var epics []gin.H

		// Convert the hex string to an ObjectID
		projectId, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid project ID: "+convertErr.Error())
			return
		}

		// Define pipeline to filter the data by ID and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "project", Value: projectId},
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
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "tasks"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "epic"},
					{Key: "as", Value: "tasks"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "tasks.members"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "members_employee"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "members_employee.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "members_userinfor"},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "title", Value: 1},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Epic and Project collections
		result, aggregateErr := epicCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &epics)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"epics":   epics,
		})
	}
}

/*
Update one or many Epics by specified ID(s)

params: None

return: gin.HandlerFunc Handler function to update one or many epics
*/
func UpdateEpic() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Register the custom required validation function
		_ = validate.RegisterValidation("customrequired", DummyValidation)

		// Create an array of the Epic
		var epics []model.Epic

		// Bind the request body to the epic model
		bindingErr := c.BindJSON(&epics)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		// Check if validation failed for any epic in the array
		var validationErrFlg = false
		// The result of the validation process, will be used to return the validation error to the client
		var validationErrResult []gin.H
		// Process the specified epics
		for i, epic := range epics {
			// Check if validation failed for the current epic
			var epicErr = false
			// The result of the validation process for the current epic
			var singleValidationErr gin.H

			// Check if the Id is empty or not
			if epic.Id.IsZero() || epic.Id.Hex() == "" {
				// Create a single error object for the current epic
				singleValidationErr = gin.H{
					"element": i + 1,
					"error":   []gin.H{},
				}

				// Add the field and tag to the error array
				singleValidationErr["error"] = append(singleValidationErr["error"].([]gin.H), gin.H{
					"field": "Id",
					"tag":   "required",
				})

				// Set the epic error flag to true
				epicErr = true
			}

			// Temp variable to decode the FindOne result
			var result bson.M
			// Validate the ID existence in DB
			decodeErr := epicCollection.FindOne(ctx, bson.M{"_id": epic.Id}).Decode(&result)
			if decodeErr != nil {
				if !epicErr {
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

				// Set the epic error flag to true
				epicErr = true
			}

			// Validate each epic
			validationErr := validate.Struct(epic)
			if validationErr != nil {
				if !epicErr {
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

			// If validation failed for the current epic
			if singleValidationErr != nil {
				// Add the single validation error to the validation error result array
				validationErrResult = append(validationErrResult, singleValidationErr)

				// Set validation error flag to true to return the error to client later
				validationErrFlg = true
			}
		}

		// If validation failed for any epic in the array
		if validationErrFlg {
			// Return the validation error to the client
			c.JSON(http.StatusBadRequest, gin.H{
				"validationError": validationErrResult,
			})
			return
		}

		// Check the length of epics array to update appropriately
		if len(epics) == 1 {
			// Update the fields of an epic in DB
			update := bson.M{
				"$set": bson.M{
					"title":       epics[0].Title,
					"description": epics[0].Description,
					"updatedAt":   time.Now(),
				},
			}

			// Find and update the epic in DB
			result := epicCollection.FindOneAndUpdate(ctx, bson.M{"_id": epics[0].Id}, update)
			if result.Err() != nil {
				c.JSON(http.StatusInternalServerError, "Error updating epic: "+result.Err().Error())
				return
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": "1 epic updated",
			})
		} else if len(epics) > 1 {
			// Count the number of documents updated
			var modifyCount int

			for i, epic := range epics {
				// Update the fields of each epic in DB
				update := bson.M{
					"$set": bson.M{
						"title":       epic.Title,
						"description": epic.Description,
						"updatedAt":   time.Now(),
					},
				}

				// Find and update each epic in DB
				result := epicCollection.FindOneAndUpdate(ctx, bson.M{"_id": epic.Id}, update)
				if result.Err() != nil {
					c.JSON(http.StatusInternalServerError, "Error updating epic "+strconv.Itoa(i+1)+": "+result.Err().Error())
					return
				}

				// After each successful update, increment the modifyCount
				modifyCount++
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": strconv.Itoa(modifyCount) + " epics updated",
			})
		} else {
			// If the epics array is empty return an error
			c.JSON(http.StatusBadRequest, "Empty JSON array")
		}
	}
}

/*
Delete one or many Epics by specified ID(s)

params: None

return: gin.HandlerFunc Handler function to delete one or many epics
*/
func DeleteEpic() gin.HandlerFunc {
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
				c.JSON(http.StatusBadRequest, "Invalid epic ID: "+convertErr.Error())
				return
			}
			deleteArr = append(deleteArr, id)
		}

		// Check the length of the delete array to delete appropriately
		if len(deleteArr) == 1 {
			// Delete the specified document from DB
			result, deleteErr := epicCollection.DeleteOne(ctx, bson.M{"_id": deleteArr[0]})
			if deleteErr != nil {
				c.JSON(http.StatusInternalServerError, "Error deleting epic: "+deleteErr.Error())
				return
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": strconv.Itoa(int(result.DeletedCount)) + " epic deleted",
			})
		} else if len(deleteArr) > 1 {
			// Delete all specified documents from DB
			result, deleteErr := epicCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": deleteArr}})
			if deleteErr != nil {
				c.JSON(http.StatusInternalServerError, "Error deleting epic: "+deleteErr.Error())
				return
			}

			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"msg": strconv.Itoa(int(result.DeletedCount)) + " epics deleted",
			})
		} else {
			// If the epics array is empty return an error
			c.JSON(http.StatusBadRequest, "Empty JSON array")
		}
	}
}

/*
Convert the epics array to an interface array

params: epics []model.Epic Array of epics to convert to interface array

return: []interface{} The result array of the conversion
*/
func ConvertEpicsToInterface(epics []model.Epic) []interface{} {
	// Create an interface array with the length of the epics array
	epicInterface := make([]interface{}, len(epics))

	// Bring each epic to the interface array
	for i, epic := range epics {
		epicInterface[i] = epic
	}

	// Return the interface array
	return epicInterface
}
