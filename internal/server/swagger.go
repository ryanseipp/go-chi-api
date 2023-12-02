package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger/v2"
)

func (s *Server) swaggerRouter(r chi.Router) {
	r.Get(
		"/swagger/*",
		httpSwagger.Handler(
			httpSwagger.URL(
				fmt.Sprintf("http://localhost:%d/swagger/doc.json", s.port),
			),
		),
	)
}
