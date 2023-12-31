package main

import (
	"context"
	"errors"
	_ "go-chi-api/docs"
	"go-chi-api/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	serviceName    = "go-chi-api"
	serviceVersion = "1.0"
)

// @title        Go Chi API
// @version      1.0
// @description  This is a sample server.
// @license.name MIT
// @license.url  http://github.com/ryanseipp/go-chi-api/LICENSE
// @host         localhost:3000
// @BasePath     /
func main() {
	server, err := server.NewServer(context.Background(), serviceName, serviceVersion)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln("Error:", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	if err := server.Shutdown(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Server shutdown")
}
