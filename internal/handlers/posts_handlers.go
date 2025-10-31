package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"crypto-analytics/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CreatePostHandler создает новый пост
func (h *Handler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("CreatePostHandler started")

	var request struct {
		Person   string `json:"person"`
		Heading  string `json:"heading"`
		MainText string `json:"mainText"`
		Date     string `json:"date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	post := models.Post{
		Person:     request.Person,
		Heading:    request.Heading,
		MainText:   request.MainText,
		Date:       request.Date,
		CommentIDs: []bson.ObjectID{},
	}

	postID, err := h.postsService.CreatePost(r.Context(), post)
	if err != nil {
		slog.Error("Failed to create post", "error", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"postId":  postID.Hex(),
		"message": "Post created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Post created successfully", "postId", postID.Hex())
}

func (h *Handler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("CreateCommentHandler started")

	var request struct {
		Person   string `json:"person"`
		MainText string `json:"mainText"`
		Date     string `json:"date"`
		PostID   string `json:"postId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	postID, err := bson.ObjectIDFromHex(request.PostID)
	if err != nil {
		slog.Error("Invalid post ID", "postId", request.PostID, "error", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		Person:   request.Person,
		MainText: request.MainText,
		Date:     request.Date,
		PostID:   postID,
	}

	if err := h.postsService.CreateComment(r.Context(), comment); err != nil {
		slog.Error("Failed to create comment", "error", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Comment created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Comment created successfully", "postId", postID.Hex())
}

// GetPostsHandler возвращает список постов
func (h *Handler) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("GetPostsHandler started")

	posts, err := h.postsService.GetLastPosts(r.Context())
	if err != nil {
		slog.Error("Failed to get posts", "error", err)
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"posts":   posts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode posts response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Posts retrieved successfully", "count", len(posts))
}

func (h *Handler) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("GetCommentsHandler started")

	postID := r.URL.Query().Get("postId")
	if postID == "" {
		slog.Error("Post ID is required")
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		slog.Error("Invalid post ID", "postId", postID, "error", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.postsService.GetLastCommentsByPost(r.Context(), objectID)
	if err != nil {
		slog.Error("Failed to get comments", "postId", postID, "error", err)
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"comments": comments,
		"postId":   postID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode comments response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Comments retrieved successfully", "postId", postID, "count", len(comments))
}
func (h *Handler) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("UpdatePostHandler started")

	var request struct {
		PostID  string `json:"postId"`
		Author  string `json:"author"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid JSON: " + err.Error(),
		})
		return
	}

	postID, err := bson.ObjectIDFromHex(request.PostID)
	if err != nil {
		slog.Error("Invalid post ID", "postId", request.PostID, "error", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = h.postsService.UpdatePost(r.Context(), postID, request.Author, request.Title, request.Content)
	if err != nil {
		slog.Error("Failed to update post", "error", err)
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Post updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Post updated successfully", "postId", request.PostID)
}

// DeletePostHandler удаляет пост
func (h *Handler) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("DeletePostHandler started")

	var request struct {
		PostID string `json:"postId"`
		Author string `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	postID, err := bson.ObjectIDFromHex(request.PostID)
	if err != nil {
		slog.Error("Invalid post ID", "postId", request.PostID, "error", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = h.postsService.DeletePost(r.Context(), postID, request.Author)
	if err != nil {
		slog.Error("Failed to delete post", "error", err)
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Post deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Post deleted successfully", "postId", request.PostID)
}

// UpdateCommentHandler обновляет комментарий
func (h *Handler) UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("UpdateCommentHandler started")

	var request struct {
		CommentID string `json:"commentId"`
		Author    string `json:"author"`
		Content   string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	commentID, err := bson.ObjectIDFromHex(request.CommentID)
	if err != nil {
		slog.Error("Invalid comment ID", "commentId", request.CommentID, "error", err)
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	err = h.postsService.UpdateComment(r.Context(), commentID, request.Author, request.Content)
	if err != nil {
		slog.Error("Failed to update comment", "error", err)
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Comment updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Comment updated successfully", "commentId", request.CommentID)
}

// DeleteCommentHandler удаляет комментарий
func (h *Handler) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("DeleteCommentHandler started")

	var request struct {
		CommentID string `json:"commentId"`
		Author    string `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	commentID, err := bson.ObjectIDFromHex(request.CommentID)
	if err != nil {
		slog.Error("Invalid comment ID", "commentId", request.CommentID, "error", err)
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	err = h.postsService.DeleteComment(r.Context(), commentID, request.Author)
	if err != nil {
		slog.Error("Failed to delete comment", "error", err)
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Comment deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	slog.Info("Comment deleted successfully", "commentId", request.CommentID)
}
