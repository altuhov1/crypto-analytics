package services

import "crypto-analytics/internal/storage"

type PostsService struct {
	postStorage storage.PostStorage
}

func NewPostService(ps storage.PostStorage) *PostsService {
	return &PostsService{
		postStorage: ps,
	}
}
