package storage

import (
	"context"
	"crypto-analytics/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

func (p *PostMongoStorage) createPost(ctx context.Context, person, heading, mainText, date string) (bson.ObjectID, error) {
	id := bson.NewObjectID() // генерируем ID вручную (опционально)
	post := models.Post{
		ID:       id,
		Person:   person,
		Heading:  heading,
		MainText: mainText,
		Date:     date,
		Comments: nil, // комментарии будут в отдельной коллекции
	}
	_, err := p.collPosts.InsertOne(ctx, post)
	if err != nil {
		return bson.ObjectID{}, err
	}
	return id, nil
}
func (p *PostMongoStorage) createComment(ctx context.Context, person, mainText, date string, postID bson.ObjectID) error {
	comment := models.Comment{
		ID:       bson.NewObjectID(),
		Person:   person,
		MainText: mainText,
		Date:     date,
		PostID:   postID, // ссылка на пост
	}

	_, err := p.collComm.InsertOne(ctx, comment)
	return err
}

func (p *PostMongoStorage) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p.client.Disconnect(ctx)
}
