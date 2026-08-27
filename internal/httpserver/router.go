package httpserver

import (
	"reportit-api/internal/app"
	"reportit-api/internal/middleware"
	"reportit-api/internal/post"
	"reportit-api/internal/user"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *app.App) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// health endpoint
	r.GET("/health", health)

	userRefreshRepo := user.NewRefreshRepo(app.DB)
	userRepo := user.NewUserRepo(app.DB)
	userSvc := user.NewUserService(userRepo, app.Config, userRefreshRepo)
	userHandler := user.NewUserHandler(userSvc)

	// unauth routes -> public access
	r.POST("/auth/register", userHandler.Register)
	r.POST("/auth/login", userHandler.Login)
	r.POST("/auth/refresh", userHandler.Refresh)
	r.POST("/auth/logout", userHandler.Logout)

	// protected routes
	userApis := r.Group("/user")
	userApis.Use(middleware.AuthRequired(app.Config.JwtSecret))
	{
		userApis.GET("", userHandler.GetAllUsers)
		userApis.GET("/:id", userHandler.GetSingleUserDetails)
		userApis.PUT("/:id", userHandler.Update)

		userApis.DELETE(
			"/:id",
			middleware.RequiredAdmin(),
			userHandler.DeleteUser,
		)
	}

	postRepo := post.NewPostRepo(app.DB)
	postSvc := post.NewPostService(postRepo, app.Config)
	postHandler := post.NewPostHandler(postSvc)

	// protected routes
	postApis := r.Group("/post")
	postApis.Use(middleware.AuthRequired(app.Config.JwtSecret))
	{
		postApis.POST("/:id", postHandler.CreatePost)
		postApis.GET("user/:id", postHandler.GetAllUserPosts)
		postApis.DELETE("/:id", postHandler.RemovePost)
		postApis.PATCH("/:id", postHandler.UpdatePost)
		postApis.GET("/:id", postHandler.GetSinglePost)
	}

	return r
}
