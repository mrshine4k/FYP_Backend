package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserInfor struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`

	FullName      string             `bson:"fullname,omitempty" validate:"required"`
	Profile_Image string             `bson:"profile_image"`
	Office        int                `bson:"office"`
	Department    int                `bson:"department"`
	Position      int                `bson:"position"`
	Manager_ID    primitive.ObjectID `bson:"manager_id"`
	Phone         string             `bson:"phone"`
	Email         string             `bson:"email"`
	Gender        int                `bson:"gender"`
	Address       string             `bson:"address"`
	CreatedAt     int64              `bson:"createdAt"`
	UpdatedAt     int64              `bson:"updatedAt"`
}
