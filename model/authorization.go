package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Authorization struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	LevelName   string             `bson:"levelName,omitempty" validate:"required"`
	Description string             `bson:"description"`
	CreatedAt   int64              `bson:"createdAt"`
	UpdatedAt   int64              `bson:"updatedAt"`
}
