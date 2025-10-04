package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kurtgray/blog-api-go/internal/middleware"
	"github.com/kurtgray/blog-api-go/internal/models"
	"github.com/kurtgray/blog-api-go/internal/repository"
)

type UserHandler struct {
	userRepo    repository.UserRepository
	authService *middleware.AuthService
}

func NewUserHandler(userRepo repository.UserRepository, authService *middleware.AuthService) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		authService: authService,
	}
}

// POST /api/users (user registration)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Fname    string `json:"fname"`
		Lname    string `json:"lname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if err := h.validateUserInput(req.Username, req.Password, req.Fname, req.Lname); err !=
		nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	existingUser, err := h.userRepo.FindByUsername(r.Context(), req.Username)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Database error",
		})
		return
	}

	if existingUser != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Username is already taken.",
			"field":   "usernameTaken",
		})
		return
	}

	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error creating user",
		})
		return
	}

	user := &models.User{
		Username:   req.Username,
		Password:   hashedPassword,
		Fname:      req.Fname,
		Lname:      req.Lname,
		Admin:      false,
		CanPublish: false,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error saving user",
		})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"user": map[string]interface{}{
			"id":         user.ID.Hex(),
			"fname":      user.Fname,
			"username":   user.Username,
			"canPublish": user.CanPublish,
			"admin":      user.Admin,
		},
	})
}

// POST /api/users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		GoogleID string `json:"googleId,omitempty"`
		Profile  struct {
			Sub        string `json:"sub"`
			Name       string `json:"name"`
			GivenName  string `json:"given_name"`
			FamilyName string `json:"family_name"`
		} `json:"profile,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	var user *models.User
	var err error

	// Google OAuth
	if req.GoogleID != "" {
		user, err = h.userRepo.FindByGoogleID(r.Context(), req.GoogleID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Database error",
			})
			return
		}

		// create new user if doesn't exist
		if user == nil {
			user = &models.User{
				GoogleID:   req.Profile.Sub,
				Username:   req.Profile.Name,
				Fname:      req.Profile.GivenName,
				Lname:      req.Profile.FamilyName,
				Admin:      false,
				CanPublish: false,
			}
			if err := h.userRepo.Create(r.Context(), user); err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Error creating user",
				})
				return
			}
		}
	} else {
		// non-oauth login
		user, err = h.userRepo.FindByUsername(r.Context(), req.Username)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Database error",
			})
			return
		}

		if user == nil {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "User does not exist",
				"field":   "username",
			})
			return
		}

		if err := h.authService.ComparePassword(user.Password, req.Password); err != nil {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Password does not match",
				"field":   "password",
			})
			return
		}
	}

	// generate jwt
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Error generating token",
		})
		return
	}

	// return token w. user info
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged in successfully",
		"token":   token,
		"user": map[string]interface{}{
			"id":         user.ID.Hex(),
			"fname":      user.Fname,
			"username":   user.Username,
			"admin":      user.Admin,
			"canPublish": user.CanPublish,
			"posts":      user.Posts,
		},
	})
}

// GET /api/users (refresh/verify token)
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// User is already in context from auth middleware
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"username":   user.Username,
			"id":         user.ID.Hex(),
			"fname":      user.Fname,
			"canPublish": user.CanPublish,
			"admin":      user.Admin,
			"posts":      user.Posts,
		},
	})
}

func (h *UserHandler) validateUserInput(username, password, fname, lname string) error {
	if len(strings.TrimSpace(fname)) == 0 {
		return jsonError("First name must be specified.")
	}
	if len(strings.TrimSpace(lname)) == 0 {
		return jsonError("Last name must be specified.")
	}
	if len(strings.TrimSpace(username)) == 0 {
		return jsonError("Username must be specified.")
	}
	if len(password) < 6 {
		return jsonError("Password must be at least 6 characters.")
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func jsonError(message string) error {
	return &validationError{message: message}
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
