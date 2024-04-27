package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Employee struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	State       int                `bson:"state"`
	AccountID   primitive.ObjectID `bson:"account_id"`
	UserInforId primitive.ObjectID `bson:"userinfor_id"`
	CreatedAt   int64              `bson:"createdAt"`
	UpdatedAt   int64              `bson:"updatedAt"`
}
