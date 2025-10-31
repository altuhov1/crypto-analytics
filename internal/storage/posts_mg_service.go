package storage

import (
	"context"
	"crypto-analytics/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DBName           = "cryptodb"
	PostsCollName    = "posts"
	CommentsCollName = "comments"
)

type PostMongoStorage struct {
	client    *mongo.Client
	collPosts *mongo.Collection
	collComm  *mongo.Collection
}

func NewPostsMongoStorage(client *mongo.Client) *PostMongoStorage {

	return &PostMongoStorage{
		client:    client,
		collPosts: client.Database(DBName).Collection(PostsCollName),
		collComm:  client.Database(DBName).Collection(CommentsCollName),
	}
}

func (p *PostMongoStorage) CreatePost(ctx context.Context, post models.Post) (bson.ObjectID, error) {
	post.ID = bson.NewObjectID()
	_, err := p.collPosts.InsertOne(ctx, post)
	if err != nil {
		return bson.ObjectID{}, err
	}
	return post.ID, nil
}
func (p *PostMongoStorage) CreateComment(
	ctx context.Context,
	comment models.Comment,
) error {
	comment.ID = bson.NewObjectID()

	_, err := p.collComm.InsertOne(ctx, comment)
	if err != nil {
		return err
	}

	_, err = p.collPosts.UpdateOne(
		ctx,
		bson.M{"_id": comment.PostID},
		bson.M{"$push": bson.M{"commentIds": comment.ID}},
	)
	if err != nil {
		return err
	}

	return nil
}
func (p *PostMongoStorage) GetLastPosts(ctx context.Context) ([]models.Post, error) {
	cursor, err := p.collPosts.Find(
		ctx,
		bson.D{},
		options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(100),
	)
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

func (p *PostMongoStorage) GetLastCommentsByPost(ctx context.Context, postID bson.ObjectID) ([]models.Comment, error) {
	filter := bson.M{"postId": postID}

	cursor, err := p.collComm.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(100),
	)
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
func (p *PostMongoStorage) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p.client.Disconnect(ctx)
}
