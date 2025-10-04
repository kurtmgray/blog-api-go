package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kurtgray/blog-api-go/internal/handlers"
	"github.com/kurtgray/blog-api-go/internal/middleware"
)

type Router struct {
	userHandler    *handlers.UserHandler
	postHandler    *handlers.PostHandler
	commentHandler *handlers.CommentHandler
	authService    *middleware.AuthService
	corsMiddleware *cors.Cors
}

func New(
	userHandler *handlers.UserHandler,
	postHandler *handlers.PostHandler,
	commentHandler *handlers.CommentHandler,
	authService *middleware.AuthService,
	corsMiddleware *cors.Cors,
) *Router {
	return &Router{
		userHandler:    userHandler,
		postHandler:    postHandler,
		commentHandler: commentHandler,
		authService:    authService,
		corsMiddleware: corsMiddleware,
	}
}

func (rt *Router) Setup() *chi.Mux {
	r := chi.NewRouter()

	// global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.StripSlashes) 
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(rt.corsMiddleware.Handler)

	// root redirect
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/posts", http.StatusMovedPermanently)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// public

		// nested "/api/users", handler method
		r.Post("/users", rt.userHandler.CreateUser)
		r.Post("/users/login", rt.userHandler.Login)
		r.Get("/posts", rt.postHandler.GetAllPosts)
		r.Get("/posts/{postId}", rt.postHandler.GetPost)
		r.Get("/posts/{postId}/comments", rt.commentHandler.GetPostComments)
		r.Get("/posts/{postId}/comments/{commentId}", rt.commentHandler.GetComment)

		// protected by auth mw
		r.Group(func(r chi.Router) {
			r.Use(rt.authService.RequireAuth)

			// user
			r.Get("/users", rt.userHandler.GetCurrentUser)
			r.Get("/users/{userId}/posts", rt.postHandler.GetUserPosts)

			// post
			r.Post("/posts", rt.postHandler.CreatePost)
			r.Put("/posts/{postId}", rt.postHandler.UpdatePost)
			r.Patch("/posts/{postId}", rt.postHandler.PatchPost)
			r.Delete("/posts/{postId}", rt.postHandler.DeletePost)

			// comment
			r.Post("/posts/{postId}/comments", rt.commentHandler.CreateComment)
			r.Patch("/posts/{postId}/comments/{commentId}", rt.commentHandler.UpdateComment)
			r.Delete("/posts/{postId}/comments/{commentId}", rt.commentHandler.DeleteComment)
		})
	})

	return r
}
