package controller

import (
	"backend/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	// "github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserInforGetAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var userInfor []model.UserInfor

		defer cancel()

		//find all userInfor
		results, errCollection := userInforCollection.Find(ctx, bson.M{})

		if errCollection != nil {
			c.JSON(http.StatusInternalServerError, errCollection.Error())
			return
		}

		//reading from the db more optimally
		defer results.Close(ctx)

		//loop through the results
		for results.Next(ctx) {

			var singleUserInfor model.UserInfor

			if err := results.Decode(&singleUserInfor); err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
			}
			userInfor = append(userInfor, singleUserInfor)
		}
		c.JSON(http.StatusOK, userInfor)

	}
}

// Get ID From http link
func GetUserInforByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userinfor_id, convertErr := primitive.ObjectIDFromHex(c.Param("id"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, "Invalid epic ID: "+convertErr.Error())
			return
		}

		fmt.Println("Get UserInfor ID", userinfor_id)
		var userInfors []gin.H

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			//Join employee table with userinfor table
			//match is  conditional sentences
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "_id", Value: userinfor_id},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "manager_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "manager"},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "manager.userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "manager_userinfor"},
				}},
			},
			// bson.D{
			// 	{Key: "$sort", Value: bson.D{
			// 		{Key: "fullname", Value: 1},
			// 	}},
			// },

			// The project variable is used to combine all values
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "profile_image", Value: "$profile_image"},
					{Key: "fullname", Value: "$fullname"},
					{Key: "office", Value: "$office"},
					{Key: "department", Value: "$department"},
					{Key: "position", Value: "$position"},
					{Key: "manager_id", Value: "$manager_id"},
					{Key: "manager_fullname", Value: "$manager_userinfor.fullname"},
					{Key: "email", Value: "$email"},
					{Key: "phone", Value: "$phone"},
					{Key: "address", Value: "$address"},
					{Key: "gender", Value: "$gender"},
				}},
			},
		}

		// Use the $lookup stage to aggregate data from the Epic and Project collections
		result, aggregateErr := userInforCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &userInfors)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}
		if userInfors != nil {
			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"count":      len(userInfors),
				"user_infor": userInfors,
			})
		}

		// id, _ := primitive.ObjectIDFromHex(userinfor_id)
		// filter := bson.D{{"_id", id}}
		// errCollection := userInforCollection.FindOne(ctx, filter).Decode(&userInfor)
		// if errCollection != nil {
		// 	c.JSON(http.StatusInternalServerError, errCollection.Error())
		// 	return
		// }
	}
}

