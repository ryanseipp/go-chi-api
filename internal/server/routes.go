package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidContentType = errors.New("Invalid Content-Type")
	ErrInvalidJsonBody    = errors.New("Invalid JSON body")
	ErrValidationFailed   = errors.New("JSON request failed validation")
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Minute))

	s.swaggerRouter(r)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/", s.HelloWorldHandler)
		s.healthRouter(r)
		s.authRouter(r)
	})

	return r
}

func (s *Server) badRequestResponse(w http.ResponseWriter, errors any) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)

	log.Println("Errors", errors)

	response := make(map[string]any)
	response["type"] = "https://tools.ietf.org/html/rfc9110#section-15.5.1"
	response["title"] = "Bad Request"
	response["status"] = http.StatusBadRequest
	response["errors"] = errors

	jsonResponse, _ := json.Marshal(&response)
	w.Write(jsonResponse)
}

func (s *Server) jsonResponse(w http.ResponseWriter, response any) {
	body, err := json.Marshal(response)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// Decodes JSON from the request body, writing the result into v. If an error
// occurs, this writes the proper response into w
func (s *Server) decodeJson(w http.ResponseWriter, r *http.Request, v any) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		s.badRequestResponse(w, "Expected Content-Type of application/json")
		return ErrInvalidContentType
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&v); err != nil || v == nil {
		log.Println("Failed to decode", err)
		s.badRequestResponse(w, err)
		return ErrInvalidJsonBody
	}

	if err := s.validate.Struct(v); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		s.badRequestResponse(w, validationErrors)
		return ErrValidationFailed
	}

	return nil
}

// HelloWorld
// @Summary Say hello!
// @Description Hello, there!
// @Tags hello
// @Accept json
// @Produce json
// @Router /v1/ [get]
func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
