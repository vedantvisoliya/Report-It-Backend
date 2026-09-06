package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reportit-api/internal/app"
	"reportit-api/internal/db"
	"reportit-api/internal/httpserver"
	"time"

	_ "reportit-api/docs"
)

const (
	ReportitOTPCol = "reportit_otp"
)

// @title                       Report It API
// @version                     1.0
// @description                 Backend API for Report It, a campus incident reporting platform. Students register with their college email, verify it over an email OTP, and file reports such as lost and found, harassment, ragging or canteen overcharging. Admins can moderate users and posts.
//
// @contact.name                Report It API Support
//
// @host                        localhost:5000
// @BasePath                    /
// @schemes                     http https
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and the JWT access token.
func main() {
	ctx := context.Background()
	app, err := app.StartApp(ctx)
	if err != nil {
		log.Fatalf("Startup failed: %v", err)
	}

	err = db.EnsureEmailIndexes(app.DB.Collection(ReportitOTPCol))
	if err != nil {
		log.Fatalf("failed to create email indexex: %v", err)
	}

	err = db.EnsureOTPIndexes(app.DB.Collection(ReportitOTPCol))
	if err != nil {
		log.Fatalf("failed to create otp indexex: %v", err)
	}

	defer func() {
		if err := app.CloseApp(ctx); err != nil {
			log.Printf("warning: shutting app: %v", err)
		}
	}()

	router := httpserver.NewRouter(app)
	addr := fmt.Sprintf(":%s", app.Config.ServerPort)

	srv := &http.Server{
		Addr:        addr,
		Handler:     router,
		ReadTimeout: 10 * time.Second,
	}

	log.Printf("Server running on port %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Printf("error: server is closed")
			return
		}
		log.Fatalf("error: server error (%v)", err)
	}
}
