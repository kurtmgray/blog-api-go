package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kurtgray/blog-api-go/internal/middleware"
	"github.com/kurtgray/blog-api-go/internal/models"
	"github.com/kurtgray/blog-api-go/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentHandler struct {
	commentRepo repository.CommentRepository
}

func NewCommentHandler(commentRepo repository.CommentRepository) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
	}
}

// GET /api/posts/:postId/comments
func (h *CommentHandler) GetPostComments(w http.ResponseWriter, r *http.Request) {
	postID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "postId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid post ID",
		})
		return
	}

	comments, err := h.commentRepo.FindByPostWithAuthor(r.Context(), postID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error fetching comments",
		})
		return
	}

	if comments == nil {
		comments = []models.CommentWithAuthor{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"comments": comments,
	})
}

// GET /api/posts/:postId/comments/:commentId
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "commentId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid comment ID",
		})
		return
	}

	comment, err := h.commentRepo.FindByIDWithAuthor(r.Context(), commentID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Comment not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"comment": comment,
	})
}

// POST /api/posts/:postId/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	postID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "postId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid post ID",
		})
		return
	}

	// comment from client
	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Comment must be entered.",
		})
		return
	}

	comment := &models.Comment{
		Author: user.ID,
		Text:   req.Text,
		Post:   postID,
	}

	if err := h.commentRepo.Create(r.Context(), comment); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error creating comment",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"comment": comment,
	})
}

// PATCH /api/posts/:postId/comments/:commentId
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "commentId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid comment ID",
		})
		return
	}

	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if err := h.commentRepo.Update(r.Context(), commentID, req.Text); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// fetch updated comment
	updatedComment, err := h.commentRepo.FindByIDWithAuthor(r.Context(), commentID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Comment not found after update",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"updatedComment": updatedComment,
	})
}

// DELETE /api/posts/:postId/comments/:commentId
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "commentId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid comment ID",
		})
		return
	}

	if err := h.commentRepo.Delete(r.Context(), commentID); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Comment deleted.",
		"id":      commentID.Hex(),
	})
}
