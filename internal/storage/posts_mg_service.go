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

func (p *PostMongoStorage) DeletePost(ctx context.Context, postID bson.ObjectID, author string) error {
	var post models.Post
	err := p.collPosts.FindOne(ctx, bson.M{"_id": postID, "person": author}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return mongo.ErrNoDocuments
		}
		return err
	}

	_, err = p.collComm.DeleteMany(ctx, bson.M{"postId": postID})
	if err != nil {
		return err
	}
	_, err = p.collPosts.DeleteOne(ctx, bson.M{"_id": postID})
	if err != nil {
		return err
	}

	return nil
}

func (p *PostMongoStorage) DeleteComment(ctx context.Context, commentID bson.ObjectID, author string) error {
	var comment models.Comment
	err := p.collComm.FindOne(ctx, bson.M{"_id": commentID, "person": author}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return mongo.ErrNoDocuments
		}
		return err
	}

	_, err = p.collComm.DeleteOne(ctx, bson.M{"_id": commentID})
	if err != nil {
		return err
	}

	_, err = p.collPosts.UpdateOne(
		ctx,
		bson.M{"_id": comment.PostID},
		bson.M{"$pull": bson.M{"commentIds": commentID}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostMongoStorage) UpdatePost(
	ctx context.Context,
	postID bson.ObjectID,
	author string,
	title string,
	content string,
) error {

	var existingPost models.Post
	err := p.collPosts.FindOne(ctx, bson.M{"_id": postID}).Decode(&existingPost)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return mongo.ErrNoDocuments
		}

		return err
	}

	if existingPost.Person != author {
		return mongo.ErrNoDocuments
	}

	_, err = p.collPosts.UpdateOne(
		ctx,
		bson.M{"_id": postID, "person": author},
		bson.M{"$set": bson.M{
			"heading":   title,
			"mainText":  content,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostMongoStorage) UpdateComment(
	ctx context.Context,
	commentID bson.ObjectID,
	author string,
	content string,
) error {
	res, err := p.collComm.UpdateOne(
		ctx,
		bson.M{"_id": commentID, "person": author},
		bson.M{"$set": bson.M{"mainText": content}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (p *PostMongoStorage) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p.client.Disconnect(ctx)
}
