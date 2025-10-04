package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kurtgray/blog-api-go/internal/middleware"
	"github.com/kurtgray/blog-api-go/internal/models"
	"github.com/kurtgray/blog-api-go/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostHandler struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostHandler(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// GET /api/posts
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postRepo.FindAllWithAuthor(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error fetching posts",
		})
		return
	}

	// TODO: Populate author data (we'll do this with aggregation later if needed)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts,
	})
}

// GET /api/posts/:postId
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// parse post id from param
	postID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "postId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid post ID",
		})
		return
	}

	post, err := h.postRepo.FindByIDWithAuthor(r.Context(), postID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Post not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"post": post,
	})
}

// POST /api/posts
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	// from client
	var req struct {
		Title     string `json:"title"`
		Text      string `json:"text"`
		ImgURL    string `json:"imgUrl"`
		Published bool   `json:"published"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// validate input
	if strings.TrimSpace(req.Title) == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Title must be specified.",
		})
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Text must be specified.",
		})
		return
	}

	post := &models.Post{
		Author:    user.ID,
		Title:     req.Title,
		Text:      req.Text,
		ImgURL:    req.ImgURL,
		Published: req.Published,
	}

	if err := h.postRepo.Create(r.Context(), post); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error creating post",
		})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"post":    post,
	})
}

// PUT /api/posts/:postId
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "postId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid post ID",
		})
		return
	}

	// from client
	var req struct {
		Title     string `json:"title"`
		Text      string `json:"text"`
		ImgURL    string `json:"imgUrl"`
		Published bool   `json:"published"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// can use any instead of interface{}
		respondJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	update := bson.M{
		"title":     req.Title,
		"text":      req.Text,
		"imgUrl":    req.ImgURL,
		"published": req.Published,
	}

	if err := h.postRepo.Update(r.Context(), postID, update); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Post updated",
	})
}

// PATCH /api/posts/:postId (update published status)
func (h *PostHandler) PatchPost(w http.ResponseWriter, r *http.Request) {
	postID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "postId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid post ID",
		})
		return
	}

	var req struct {
		Published bool `json:"published"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	update := bson.M{"published": req.Published}

	if err := h.postRepo.Update(r.Context(), postID, update); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// fetch updated post
	updatedPost, err := h.postRepo.FindByIDWithAuthor(r.Context(), postID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Post not found after update",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"updatedPost": updatedPost,
	})
}

// DELETE /api/posts/:postId
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "postId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid post ID",
		})
		return
	}

	if err := h.postRepo.Delete(r.Context(), postID); err != nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Post deleted.",
		"id":      postID.Hex(),
	})
}

// GET /api/users/:userId/posts
func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	userID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "userId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	posts, err := h.postRepo.FindByAuthor(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error fetching posts",
		})
		return
	}

	// separate published and unpublished
	var published, unpublished []models.Post
	for _, post := range posts {
		if post.Published {
			published = append(published, post)
		} else {
			unpublished = append(unpublished, post)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"posts": map[string][]models.Post{
			"published":   published,
			"unpublished": unpublished,
		},
	})
}
