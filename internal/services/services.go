package services

import (
	"context"
	"crypto-analytics/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AnalysisGService interface {
	GetPairInfo(pair, timeframe string) (*models.AnalysisData, error)
}

type GetAllPairsService interface {
	GetTopCryptos(limit int) ([]models.Coin, error)
	GetCacheInfo() (int, time.Time)
}

type NewsRssService interface {
	GetNews() ([]models.NewsItem, error)
	GetNewsCount() (int, error)
}

type Notifier interface {
	NotifyAdmContForm(contact *models.ContactForm)
	NotifyAdmNewUserForm(contact *models.User)
}

type AIAnalysisService interface {
	GetAllPairs() ([]string, error)
	GetPairsCount() int
}

type PostPService interface {
	CreatePost(ctx context.Context, post models.Post) (bson.ObjectID, error)
	CreateComment(ctx context.Context, comment models.Comment) error
	GetLastPosts(ctx context.Context) ([]models.Post, error)
	GetLastCommentsByPost(ctx context.Context, postID bson.ObjectID) ([]models.Comment, error)
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
}
type UserLogService interface {
	RegisterUser(user *models.User) error
	LoginUser(username, password string) error
	HashPassword(password string) (string, error)
	AddFavorite(username, CoinID string) error
	RemoveFavorite(username, CoinID string) error
	GetFavorites(username string) ([]string, error)
	PrintJsonAllUsers(fileName string) error
}
