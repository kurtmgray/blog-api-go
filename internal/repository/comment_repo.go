package repository

import (
	"context"
	"errors"
	"time"

	"github.com/kurtgray/blog-api-go/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	FindByPost(ctx context.Context, postID primitive.ObjectID) ([]models.Comment, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Comment, error)
	Update(ctx context.Context, id primitive.ObjectID, text string) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type commentRepository struct {
	collection *mongo.Collection
}

func NewCommentRepository(db *mongo.Database) CommentRepository {
	return &commentRepository{
		collection: db.Collection("comments"),
	}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	comment.ID = primitive.NewObjectID()
	comment.Timestamp = time.Now()

	_, err := r.collection.InsertOne(ctx, comment)
	return err
}

func (r *commentRepository) FindByPost(ctx context.Context, postID primitive.ObjectID) ([]models.Comment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"post": postID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *commentRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Comment, error) {
	var comment models.Comment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&comment)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}

	return &comment, nil
}

func (r *commentRepository) Update(ctx context.Context, id primitive.ObjectID, text string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"text": text}},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("comment not found")
	}

	return nil
}
