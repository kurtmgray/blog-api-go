package repository

import (
	"context"
	"time"

	"github.com/kurtgray/blog-api-go/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// interface for db operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*models.User, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

// receiver pattern like 'this'
// adding methods to the struct to satisfy the interface

// creates new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.Posts = []primitive.ObjectID{}
	user.Comments = []primitive.ObjectID{}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// finds user by id
func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	// declare
	var user models.User
	// decode into user var with pointer
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		// not an err, just no user found
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
	}
	// return user and error (one will be nil)
	return &user, err
}

// finds user by username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
	}
	return &user, err
}

// finds user by google id
func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"googleId": googleID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
	}
	return &user, err
}