func GetUserInforByIDInDirect(UserInforID string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userinfor_id, convertErr := primitive.ObjectIDFromHex(UserInforID)
	if convertErr != nil {
		// c.JSON(http.StatusBadRequest, "Invalid epic ID: "+convertErr.Error())
		fmt.Println("Invalid user information ID")
		return nil, errors.New("Invalid user information ID")
	}

	fmt.Println("Get UserInfor ID", userinfor_id)
	// var userInfors []gin.H
	// var userInfors map[string]interface{}
	type userInforsTypStruct struct {
		Id               string `bson:"_id"`
		FullName         string `bson:"fullname"`
		Profile_Image    string `bson:"profile_image"`
		Office           string `bson:"office"`
		Department       string `bson:"department"`
		Position         string `bson:"position"`
		Manager_ID       string `bson:"manager_id"`
		Manager_FullName string `bson:"manager_fullname"`
		Subordinates     string `bson:"subordinates"`
		Phone            string `bson:"phone"`
		Email            string `bson:"email"`
		Gender           string `bson:"gender"`
		Address          string `bson:"address"`
	}
	var userInfors []userInforsTypStruct

	// Define pipeline to join collections and sort the result
	pipeline := mongo.Pipeline{
		//Join employee table with userinfor table
		//match is  conditional sentences
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "_id", Value: userinfor_id},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "employee"},
				{Key: "localField", Value: "manager_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "manager"},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "user_infor"},
				{Key: "localField", Value: "manager.userinfor_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "manager_userinfor"},
			}},
		},
		// bson.D{
		// 	{Key: "$sort", Value: bson.D{
		// 		{Key: "fullname", Value: 1},
		// 	}},
		// },

		// The project variable is used to combine all values
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: "$_id"},
				{Key: "profile_image", Value: "$profile_image"},
				{Key: "fullname", Value: "$fullname"},
				{Key: "office", Value: "$office"},
				{Key: "department", Value: "$department"},
				{Key: "position", Value: "$position"},
				{Key: "manager_id", Value: "$manager_id"},
				// {Key: "manager_fullname", Value: "$manager_userinfor.fullname"},
				{Key: "manager_fullname", Value: bson.D{
					{Key: "$arrayElemAt", Value: bson.A{"$manager_userinfor.fullname", 0}},
				}},
				{Key: "email", Value: "$email"},
				{Key: "phone", Value: "$phone"},
				{Key: "address", Value: "$address"},
				{Key: "gender", Value: "$gender"},
			}},
		},
	}

	// Use the $lookup stage to aggregate data from the Epic and Project collections
	result, aggregateErr := userInforCollection.Aggregate(ctx, pipeline)
	if aggregateErr != nil {
		// c.JSON(http.StatusInternalServerError, "Error aggregating user information: "+aggregateErr.Error())
		fmt.Println("Error aggregating user information:" + aggregateErr.Error())
		return nil, errors.New("Error aggregating user information:" + aggregateErr.Error())
	}

	// Decode the data from DB to the epics array
	decodeErr := result.All(ctx, &userInfors)
	if decodeErr != nil {
		// c.JSON(http.StatusInternalServerError, "Error decoding user information: "+decodeErr.Error())
		fmt.Println("Error decoding user information: ", decodeErr.Error())
		return nil, errors.New("Error decoding user information:" + decodeErr.Error())
	}

	if userInfors != nil {
		// Send response to client
		// c.JSON(http.StatusOK, gin.H{
		// 	"count":      len(userInfors),
		// 	"user_infor": userInfors,
		// })
		fmt.Println("User infor:", userInfors[0].Id)

		data := map[string]interface{}{
			"_id":              userInfors[0].Id,
			"profile_image":    userInfors[0].Profile_Image,
			"fullname":         userInfors[0].FullName,
			"office":           userInfors[0].Office,
			"department":       userInfors[0].Department,
			"position":         userInfors[0].Position,
			"manager_id":       userInfors[0].Manager_ID,
			"manager_fullname": userInfors[0].Manager_FullName,
			"email":            userInfors[0].Email,
			"phone":            userInfors[0].Phone,
			"address":          userInfors[0].Address,
			"gender":           userInfors[0].Gender,
		}
		return data, nil
	}
	return nil, errors.New("No Data")
}

// func AddUserInfor() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		// user := c.MustGet("account_logged").(jwt.MapClaims)
// 		// accountId := user["account_id"].(string)
// 		var my_account = global_account_token
// 		if my_account == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		token, err := jwt.Parse(my_account, func(token *jwt.Token) (interface{}, error) {
// 			return []byte("secret_key"), nil
// 		})
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 			c.Abort()
// 			return
// 		}

// 		//Get accountId
// 		var accountId string
// 		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 			// Access user information from claims
// 			accountId = claims["account_id"].(string)
// 		} else {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
// 			c.Abort()
// 		}

// 		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
// 		defer cancel()

// 		var user_infor model.UserInfor

// 		//validate the request body
// 		if err := c.BindJSON(&user_infor); err != nil {
// 			c.JSON(http.StatusBadRequest, "Request error"+err.Error())
// 			return
// 		}

// 		//use the validator library to validate required fields
// 		if validationErr := validate.Struct(&user_infor); validationErr != nil {
// 			c.JSON(http.StatusBadRequest, "Error, missing field"+validationErr.Error())
// 			return
// 		}

