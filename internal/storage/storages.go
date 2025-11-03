package storage

import (
	"context"
	"crypto-analytics/internal/models"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Storage определяет контракт для работы с данными
type FormStorage interface {
	SaveContactFrom(contact *models.ContactForm) error
	ExportContactsToJSON(filename string) error
	Close()
}

type UserStorage interface {
	CreateUser(user *models.User) error
	GetUserByName(nameU string) (*models.User, error)
	GetAllFavoriteCoins(nameU string) ([]string, error)
	NewFavoriteCoin(nameU string, nameCoin string) error
	RemoveFavoriteCoin(nameU string, nameCoin string) error
	ExportUsersToJSON(filename string) error
	Close()
}

type NewsStorage interface {
	AddNews([]models.NewsItem) error
	GetAllNews() ([]models.NewsItem, error)
	UpdateNews([]models.NewsItem) error
}

type CacheStorage interface {
	Save(data []byte, amountPairs int) error
	Load() ([]string, error)
}

type AnalysisStorage interface {
	SaveAnalysisData(data models.PairsCrypto) error
	LoadAnalysisData() (models.PairsCrypto, error)
}

type PostStorage interface {
	CreatePost(ctx context.Context, post models.Post) (bson.ObjectID, error)
	CreateComment(
		ctx context.Context,
		comment models.Comment,
	) error
	GetLastPosts(ctx context.Context) ([]models.Post, error)
	GetLastCommentsByPost(ctx context.Context,
		postID bson.ObjectID,
	) ([]models.Comment, error)
	DeletePost(ctx context.Context, postID bson.ObjectID, author string) error
	DeleteComment(ctx context.Context, commentID bson.ObjectID, author string) error
	UpdatePost(
		ctx context.Context,
		postID bson.ObjectID,
		author string,
		title string,
		content string,
	) error
	UpdateComment(
		ctx context.Context,
		commentID bson.ObjectID,
		author string,
		content string,
	) error
	Close()
}

type AnalysisTempStorage interface {
	SaveAnalysisData(data models.AnalysisData) error
	SavePairs(pairs models.PairsCrypto) error
	GetAnalysisData(pair,
		timeframe string,
	) (*models.AnalysisData, error)
	GetStats() string
	Close(client *redis.Client)
}
