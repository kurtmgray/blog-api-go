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

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	FindAll(ctx context.Context) ([]models.Post, error)
	FindAllWithAuthor(ctx context.Context) ([]models.PostWithAuthor, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Post, error)
	FindByIDWithAuthor(ctx context.Context, id primitive.ObjectID) (*models.PostWithAuthor, error)
	FindByAuthor(ctx context.Context, author primitive.ObjectID) ([]models.Post, error)
	Update(ctx context.Context, id primitive.ObjectID, update interface{}) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type postRepository struct {
	collection *mongo.Collection
}

func NewPostRepository(db *mongo.Database) PostRepository {
	return &postRepository{
		collection: db.Collection("posts"),
	}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	post.ID = primitive.NewObjectID()
	post.Timestamp = time.Now()

	_, err := r.collection.InsertOne(ctx, post)
	return err
}

func (r *postRepository) FindAll(ctx context.Context) ([]models.Post, error) {
	// cursor for multiple results
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	// close to free up resources
	defer cursor.Close(ctx)

	// call .All() to decode into posts
	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) FindAllWithAuthor(ctx context.Context) ([]models.PostWithAuthor, error) {
	pipeline := mongo.Pipeline{
		// lookup users collection
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "author"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "authorData"},
		}}},
		// unwind the author array (converts [author] to author)
		{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$authorData"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		// project only needed user fields
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$toString", Value: "$_id"}}},
			{Key: "title", Value: 1},
			{Key: "text", Value: 1},
			{Key: "imgUrl", Value: 1},
			{Key: "published", Value: 1},
			{Key: "timestamp", Value: 1},
			{Key: "author", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$toString", Value: "$authorData._id"}}},
				{Key: "username", Value: "$authorData.username"},
				{Key: "fname", Value: "$authorData.fname"},
				{Key: "lname", Value: "$authorData.lname"},
				{Key: "admin", Value: "$authorData.admin"},
				{Key: "canPublish", Value: "$authorData.canPublish"},
			}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.PostWithAuthor
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *postRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Post, error) {
	var post models.Post
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
	}
	return &post, err
}

func (r *postRepository) FindByIDWithAuthor(ctx context.Context, id primitive.ObjectID) (*models.PostWithAuthor, error) {
	pipeline := mongo.Pipeline{
		// match the specific post
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}},
		// lookup author
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "author"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "authorData"},
		}}},
		// unwind author array
		{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$authorData"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		// project only needed fields
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$toString", Value: "$_id"}}},
			{Key: "title", Value: 1},
			{Key: "text", Value: 1},
			{Key: "imgUrl", Value: 1},
			{Key: "published", Value: 1},
			{Key: "timestamp", Value: 1},
			{Key: "author", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$toString", Value: "$authorData._id"}}},
				{Key: "username", Value: "$authorData.username"},
				{Key: "fname", Value: "$authorData.fname"},
				{Key: "lname", Value: "$authorData.lname"},
				{Key: "admin", Value: "$authorData.admin"},
				{Key: "canPublish", Value: "$authorData.canPublish"},
			}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.PostWithAuthor
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, errors.New("post not found")
	}

	return &posts[0], nil
}

func (r *postRepository) FindByAuthor(ctx context.Context, authorID primitive.ObjectID) ([]models.Post, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"author": authorID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, id primitive.ObjectID, update interface{}) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("post not found")
	}

	return nil
}