// 		newUserInfor := model.UserInfor{
// 			Id:           primitive.NewObjectID(),
// 			FullName:     user_infor.FullName,
// 			Office:       user_infor.Office,
// 			Department:   user_infor.Department,
// 			Position:     user_infor.Position,
// 			Manager_ID:   user_infor.Manager_ID,
// 			Subordinates: user_infor.Subordinates,
// 			Phone:        user_infor.Phone,
// 			Email:        user_infor.Email,
// 			Gender:       user_infor.Gender,
// 			Address:      user_infor.Address,
// 			CreatedAt:    time.Now().Unix(),
// 			UpdatedAt:    time.Now().Unix(),
// 		}

// 		result, insertErr := userInforCollection.InsertOne(ctx, newUserInfor)
// 		if insertErr != nil {
// 			c.JSON(http.StatusInternalServerError, "Error inserting account"+insertErr.Error())
// 			return
// 		} else {
// 			fmt.Println("Insert user information Successful into Database with ID:", result.InsertedID)
// 			c.JSON(http.StatusCreated, gin.H{
// 				"state":     "success",
// 				"userInfor": newUserInfor,
// 				"result":    result,
// 			})
// 			var employeeModel model.Employee
// 			UserObjecID, _ := primitive.ObjectIDFromHex(newUserInfor.Id.Hex())
// 			accountid, _ := primitive.ObjectIDFromHex(accountId)

// 			filter := bson.D{{Key: "account_id", Value: accountid}}
// 			update := bson.D{{Key: "$set", Value: bson.D{{Key: "userinfor_id", Value: UserObjecID}}}}

// 			errFindEmployee := employeeCollection.FindOneAndUpdate(
// 				ctx, filter, update).Decode(&employeeModel)
// 			if errFindEmployee != nil {
// 				c.JSON(http.StatusNotImplemented, gin.H{
// 					"message": "Update user information fail",
// 					"state":   "fail",
// 					"error":   errFindEmployee.Error(),
// 				})
// 			} else {
// 				c.JSON(http.StatusCreated, gin.H{
// 					"message":               "Create userinfor and update employee success",
// 					"state":                 "success",
// 					"employee just updated": employeeModel,
// 				})
// 			}

// 		}

// 	}
// }

// func AddUserInforJustCreate() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
// 		defer cancel()

// 		var user_infor model.UserInfor
// 		//validate the request body
// 		if err := c.BindJSON(&user_infor); err != nil {
// 			c.JSON(http.StatusBadRequest, "Request error"+err.Error())
// 			return
// 		}

// 		//use the validator library to validate required fields
// 		if validationErr := validate.Struct(&user_infor); validationErr != nil {
// 			c.JSON(http.StatusBadRequest, "Error, missing field"+validationErr.Error())
// 			return
// 		}

// 		newUserInfor := model.UserInfor{
// 			Id:           primitive.NewObjectID(),
// 			FullName:     user_infor.FullName,
// 			Office:       user_infor.Office,
// 			Department:   user_infor.Department,
// 			Position:     user_infor.Position,
// 			Manager_ID:   user_infor.Manager_ID,
// 			Subordinates: user_infor.Subordinates,
// 			Phone:        user_infor.Phone,
// 			Email:        user_infor.Email,
// 			Gender:       user_infor.Gender,
// 			Address:      user_infor.Address,
// 			CreatedAt:    time.Now().Unix(),
// 			UpdatedAt:    time.Now().Unix(),
// 		}
// 		result, insertErr := userInforCollection.InsertOne(ctx, newUserInfor)
// 		if insertErr != nil {
// 			c.JSON(http.StatusInternalServerError, "Error inserting account"+insertErr.Error())
// 			return
// 		} else {
// 			fmt.Println("Insert user information Successful into Database with ID:", result.InsertedID)
// 			c.JSON(http.StatusCreated, gin.H{
// 				"state":     "success",
// 				"userInfor": newUserInfor,
// 				"result":    result,
// 			})

// 			UserObjecID, _ := primitive.ObjectIDFromHex(newUserInfor.Id.Hex())
// 			employeeFilter, _ := primitive.ObjectIDFromHex(global_employee_just_created)

