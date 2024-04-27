package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // No update
	Leader      primitive.ObjectID `json:"leader,omitempty" bson:"leader,omitempty" validate:"required"`
	Title       string             `json:"title,omitempty" bson:"title,omitempty" validate:"required"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt"` // No update
	UpdatedAt   time.Time          `bson:"updatedAt"`
}

// [Project] ->> Epic ->> Task
