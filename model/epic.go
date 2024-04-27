package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Epic struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`                                   // No update
	Project     primitive.ObjectID `json:"project,omitempty" bson:"project,omitempty" validate:"customrequired"` // No update
	Title       string             `json:"title,omitempty" bson:"title,omitempty" validate:"customrequired"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt"` // No update
	UpdatedAt   time.Time          `bson:"updatedAt"`
}

// Project ->> [Epic] ->> Task
