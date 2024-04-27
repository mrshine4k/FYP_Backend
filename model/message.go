package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Id           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Sender       primitive.ObjectID `json:"sender,omitempty" bson:"sender,omitempty"`
	Message      string             `json:"message,omitempty" bson:"message,omitempty"`
	Files        []string           `json:"files,omitempty" bson:"files,omitempty"`
	InitialFiles []string           `json:"initialFiles,omitempty" bson:"initialFiles,omitempty"`
	Project      primitive.ObjectID `json:"project,omitempty" bson:"project,omitempty"`
	CreatedAt    int64              `bson:"createdAt"`
	UpdatedAt    int64              `bson:"updatedAt"`
}
