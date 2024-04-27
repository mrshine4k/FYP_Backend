package controller

import (
	"backend/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		var dataMap = make(map[string]interface{})

		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Parse JSON từ dữ liệu đọc được
		var jsonData map[string]interface{}
		if err := json.Unmarshal(requestBody, &jsonData); err != nil {
			c.JSON(http.StatusBadRequest, "Failed to parse JSON")
			return
		}
		// fmt.Println("Get Employee")
		// Thêm dữ liệu vào map
		for key, value := range jsonData {
			dataMap[key] = value
			//Print data in map[]interface{}
			// fmt.Println(key, ": ", dataMap[key])
		}

		GetUsername := strings.TrimSpace(dataMap["username"].(string))
		GetAccountName := strings.TrimSpace(dataMap["account_name"].(string))
		GetAccount_Authorization_ID := strings.TrimSpace(dataMap["account_authorization_id"].(string))
		GetProfile_Image := strings.TrimSpace(dataMap["profile_image"].(string))
		GetFullName := strings.TrimSpace(dataMap["fullname"].(string))
		GetOfficeString := strings.TrimSpace(dataMap["office"].(string))
		GetDepartmentString := strings.TrimSpace(dataMap["department"].(string))
		GetPositionString := strings.TrimSpace(dataMap["position"].(string))
		GetManagerId := strings.TrimSpace(dataMap["manager_id"].(string))
		GetEmail := strings.TrimSpace(dataMap["email"].(string))
		GetPhone := strings.TrimSpace(dataMap["phone"].(string))
		GetGenderString := strings.TrimSpace(dataMap["gender"].(string))
		GetAddress := strings.TrimSpace(dataMap["address"].(string))
		var GetOffice int
		var GetDepartment int
		var GetPosition int
		var GetManager_ID primitive.ObjectID
		if GetManagerId != "" {
			GetManagerObjectId, errorManagerObjectID := primitive.ObjectIDFromHex(GetManagerId)
			if errorManagerObjectID != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error get manager ",
				})
				return
			}
			GetManager_ID = GetManagerObjectId
		} else {
			GetManager_ID = primitive.NilObjectID
		}
		if GetOfficeString == "" {
			GetOffice = -1
		} else {
			GetOffice, _ = strconv.Atoi(GetOfficeString)
		}
		if GetDepartmentString == "" {
			GetDepartment = -1
		} else {
			GetDepartment, _ = strconv.Atoi(GetDepartmentString)
		}
		if GetPositionString == "" {
			GetPosition = -1
		} else {
			GetPosition, _ = strconv.Atoi(GetPositionString)
		}
		GetGender, _ := strconv.Atoi(GetGenderString)

		var account_authorization_objectid, errorAccountObjectID = primitive.ObjectIDFromHex(GetAccount_Authorization_ID)
		if errorAccountObjectID != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error get authorization ",
			})
			return
		}

		var accountModel model.Account
		FindUsername := accountCollection.FindOne(
			ctx, bson.M{"username": GetUsername}).Decode(&accountModel)
		if FindUsername != nil {
			var password, unhashedPassword = GenerateAndHashPassword()

			if password == "" {
				c.JSON(http.StatusInternalServerError, "Error generating password")
				return
			}

			newAccount := model.Account{
				Id:                       primitive.NewObjectID(),
				Username:                 GetUsername,
				Password:                 password,
				Account_Name:             GetAccountName,
				Account_Authorization_Id: account_authorization_objectid,
				CreatedAt:                time.Now().Unix(),
				UpdatedAt:                time.Now().Unix(),
			}

			//insert the account into the database
			_, insertErr := accountCollection.InsertOne(ctx, newAccount)
			accountID, _ := primitive.ObjectIDFromHex(newAccount.Id.Hex())
			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, "Error inserting account"+insertErr.Error())
				return
			}

			newUserInfor := model.UserInfor{
				Id:            primitive.NewObjectID(),
				Profile_Image: GetProfile_Image,
				FullName:      GetFullName,
				Office:        GetOffice,
				Department:    GetDepartment,
				Position:      GetPosition,
				Manager_ID:    GetManager_ID,
				Phone:         GetPhone,
				Email:         GetEmail,
				Gender:        GetGender,
				Address:       GetAddress,
				CreatedAt:     time.Now().Unix(),
				UpdatedAt:     time.Now().Unix(),
			}

			_, userInforInsertError := userInforCollection.InsertOne(ctx, newUserInfor)
			if userInforInsertError != nil {
				// c.JSON(http.StatusInternalServerError, "Error inserting user information :"+insertErr.Error())

				fmt.Println("Deleting account is in progress")

				//Delete account just created if have an error
				filter := bson.D{{Key: "_id", Value: accountID}}
				if DeleteAccountResult, errDeleteAccountResult := accountCollection.DeleteOne(ctx, filter); errDeleteAccountResult != nil {
					c.JSON(http.StatusInternalServerError, "Error deleting account: "+errDeleteAccountResult.Error())
					return
				} else {
					c.JSON(http.StatusInternalServerError, "Have error create new employee: "+userInforInsertError.Error())
					fmt.Println("Delete account success: ", DeleteAccountResult)
				}

				return
			} else {
				UserInforID, _ := primitive.ObjectIDFromHex(newUserInfor.Id.Hex())
				newEmployee := model.Employee{
					Id:          primitive.NewObjectID(),
					State:       0,
					AccountID:   accountID,
					UserInforId: UserInforID,
					CreatedAt:   time.Now().Unix(),
					UpdatedAt:   time.Now().Unix(),
				}

				_, insertEmployeeErr := employeeCollection.InsertOne(ctx, newEmployee)
				if insertEmployeeErr != nil {
					// c.JSON(http.StatusInternalServerError, "error occured while create employee: "+insertEmployeeErr.Error())

					fmt.Println("Deleting account is in progress")

					//Delete account just created if have an error
					filterAccount1 := bson.D{{Key: "_id", Value: accountID}}
					if DeleteAccountResult1, err1 := accountCollection.DeleteOne(ctx, filterAccount1); err1 != nil {
						c.JSON(http.StatusInternalServerError, "Error deleting account: "+err1.Error())
						return
					} else {
						fmt.Println("Delete account success: ", DeleteAccountResult1)

					}

					filterUserInfor := bson.D{{Key: "_id", Value: UserInforID}}
					if DeleteUserInforResult, errDeleteUserInforResult := userInforCollection.DeleteOne(ctx, filterUserInfor); errDeleteUserInforResult != nil {
						c.JSON(http.StatusInternalServerError, "Error deleting user infor: "+errDeleteUserInforResult.Error())
						return
					} else {
						c.JSON(http.StatusInternalServerError, "Error creating new employee: "+insertEmployeeErr.Error())
						fmt.Println("Delete account success: ", DeleteUserInforResult)

					}

					return
				} else {
					fmt.Println("Insert employee Successful into Database with employee ID:", newEmployee.Id)
					// Send email to user
					sendEmailSuccess := SendNewUserEmail(GetEmail, GetFullName, GetUsername, unhashedPassword)

					if !sendEmailSuccess {
						c.JSON(http.StatusInternalServerError, "Error sending emails")
						return
					}

					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"message": "Employee account created",
					})
				}
			}
		} else {
			fmt.Println("Username exist!!!")
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Username exist",
				"success": false,
			})
		}
	}
}

func UpdateEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()
		var dataMap = make(map[string]interface{})

		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Parse JSON from request body
		var jsonData map[string]interface{}
		if err := json.Unmarshal(requestBody, &jsonData); err != nil {
			c.JSON(http.StatusBadRequest, "Failed to parse JSON")
			return
		}

		// Append data to map
		for key, value := range jsonData {
			dataMap[key] = value
		}

		userInforString := dataMap["userInforId"].(string)
		accountString := dataMap["accountId"].(string)
		employeeString := dataMap["employeeId"].(string)
		username := dataMap["username"].(string)
		accountName := dataMap["account_name"].(string)
		accountAuthorizationString := dataMap["account_authorization_id"].(string)
		fullname := dataMap["fullname"].(string)
		officeString := dataMap["office"].(string)
		departmentString := dataMap["department"].(string)
		positionString := dataMap["position"].(string)
		stateString := dataMap["state"].(string)
		managerString := dataMap["managerId"].(string)
		email := dataMap["email"].(string)
		genderString := dataMap["gender"].(string)
		phone := dataMap["phone"].(string)
		address := dataMap["address"].(string)
		var office, department, position int
		if officeString == "" {
			office = -1
		} else {
			office, _ = strconv.Atoi(officeString)
		}
		if departmentString == "" {
			department = -1
		} else {
			department, _ = strconv.Atoi(departmentString)
		}
		if positionString == "" {
			position = -1
		} else {
			position, _ = strconv.Atoi(positionString)
		}
		state, _ := strconv.Atoi(stateString)
		gender, _ := strconv.Atoi(genderString)

		userInforId, convertErr := primitive.ObjectIDFromHex(userInforString)
		if convertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error converting user infor ID")
			return
		}
		accountId, convertErr := primitive.ObjectIDFromHex(accountString)
		if convertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error converting account ID")
			return
		}
		employeeId, convertErr := primitive.ObjectIDFromHex(employeeString)
		if convertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error converting employee ID")
			return
		}
		var managerId primitive.ObjectID
		if managerString != "" {
			managerId, convertErr = primitive.ObjectIDFromHex(managerString)
			if convertErr != nil {
				c.JSON(http.StatusInternalServerError, "Error converting manager ID")
				return
			}
		} else {
			managerId = primitive.NilObjectID
		}
		accountAuthorizationId, convertErr := primitive.ObjectIDFromHex(accountAuthorizationString)
		if convertErr != nil {
			c.JSON(http.StatusInternalServerError, "Error converting account authorization ID")
			return
		}

		accountUpdateResult := accountCollection.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "_id", Value: accountId}},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "username", Value: username},
					{Key: "account_name", Value: accountName},
					{Key: "account_authorization_id", Value: accountAuthorizationId},
					{Key: "updatedAt", Value: time.Now().Unix()},
				}},
			},
		)
		if accountUpdateResult.Err() != nil {
			c.JSON(http.StatusInternalServerError, "Error updating account: "+accountUpdateResult.Err().Error())
			return
		}

		userinforUpdateResult := userInforCollection.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "_id", Value: userInforId}},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "office", Value: office},
					{Key: "department", Value: department},
					{Key: "position", Value: position},
					{Key: "manager_id", Value: managerId},
					{Key: "phone", Value: phone},
					{Key: "email", Value: email},
					{Key: "gender", Value: gender},
					{Key: "address", Value: address},
					{Key: "fullname", Value: fullname},
					{Key: "updatedAt", Value: time.Now().Unix()},
				}},
			},
		)
		if userinforUpdateResult.Err() != nil {
			c.JSON(http.StatusInternalServerError, "Error updating user infor: "+userinforUpdateResult.Err().Error())
			return
		}

		employeeUpdateResult := employeeCollection.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "_id", Value: employeeId}},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "state", Value: state},
					{Key: "updatedAt", Value: time.Now().Unix()},
				}},
			},
		)
		if employeeUpdateResult.Err() != nil {
			c.JSON(http.StatusInternalServerError, "Error updating employee: "+employeeUpdateResult.Err().Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Employee account updated",
		})
	}
}

