package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-chi-api/internal/authentication"
	"go-chi-api/internal/database"
	"go-chi-api/internal/otel"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	server   *http.Server
	port     int
	db       database.Service
	auth     authentication.Service
	validate *validator.Validate
	otel     otel.Service
}

func NewServer(ctx context.Context, serviceName string, serviceVersion string) (*Server, error) {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port > math.MaxUint16 {
		log.Fatalf("Error: PORT is not a valid port (%s)", portStr)
	}

	otelService, err := otel.New(ctx, serviceName, serviceVersion)
	if err != nil {
		return nil, err
	}

	server := &Server{
		port:     port,
		db:       database.New(),
		auth:     authentication.New(),
		validate: validator.New(),
		otel:     otelService,
	}

	// Declare Server config
	server.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}

func (s *Server) ListenAndServe() error {
	log.Println(fmt.Sprintf("Starting server at %s", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() (err error) {
	log.Println("Shutdown requested, gracefully closing connections. This may take a minute...")
	stopCtx, cancelStopCtx := context.WithTimeout(context.Background(), time.Minute)
	defer cancelStopCtx()
	defer func() {
		err = errors.Join(err, s.otel.Shutdown(context.Background()))
	}()

	err = s.server.Shutdown(stopCtx)
	return
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
