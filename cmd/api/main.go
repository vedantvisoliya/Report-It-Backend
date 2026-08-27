package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reportit-api/internal/app"
	"reportit-api/internal/httpserver"
	"time"
)

func main() {
	ctx := context.Background()
	app, err := app.StartApp(ctx)
	if err != nil {
		log.Fatalf("Startup failed: %v", err)
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
