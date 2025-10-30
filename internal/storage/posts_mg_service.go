package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PostMongoStorage struct {
	client *mongo.Client
}

func NewPostsMongoStorage(client *mongo.Client) *PostMongoStorage {

	return &PostMongoStorage{client: client}
}
func (p *PostMongoStorage) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p.client.Disconnect(ctx)
}
