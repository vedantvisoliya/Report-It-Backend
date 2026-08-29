package httpserver

import (
	"reportit-api/internal/app"
	emailotp "reportit-api/internal/email-otp"
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

	otpRepo := emailotp.NewOTPRepo(app.DB)
	otpService := emailotp.NewOTPService(otpRepo, app.Config.GmailAppPassword)
	otpHandler := emailotp.NewOTPHandler(otpService)

	userRefreshRepo := user.NewRefreshRepo(app.DB)
	userRepo := user.NewUserRepo(app.DB)
	userSvc := user.NewUserService(userRepo, app.Config, userRefreshRepo)
	userHandler := user.NewUserHandler(userSvc)

	// unauth routes -> public access
	r.POST("/auth/register", userHandler.Register)
	r.POST("/auth/login", userHandler.Login)
	r.POST("/auth/refresh", userHandler.Refresh)
	r.POST("/auth/logout", userHandler.Logout)
	r.POST("/auth/send-otp", otpHandler.SendOTP)
	r.POST("/auth/verify-otp", otpHandler.VerifyOTP)

	// protected user routes
	userApis := r.Group("/user")
	userApis.Use(middleware.AuthRequired(app.Config.JwtSecret))
	{
		userApis.GET("", userHandler.GetAllUsers)
		userApis.GET("/:id", userHandler.GetSingleUserDetails)
		userApis.PUT("/:id", userHandler.Update)
		userApis.DELETE("/:id", userHandler.DeleteUser)
	}

	postRepo := post.NewPostRepo(app.DB)
	postSvc := post.NewPostService(postRepo, app.Config)
	postHandler := post.NewPostHandler(postSvc)

	// protected post routes
	postApis := r.Group("/post")
	postApis.Use(middleware.AuthRequired(app.Config.JwtSecret))
	{
		postApis.POST("/:id", postHandler.CreatePost)
		postApis.GET("user/:id", postHandler.GetAllUserPosts)
		postApis.DELETE("/:id", postHandler.RemovePost)
		postApis.PATCH("/:id", postHandler.UpdatePost)
		postApis.GET("/:id", postHandler.GetSinglePost)
	}

	// admin routes
	adminApis := r.Group("/authorized/admin")
	adminApis.Use(middleware.AuthRequired(app.Config.JwtSecret), middleware.RequiredAdmin())
	{
		adminApis.DELETE("/remove-user/:id", userHandler.DeleteUser)
		adminApis.DELETE("remove-post/:id", postHandler.RemovePost)
		adminApis.POST("/auth/login", userHandler.Login)
		adminApis.POST("/auth/logout", userHandler.Logout)
	}

	return r
}
