package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Account struct {
	Id                       primitive.ObjectID `bson:"_id"`
	Username                 string             `bson:"username" json:"username"`
	Password                 string             `bson:"password" json:"password"`
	Account_Name             string             `bson:"account_name"`
	Account_Authorization_Id primitive.ObjectID `bson:"account_authorization_id"`
	CreatedAt                int64              `bson:"createdAt"`
	UpdatedAt                int64              `bson:"updatedAt"`
}
