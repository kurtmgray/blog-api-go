package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Author    primitive.ObjectID `json:"author" bson:"author"`
	Title     string             `json:"title" bson:"title"`
	Text      string             `json:"text" bson:"text"`
	ImgURL    string             `json:"imgUrl,omitempty" bson:"imgUrl,omitempty"`
	Published bool               `json:"published" bson:"published"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}

type PostWithAuthor struct {
	ID        string        `json:"_id" bson:"_id,omitempty"`
	Author    *UserResponse `json:"author,omitempty" bson:"author,omitempty"`
	Title     string        `json:"title" bson:"title"`
	Text      string        `json:"text" bson:"text"`
	ImgURL    string        `json:"imgUrl,omitempty" bson:"imgUrl,omitempty"`
	Published bool          `json:"published" bson:"published"`
	Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
}
