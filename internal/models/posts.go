package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Post структура для постов
type Post struct {
	ID         bson.ObjectID   `bson:"_id,omitempty"`
	Person     string          `bson:"person"`
	Heading    string          `bson:"heading"`
	MainText   string          `bson:"mainText"`
	Date       string          `bson:"date"`
	CommentIDs []bson.ObjectID `bson:"commentIds,omitempty"`
}

// Comment структура для комментариев
type Comment struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Person   string        `bson:"person"`
	MainText string        `bson:"mainText"`
	Date     string        `bson:"date"`
	PostID   bson.ObjectID `bson:"postId,omitempty"`
}