func GetEmployeeIDByFullname() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user_infor model.UserInfor
		//validate the request body
		if err := c.BindJSON(&user_infor); err != nil {
			c.JSON(http.StatusBadRequest, "Request error"+err.Error())
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&user_infor); validationErr != nil {
			c.JSON(http.StatusBadRequest, "Error, missing field"+validationErr.Error())
			return
		}

		fmt.Println("Get full name:", user_infor.FullName)
	}
}

func EmployeeGetAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Epic model
		var employees []gin.H

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			//Join employee table with userinfor table
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
					{Key: "from", Value: "accounts"},
					{Key: "localField", Value: "account_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "account"},
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
				{Key: "$sort", Value: bson.D{
					{Key: "userinfor.fullname", Value: 1},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "employee_id", Value: "$_id"},
					{Key: "account_id", Value: "$account_id"},
					{Key: "user_infor_id", Value: "$userinfor_id"},
					{Key: "profile_image", Value: "$userinfor.profile_image"},
					{Key: "state", Value: "$state"},
					{Key: "fullname", Value: "$userinfor.fullname"},
					{Key: "office", Value: "$userinfor.office"},
					{Key: "department", Value: "$userinfor.department"},
					{Key: "position", Value: "$userinfor.position"},
					{Key: "email", Value: "$userinfor.email"},
					{Key: "phone", Value: "$userinfor.phone"},
					{Key: "gender", Value: "$userinfor.gender"},
					{Key: "address", Value: "$userinfor.address"},
					{Key: "authorization_level", Value: "$authorization.levelName"},
				}},
			},
		}

		// Use the $lookup stage to aggregate data from the Epic and Project collections
		result, aggregateErr := employeeCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &employees)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"employees": employees,
		})
	}
}

