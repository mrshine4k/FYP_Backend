package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	Id          primitive.ObjectID   `bson:"_id,omitempty"`
	Epic        primitive.ObjectID   `bson:"epic,omitempty" validate:"required"` // No update
	Members     []primitive.ObjectID `bson:"members,omitempty"`
	Status      string               `bson:"status,omitempty"`
	Title       string               `bson:"title,omitempty" validate:"required"`
	Description string               `bson:"description,omitempty"`
	Note        string               `bson:"note,omitempty"`
	Attachments []string             `bson:"attachments,omitempty"`
	CreatedAt   time.Time            `bson:"createdAt"` // No update
	UpdatedAt   time.Time            `bson:"updatedAt"`
}

// Project ->> Epic ->> [Task]
