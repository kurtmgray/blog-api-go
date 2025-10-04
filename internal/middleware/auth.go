package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kurtgray/blog-api-go/internal/models"
	"github.com/kurtgray/blog-api-go/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserContextKey contextKey = "user"

// payload structure
type JWTClaims struct {
	UserID     string `json:"id"`
	Username   string `json:"username"`
	Admin      bool   `json:"admin"`
	CanPublish bool   `json:"canPublish"`
	// embedded struct adds direct access to expiresat, issuedat, issuer
	jwt.RegisteredClaims 
}

// service to handle operations
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// hashes plain string
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// checks if password matches hash
func (s *AuthService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// creates a JWT token for a user
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := JWTClaims{
		// decode the primitive.ObjectID -> string
		UserID:     user.ID.Hex(),
		Username:   user.Username,
		Admin:      user.Admin,
		CanPublish: user.CanPublish,
		RegisteredClaims: jwt.RegisteredClaims{
			// exp in 24h
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// create new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// sign and return
	return token.SignedString([]byte(s.jwtSecret))
}

// validates a JWT token, returns claims
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// middleware that validates JWT and adds user to context
func (s *AuthService) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		// <token>
		tokenString := parts[1]

		// validate
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// string ID to ObjectID
		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid user ID")
			return
		}

		// Fetch user from database
		user, err := s.userRepo.FindByID(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "user not found")
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// helper to get user from context
func GetUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// helper to send error responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