// 			filter := bson.D{{Key: "_id", Value: employeeFilter}}
// 			update := bson.D{{Key: "$set", Value: bson.D{{Key: "userinfor_id", Value: UserObjecID}}}}

// 			var employeeModel model.Employee
// 			errFindEmployee := employeeCollection.FindOneAndUpdate(
// 				ctx, filter, update).Decode(&employeeModel)
// 			if errFindEmployee != nil {
// 				c.JSON(http.StatusNotImplemented, gin.H{
// 					"message": "Update employee fail",
// 					"state":   "fail",
// 					"error":   errFindEmployee.Error(),
// 				})
// 			} else {
// 				c.JSON(http.StatusCreated, gin.H{
// 					"message":               "Create userinfor and update employee success",
// 					"state":                 "success",
// 					"employee just updated": employeeModel,
// 				})
// 			}
// 		}

// 	}
// }

func UpdateProfileImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		//must have to connect database
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		var user_infor model.UserInfor

		//get id from link
		var GetUserInforID = c.Param("id")
		//convert to primary object for ID
		id, _ := primitive.ObjectIDFromHex(GetUserInforID)

		// if err := c.BindJSON(&getUserInforUpdate); err != nil {
		// 	c.JSON(http.StatusBadRequest, "Request error"+err.Error())
		// 	return
		// }

		GetData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("request data error!!!")
		}

		// fmt.Printf("ctx.Request.body: %v", string(Data))
		// var jsonData account_struct // map[string]interface{}
		//Create struct type to save data with model
		var getUserInforUpdate model.UserInfor // map[string]interface{}
		e := json.Unmarshal(GetData, &getUserInforUpdate)
		if e != nil {
			fmt.Println("Error Get Data from frontend:", err)
		}

		fmt.Println("Get Profile Image Name: ", getUserInforUpdate.Profile_Image)
		filter := bson.D{{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "profile_image", Value: getUserInforUpdate.Profile_Image},
			{Key: "createdAt", Value: time.Now().Unix()},
		}}}

		if errUpdateProfileImage := userInforCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user_infor); errUpdateProfileImage != nil {
			fmt.Println("Error update Profile Image : ", errUpdateProfileImage)
			c.JSON(http.StatusInternalServerError, "Error updating project"+errUpdateProfileImage.Error())
			return
		} else {
			fmt.Println("Update Profile Image successfully: ", user_infor.Profile_Image)
			c.JSON(http.StatusOK, gin.H{
				"message":   "Updated account successfully",
				"state":     "success",
				"userInfor": user_infor,
			})
		}

	}
}

