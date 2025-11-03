package services

import (
	"context"
	"errors"
	"fmt"

	"unicode/utf8"

	"crypto-analytics/internal/models"
	"crypto-analytics/internal/storage"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PostsService struct {
	postStorage storage.PostStorage
}

func NewPostService(ps storage.PostStorage) *PostsService {
	return &PostsService{
		postStorage: ps,
	}
}

func (s *PostsService) CreatePost(ctx context.Context, post models.Post) (bson.ObjectID, error) {
	if err := s.validatePost(post); err != nil {
		return bson.ObjectID{}, err
	}
	return s.postStorage.CreatePost(ctx, post)
}
func (s *PostsService) CreateComment(ctx context.Context, comment models.Comment) error {
	if err := s.validateComment(comment); err != nil {
		return err
	}
	return s.postStorage.CreateComment(ctx, comment)
}

func (s *PostsService) GetLastPosts(ctx context.Context) ([]models.Post, error) {
	return s.postStorage.GetLastPosts(ctx)
}

func (s *PostsService) GetLastCommentsByPost(ctx context.Context, postID bson.ObjectID) ([]models.Comment, error) {
	return s.postStorage.GetLastCommentsByPost(ctx, postID)
}

func (s *PostsService) validatePost(post models.Post) error {
	if post.Person == "" {
		return ErrEmptyPerson
	}
	if post.Heading == "" {
		return ErrEmptyHeading
	}
	if post.MainText == "" {
		return ErrEmptyMainText
	}
	if post.Date == "" {
		return ErrEmptyDate
	}

	if utf8.RuneCountInString(post.Person) > 100 {
		return ErrPersonTooLong
	}
	if utf8.RuneCountInString(post.Heading) > 200 {
		return ErrHeadingTooLong
	}
	if utf8.RuneCountInString(post.MainText) > 5000 {
		return ErrMainTextTooLong
	}

	return nil
}

// validateComment валидирует структуру комментария
func (s *PostsService) validateComment(comment models.Comment) error {
	if comment.Person == "" {
		return ErrEmptyPerson
	}
	if comment.MainText == "" {
		return ErrEmptyMainText
	}
	if comment.Date == "" {
		return ErrEmptyDate
	}
	if comment.PostID.IsZero() {
		return ErrInvalidPostID
	}

	if utf8.RuneCountInString(comment.Person) > 100 {
		return ErrPersonTooLong
	}
	if utf8.RuneCountInString(comment.MainText) > 1000 {
		return ErrCommentTooLong
	}

	return nil
}

func (s *PostsService) DeletePost(ctx context.Context, postID bson.ObjectID, author string) error {
	if author == "" {
		return ErrEmptyPerson
	}
	return s.postStorage.DeletePost(ctx, postID, author)
}

// DeleteComment удаляет комментарий, если автор совпадает
func (s *PostsService) DeleteComment(ctx context.Context, commentID bson.ObjectID, author string) error {
	if author == "" {
		return ErrEmptyPerson
	}
	return s.postStorage.DeleteComment(ctx, commentID, author)
}

// UpdatePost обновляет заголовок и основной текст поста с валидацией
func (s *PostsService) UpdatePost(
	ctx context.Context,
	postID bson.ObjectID,
	author string,
	title string,
	content string,
) error {
	if author == "" {
		return fmt.Errorf("%w: author is required", ErrEmptyPerson)
	}
	if title == "" {
		return fmt.Errorf("%w: title is required", ErrEmptyHeading)
	}
	if content == "" {
		return fmt.Errorf("%w: content is required", ErrEmptyMainText)
	}
	if utf8.RuneCountInString(title) > 200 {
		return fmt.Errorf("%w: title exceeds 200 characters", ErrHeadingTooLong)
	}
	if utf8.RuneCountInString(content) > 5000 {
		return fmt.Errorf("%w: content exceeds 5000 characters", ErrMainTextTooLong)
	}

	err := s.postStorage.UpdatePost(ctx, postID, author, title, content)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("post not found or you don't have permission to edit it")
		}
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func (s *PostsService) UpdateComment(
	ctx context.Context,
	commentID bson.ObjectID,
	author string,
	content string,
) error {
	if author == "" {
		return ErrEmptyPerson
	}
	if content == "" {
		return ErrEmptyMainText
	}
	if utf8.RuneCountInString(content) > 1000 {
		return ErrCommentTooLong
	}

	return s.postStorage.UpdateComment(ctx, commentID, author, content)
}

var (
	ErrEmptyPerson     = errors.New("person cannot be empty")
	ErrEmptyHeading    = errors.New("heading cannot be empty")
	ErrEmptyMainText   = errors.New("main text cannot be empty")
	ErrEmptyDate       = errors.New("date cannot be empty")
	ErrInvalidPostID   = errors.New("invalid post ID")
	ErrPersonTooLong   = errors.New("person name too long (max 100 characters)")
	ErrHeadingTooLong  = errors.New("heading too long (max 200 characters)")
	ErrMainTextTooLong = errors.New("main text too long (max 5000 characters)")
	ErrCommentTooLong  = errors.New("comment too long (max 1000 characters)")
)
