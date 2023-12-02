package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-chi-api/internal/authentication"
	"go-chi-api/internal/database"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port     int
	db       database.Service
	auth     authentication.Service
	validate *validator.Validate
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:     port,
		db:       database.New(),
		auth:     authentication.New(),
		validate: validator.New(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

type JsonTime struct {
	*time.Time
}

func (t *JsonTime) MarshalJSON() ([]byte, error) {
	if t.Time == nil {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", t.Time.UTC().Round(time.Second).Format(time.RFC3339))), nil
}

type JsonDuration struct {
	time.Duration
}

func (t *JsonDuration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"\"%02d:%02d:%02d.%06d\"",
		int(t.Duration.Hours()),
		int(t.Duration.Minutes()),
		int(t.Duration.Seconds()),
		t.Duration.Microseconds()),
	), nil
}