func UpdateUserInfor1() gin.HandlerFunc {
	return func(c *gin.Context) {

		//must have to connect database
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		var user_infor model.UserInfor

		//get id from link
		var GetUserInforID = c.Param("id")
		//convert to primary object for ID
		id, _ := primitive.ObjectIDFromHex(GetUserInforID)

		//Create struct type to save data with model
		var getUserInforUpdate model.UserInfor

		//Get Data from frontend
		jsonData := c.Request.FormValue("userInforData")
		errUnmarShal := json.Unmarshal([]byte(jsonData), &getUserInforUpdate)
		if errUnmarShal != nil {
			c.JSON(http.StatusBadRequest, "Request error"+errUnmarShal.Error())
			return
		}

		//Check if file that get from frontend  is null

		fmt.Println("Get Fullname to update:", getUserInforUpdate.FullName)
		fmt.Println("Get Office to update:", getUserInforUpdate.Office)

		// Manager_ID_Convert, _ := primitive.ObjectIDFromHex(getUserInforUpdate.Manager_ID.Hex())
		filter := bson.D{{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "fullname", Value: getUserInforUpdate.FullName},
			{Key: "office", Value: getUserInforUpdate.Office},
			{Key: "department", Value: getUserInforUpdate.Department},
			{Key: "position", Value: getUserInforUpdate.Position},
			{Key: "email", Value: getUserInforUpdate.Email},
			{Key: "phone", Value: getUserInforUpdate.Phone},
			{Key: "gender", Value: getUserInforUpdate.Gender},
			{Key: "address", Value: getUserInforUpdate.Address},
			{Key: "manager_id", Value: getUserInforUpdate.Manager_ID},
			{Key: "createdAt", Value: time.Now().Unix()},
		}}} // "password", "asdfadfafs",

		if err1 := userInforCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user_infor); err1 != nil {
			c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Updated account successfully",
			})
		}
	}
}
func UpdateUserInfor() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		validate := validator.New()

		// Register the custom required validation function
		_ = validate.RegisterValidation("customrequired", CheckCustomRequired)

		// Create an array of the Task model
		var user_infor model.UserInfor

		//get id from link
		var GetUserInforID = c.Param("id")
		//convert to primary object for ID
		id, _ := primitive.ObjectIDFromHex(GetUserInforID)

		// Bind the request body to the task model
		bindingErr := c.BindJSON(&user_infor)
		if bindingErr != nil {
			c.JSON(http.StatusBadRequest, "Request binding error: "+bindingErr.Error())
			return
		}

		filter := bson.D{{Key: "_id", Value: id}}
		var update bson.D
		if user_infor.Profile_Image == "" {
			update = bson.D{{Key: "$set", Value: bson.D{
				{Key: "fullname", Value: user_infor.FullName},
				{Key: "email", Value: user_infor.Email},
				{Key: "phone", Value: user_infor.Phone},
				{Key: "gender", Value: user_infor.Gender},
				{Key: "address", Value: user_infor.Address},
				{Key: "updatedAt", Value: time.Now().Unix()},
			}}}
		} else {
			update = bson.D{{Key: "$set", Value: bson.D{
				{Key: "fullname", Value: user_infor.FullName},
				{Key: "email", Value: user_infor.Email},
				{Key: "phone", Value: user_infor.Phone},
				{Key: "gender", Value: user_infor.Gender},
				{Key: "address", Value: user_infor.Address},
				{Key: "profile_image", Value: user_infor.Profile_Image},
				{Key: "updatedAt", Value: time.Now().Unix()},
			}}}
		}

		if err1 := userInforCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user_infor); err1 != nil {
			c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Account updated",
			})
		}
	}
}

// func UpdateUserInfor() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		//must have to connect database
// 		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
// 		defer cancel()

// 		var user_infor model.UserInfor

// 		//get id from link
// 		var GetUserInforID = c.Param("id")
// 		//convert to primary object for ID
// 		id, _ := primitive.ObjectIDFromHex(GetUserInforID)

// 		//Create struct type to save data with model
// 		var getUserInforUpdate model.UserInfor

// 		//Get Data from frontend
// 		jsonData := c.Request.FormValue("userInforData")
// 		errUnmarShal := json.Unmarshal([]byte(jsonData), &getUserInforUpdate)
// 		if errUnmarShal != nil {
// 			c.JSON(http.StatusBadRequest, "Request error"+errUnmarShal.Error())
// 			return
// 		}

// 		//Get File Data
// 		file, errFormFile := c.FormFile("profile_image")
// 		if errFormFile != nil {
// 			// Handle the error gracefully
// 			fmt.Println("Error retrieving file:", errFormFile)
// 			c.Error(errFormFile) // Report the error but continue execution
// 		}

// 		//Check if file that get from frontend  is null
// 		if file == nil {
// 			fmt.Println("Get Fullname to update:", getUserInforUpdate.FullName)
// 			fmt.Println("Get Office to update:", getUserInforUpdate.Office)

