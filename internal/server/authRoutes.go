package server

import (
	"go-chi-api/internal/authentication"
	"go-chi-api/internal/domain"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) authRouter(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.registerUser)
		r.Post("/login", s.loginUser)

		r.Route("/current", func(r chi.Router) {
			r.Use(s.auth.UseAuthentication)
			r.Get("/", s.getCurrentUser)
		})
	})
}

type LoginUserRequest struct {
	Username string `json:"username" example:"myusername123" validate:"required,min=1,max=256"`
	Password string `json:"password" example:"superpassword" format:"password" validate:"required,min=16,max=64"`
}

// LoginUser
// @Summary Login user
// @Description Log in user via username and password
// @Tags auth
// @Accept json
// @Param request body server.LoginUserRequest true "Login Request Body"
// @Success 204
// @Failure 401
// @Failure 500
// @Router /v1/auth/login [post]
func (s *Server) loginUser(w http.ResponseWriter, r *http.Request) {
	var request LoginUserRequest
	if err := s.decodeJson(w, r, &request); err != nil {
		log.Println(err)
		return
	}

	user, dbErr := s.db.GetUserByUsername(r.Context(), request.Username)
	var passwordHash *string
	if user != nil {
		passwordHash = &user.PasswordHash
	}

	hashResult, hashErr := s.auth.VerifyHashedPassword(request.Password, passwordHash)
	if dbErr != nil || user == nil || hashErr != nil || hashResult == authentication.Invalid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if hashResult == authentication.ValidRehashNeeded {
		passwordHash, err := s.auth.HashPassword(request.Password)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user.PasswordHash = passwordHash
	}

	if err := s.auth.SetAuthenticationCookie(w, user); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type RegisterUserRequest struct {
	Username string `json:"username" example:"myusername123" validate:"required,min=1,max=256"`
	Password string `json:"password" example:"superpassword" format:"password" validate:"required,min=16,max=64"`
}

// RegisterUser
// @Summary Register user
// @Description Register user with the given username and password
// @Tags auth
// @Accept json
// @Param request body server.RegisterUserRequest true "Register Request Body"
// @Success 204
// @Failure 401
// @Router /v1/auth/register [post]
func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {
	var request RegisterUserRequest
	if err := s.decodeJson(w, r, &request); err != nil {
		log.Println(err)
		return
	}

	passwordHash, err := s.auth.HashPassword(request.Password)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := domain.NewUser(request.Username, passwordHash)
	if err := s.db.CreateUser(r.Context(), user); err != nil {
		log.Println(err)
		s.badRequestResponse(w, "Cannot create user with the information provided")
	}

	w.Header().Set("Location", "/auth/current")
	w.WriteHeader(http.StatusCreated)
}

type GetCurrentUserResponse struct {
	Id        int64    `json:"id"`
	Username  string   `json:"username"`
	CreatedAt JsonTime `json:"created_at" swaggertype:"string" format:"date-time"`
	UpdatedAt JsonTime `json:"updated_at,omitempty" swaggertype:"string" format:"date-time"`
}

// GetCurrentUser
// @Summary Get current user details
// @Description Gets the details of the authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} server.GetCurrentUserResponse
// @Failure 401
// @Router /v1/auth/current [get]
func (s *Server) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(authentication.ContextValueUserId).(int64)
	user, err := s.db.GetUserById(r.Context(), userId)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	response := GetCurrentUserResponse{
		Id:        user.Id,
		Username:  user.Username,
		CreatedAt: JsonTime{&user.CreatedAtTimestamp},
		UpdatedAt: JsonTime{user.UpdatedAtTimestamp},
	}

	s.jsonResponse(w, &response)
}