func EmployeeSearchFullName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		var getfullname model.UserInfor
		//validate the request body
		if err := c.BindJSON(&getfullname); err != nil {
			c.JSON(http.StatusBadRequest, "Request error"+err.Error())
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&getfullname); validationErr != nil {
			c.JSON(http.StatusBadRequest, "Error, missing field"+validationErr.Error())
			return
		}

		var fullname = getfullname.FullName
		fmt.Println("Get full name:", fullname)

		// Create an array of the Epic model
		var employees []gin.H

		// Define pipeline to join collections and sort the result
		pipeline := mongo.Pipeline{
			//Join employee table with userinfor table
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "userinfor"},
				}},
			},
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "userinfor.fullname", Value: bson.D{
						{Key: "$regex", Value: fullname},
						{Key: "$options", Value: "i"},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "userinfor.fullname", Value: 1},
				}},
			},
			//The project variable is used to combine all values
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "employee_id", Value: "$_id"},
					{Key: "account_id", Value: "$account_id"},
					{Key: "user_infor_id", Value: "$userinfor_id"},
					{Key: "state", Value: "$state"},
					{Key: "fullname", Value: "$userinfor.fullname"},
					{Key: "email", Value: "$userinfor.email"},
					{Key: "profile_image", Value: "$userinfor.profile_image"},
					{Key: "position", Value: "$userinfor.position"},
					{Key: "office", Value: "$userinfor.office"},
					{Key: "department", Value: "$userinfor.department"},
				}},
			},
		}

		// Use the $lookup stage to aggregate data from the Epic and Project collections
		result, aggregateErr := employeeCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating epics: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &employees)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding epics: "+decodeErr.Error())
			return
		}
		if employees != nil {
			// Send response to client
			c.JSON(http.StatusOK, gin.H{
				"count":     len(employees),
				"employees": employees,
			})
		}
	}
}

func EmployeeUpdateState() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		//Get employee from post api
		var requestData map[string]interface{}
		if err := c.BindJSON(&requestData); err != nil {
			// If there's an error parsing the request body, return an error response
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var state = requestData["state"].(string)
		var employee_id = requestData["employee_id"].(string)

		//convert to primitive object ID
		id, _ := primitive.ObjectIDFromHex(employee_id)

		filter := bson.D{{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "state", Value: state},
		}}} // "password", "asdfadfafs",

		var employeeUpdate model.Employee
		if err1 := employeeCollection.FindOneAndUpdate(ctx, filter, update).Decode(&employeeUpdate); err1 != nil {
			c.JSON(http.StatusInternalServerError, "Error updating project"+err1.Error())
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Updated employee successfully",
			})
		}

	}
}

func GetEmployeeByRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Epic model
		var employees []gin.H

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
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
					{Key: "from", Value: "authorizations"},
					{Key: "localField", Value: "account.account_authorization_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "authorization"},
				}},
			},
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "authorization.levelName", Value: c.Param("role")},
				}},
			},
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "fullname", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$user_info.fullname", 0}},
					}},
					{Key: "office", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$user_info.office", 0}},
					}},
					{Key: "department", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$user_info.department", 0}},
					}},
					{Key: "email", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$user_info.email", 0}},
					}},
					{Key: "profile_image", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$user_info.profile_image", 0}},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "fullname", Value: 1},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Account, Authorization, and User_Infor collections
		result, aggregateErr := employeeCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating employees: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &employees)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding employees: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"employees": employees,
		})
	}
}

func GetEmployeeDetailWithID() gin.HandlerFunc {
	return func(c *gin.Context) {
		//get id from link
		var GetEmployeeID = c.Param("id")

		//convert to primary object for ID
		EmployeeIDPrimaryObject, _ := primitive.ObjectIDFromHex(GetEmployeeID) //

		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Create an array of the Epic model
		var employee []gin.H

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "_id", Value: EmployeeIDPrimaryObject},
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
					{Key: "from", Value: "employee"},
					{Key: "localField", Value: "userinfor.manager_id"},
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
			//Only get 3 table
			bson.D{
				{Key: "$project", Value: bson.D{
					// {Key: "userinfor_id", Value: "$userinfor_id"},
					{Key: "account", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$account", 0}},
					}},
					{Key: "manager_userinfor", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$manager_userinfor", 0}},
					}},
					{Key: "userinfor", Value: bson.D{
						{Key: "$arrayElemAt", Value: bson.A{"$userinfor", 0}},
					}},
					{Key: "state", Value: 1},
					{Key: "createdAt", Value: 1},
					{Key: "updatedAt", Value: 1},
				}},
			},
		}

		// Use the defined stages to aggregate data from the Account, Authorization, and User_Infor collections
		result, aggregateErr := employeeCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating employees: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &employee)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding employees: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"employee": employee[0],
		})
	}
}

func GetEmployeeByManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("manager"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Manager not found",
			})
			return
		}

		// Create an array for the employees
		var employees []gin.H

		// Define a pipeline to filter the data by title and join collections
		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$lookup", Value: bson.D{
					{Key: "from", Value: "user_infor"},
					{Key: "localField", Value: "userinfor_id"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user_info"},
				}},
			},
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "user_info.manager_id", Value: queryId},
				}},
			},
		}

		// Use the defined stages to aggregate data from User_Infor collection
		result, aggregateErr := employeeCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating employees: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &employees)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding employees: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"employees": employees,
		})
	}
}

func GetEmployeeByProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutLimit)
		defer cancel()

		// Convert the hex string to ObjectID
		queryId, convertErr := primitive.ObjectIDFromHex(c.Param("project"))
		if convertErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Project Id invalid",
			})
			return
		}

		// Create an array for the employees
		var employees []gin.H

		// Define a pipeline to filter the data by title and join collections
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
					{Key: "from", Value: "projects"},
					{Key: "localField", Value: "epic.project"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "project"},
				}},
			},
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "project._id", Value: queryId},
				}},
			},
			bson.D{
				{Key: "$unwind", Value: "$members"},
			},
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "members", Value: 1},
				}},
			},
		}

		// Use the defined stages to aggregate data from User_Infor collection
		result, aggregateErr := taskCollection.Aggregate(ctx, pipeline)
		if aggregateErr != nil {
			c.JSON(http.StatusInternalServerError, "Error aggregating employees: "+aggregateErr.Error())
			return
		}

		// Decode the data from DB to the epics array
		decodeErr := result.All(ctx, &employees)
		if decodeErr != nil {
			c.JSON(http.StatusInternalServerError, "Error decoding employees: "+decodeErr.Error())
			return
		}

		// Close the cursor after getting data to prevent memory leak
		result.Close(ctx)

		// Send response to client
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"employees": employees,
		})
	}
}