// 			// Manager_ID_Convert, _ := primitive.ObjectIDFromHex(getUserInforUpdate.Manager_ID.Hex())
// 			filter := bson.D{{Key: "_id", Value: id}}
// 			update := bson.D{{Key: "$set", Value: bson.D{
// 				{Key: "fullname", Value: getUserInforUpdate.FullName},
// 				{Key: "office", Value: getUserInforUpdate.Office},
// 				{Key: "department", Value: getUserInforUpdate.Department},
// 				{Key: "position", Value: getUserInforUpdate.Position},
// 				{Key: "email", Value: getUserInforUpdate.Email},
// 				{Key: "phone", Value: getUserInforUpdate.Phone},
// 				{Key: "gender", Value: getUserInforUpdate.Gender},
// 				{Key: "address", Value: getUserInforUpdate.Address},
// 				{Key: "manager_id", Value: getUserInforUpdate.Manager_ID},
// 				{Key: "createdAt", Value: time.Now().Unix()},
// 			}}} // "password", "asdfadfafs",

// 			if err1 := userInforCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user_infor); err1 != nil {
// 				c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
// 				return
// 			} else {
// 				c.JSON(http.StatusOK, gin.H{
// 					"message": "Updated account successfully",
// 				})
// 			}
// 		} else {

// 			//Show file name just get frome frontend
// 			// fmt.Println("File name:", file.Filename)

// 			//Create profile image name auto
// 			profile_image_name := uuid.New()
// 			profile_image_Save := profile_image_name.String() + ".jpg"

// 			fmt.Println("File name:", profile_image_name.String())

// 			//Link folder to save image
// 			folderPathToSaveImage := "../frontend/src/lib/images/profile/"

// 			//Check if the folder exists
// 			if !folderExists(folderPathToSaveImage) {
// 				//Create the folder if it doesn't exist
// 				err := os.MkdirAll(folderPathToSaveImage, os.ModePerm)
// 				if err != nil {
// 					// If there's an error creating the folder, return an error response
// 					c.JSON(http.StatusInternalServerError, gin.H{"error Create Folder ": err.Error()})
// 					return
// 				}
// 			}

// 			// Define the desired path where you want to save the file
// 			filePath := folderPathToSaveImage + profile_image_Save

// 			// Create a new file at the desired path
// 			newFile, err := os.Create(filePath)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error create new image": err.Error()})
// 				return
// 			}
// 			defer newFile.Close()

// 			// Save the uploaded file to the newly created file
// 			if err := c.SaveUploadedFile(file, filePath); err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error save upload image": err.Error()})
// 				return
// 			}

// 			fmt.Println("Get Fullname to update:", getUserInforUpdate.FullName)
// 			fmt.Println("Get Office to update:", getUserInforUpdate.Office)

// 			// Manager_ID_Convert, _ := primitive.ObjectIDFromHex(getUserInforUpdate.Manager_ID.Hex())
// 			filter := bson.D{{Key: "_id", Value: id}}
// 			update := bson.D{{Key: "$set", Value: bson.D{
// 				{Key: "profile_image", Value: profile_image_Save},
// 				{Key: "fullname", Value: getUserInforUpdate.FullName},
// 				{Key: "office", Value: getUserInforUpdate.Office},
// 				{Key: "department", Value: getUserInforUpdate.Department},
// 				{Key: "position", Value: getUserInforUpdate.Position},
// 				{Key: "email", Value: getUserInforUpdate.Email},
// 				{Key: "phone", Value: getUserInforUpdate.Phone},
// 				{Key: "gender", Value: getUserInforUpdate.Gender},
// 				{Key: "address", Value: getUserInforUpdate.Address},
// 				{Key: "manager_id", Value: getUserInforUpdate.Manager_ID},
// 				{Key: "createdAt", Value: time.Now().Unix()},
// 			}}} // "password", "asdfadfafs",

// 			//Get User Infor contain profile image that prepare to update
// 			GetUserInforToDelete, GetUserInforToDeleteError := GetUserInforByIDInDirect(GetUserInforID)

// 			if err1 := userInforCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user_infor); err1 != nil {
// 				c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
// 				return
// 			} else {

// 				if GetUserInforToDeleteError != nil {
// 					c.JSON(http.StatusInternalServerError, "Error Get User Infor To Delete")
// 					return
// 				}

// 				//Get profile image to delete
// 				ProfileImageToDelete, _ := GetUserInforToDelete["profile_image"].(string)
// 				folderPathToDelete := folderPathToSaveImage
// 				LinkAndProfileImageToDelete := folderPathToDelete + ProfileImageToDelete

