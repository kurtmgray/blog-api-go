package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	GoogleID   string               `json:"googleId,omitempty" bson:"googleId,omitempty"`
	Username   string               `json:"username" bson:"username"`
	Password   string               `json:"password,omitempty" bson:"password,omitempty"`
	Fname      string               `json:"fname" bson:"fname"`
	Lname      string               `json:"lname" bson:"lname"`
	Admin      bool                 `json:"admin" bson:"admin"`
	CanPublish bool                 `json:"canPublish" bson:"canPublish"`
	Posts      []primitive.ObjectID `json:"posts" bson:"posts"`
	Comments   []primitive.ObjectID `json:"comments" bson:"comments"`
	CreatedAt  time.Time            `json:"createdAt" bson:"createdAt"`
}

type UserResponse struct {
	ID         string    `json:"_id" bson:"_id"`
	Username   string    `json:"username"`
	Fname      string    `json:"fname"`
	Lname      string    `json:"lname"`
	Admin      bool      `json:"admin"`
	CanPublish bool      `json:"canPublish"`
	CreatedAt  time.Time `json:"createdAt"`
}
