package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Post      primitive.ObjectID `json:"post" bson:"post"`
	Author    primitive.ObjectID `json:"author" bson:"author"`
	Text      string             `json:"text" bson:"text"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}

type CommentWithAuthor struct {
	Comment
	AuthorData *User `json:"authorData,omitempty" bson:"authorData,omitempty"`
}