// 				errDeleteProfileImage := os.Remove(LinkAndProfileImageToDelete)
// 				if err != nil {
// 					c.JSON(http.StatusInternalServerError, gin.H{"Error Delete Profile Image": errDeleteProfileImage.Error()})
// 					return
// 				}

// 				c.JSON(http.StatusOK, gin.H{
// 					"message": "Updated account successfully",
// 				})
// 			}
// 		}

// 	}
// }

// func UpdateUserInfor1() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		//must have to connect database
// 		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
// 		defer cancel()

// 		var user_infor model.UserInfor

// 		//get id from link
// 		var GetUserInforID = c.Param("id")
// 		//convert to primary object for ID
// 		id, _ := primitive.ObjectIDFromHex(GetUserInforID)

// 		var getUserInforUpdate model.UserInfor
// 		//validate the request body
// 		if err := c.BindJSON(&getUserInforUpdate); err != nil {
// 			c.JSON(http.StatusBadRequest, "Request error"+err.Error())
// 			return
// 		}

// 		//use the validator library to validate required fields
// 		if validationErr := validate.Struct(&getUserInforUpdate); validationErr != nil {
// 			c.JSON(http.StatusBadRequest, "Error, missing field"+validationErr.Error())
// 			return
// 		}
// 		fmt.Println("Get Fullname to update:", getUserInforUpdate.FullName)
// 		fmt.Println("Get Office to update:", getUserInforUpdate.Office)

// 		//get file from request
// 		// file, err := c.FormFile("image")
// 		// if err != nil {
// 		// 	c.JSON(http.StatusInternalServerError, gin.H{
// 		// 		"message": "get bad request file",
// 		// 	})
// 		// 	return
// 		// }
// 		// // var filePath string
// 		// fmt.Println("File name", file.Filename)

// 		// profile_image := uuid.New()
// 		// // Define the desired path where you want to save the file
// 		// filePath = "../frontend/src/lib/images/profile/" + profile_image.String() + ".jpg"

// 		// // Create a new file at the desired path
// 		// newFile, err := os.Create(filePath)
// 		// if err != nil {
// 		// 	c.JSON(http.StatusInternalServerError, gin.H{"error create new image": err.Error()})
// 		// 	return
// 		// }
// 		// defer newFile.Close()

// 		// // Save the uploaded file to the newly created file
// 		// if err := c.SaveUploadedFile(file, filePath); err != nil {
// 		// 	c.JSON(http.StatusInternalServerError, gin.H{"error save upload image": err.Error()})
// 		// 	return
// 		// }

// 		// Manager_ID_Convert, _ := primitive.ObjectIDFromHex(getUserInforUpdate.Manager_ID.Hex())
// 		filter := bson.D{{Key: "_id", Value: id}}
// 		update := bson.D{{Key: "$set", Value: bson.D{
// 			// {Key: "profile_image", Value: profile_image},
// 			{Key: "fullname", Value: getUserInforUpdate.FullName},
// 			{Key: "office", Value: getUserInforUpdate.Office},
// 			{Key: "department", Value: getUserInforUpdate.Department},
// 			{Key: "position", Value: getUserInforUpdate.Position},
// 			{Key: "email", Value: getUserInforUpdate.Email},
// 			{Key: "phone", Value: getUserInforUpdate.Phone},
// 			{Key: "gender", Value: getUserInforUpdate.Gender},
// 			{Key: "address", Value: getUserInforUpdate.Address},
// 			{Key: "manager_id", Value: getUserInforUpdate.Manager_ID},
// 			{Key: "createdAt", Value: time.Now().Unix()},
// 		}}} // "password", "asdfadfafs",

// 		if err1 := userInforCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user_infor); err1 != nil {
// 			c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
// 			return
// 		} else {
// 			c.JSON(http.StatusOK, gin.H{
// 				"message": "Updated account successfully",
// 			})
// 		}
// 	}
// }
