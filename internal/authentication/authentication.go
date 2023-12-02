package authentication

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"go-chi-api/internal/domain"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/argon2"
)

const (
	Valid HashValidationResult = iota + 1
	ValidRehashNeeded
	Invalid

	argonMemory      uint32 = 12288
	argonIterations  uint32 = 3
	argonParallelism uint8  = 1
	argonSaltLength  uint8  = 16
	argonKeyLength   uint32 = 32

	jwtIssuer  string = "go-api"
	cookieName string = "token"

	ContextValueUserId string = "userId"
)

type (
	HashValidationResult int

	argon2Type     int
	argon2idParams struct {
		variant     argon2Type
		memory      uint32
		iterations  uint32
		parallelism uint8
		saltLength  uint8
		keyLength   uint32
	}

	appClaims struct {
		Name string `json:"name"`
		jwt.RegisteredClaims
	}

	Service interface {
		// Authentication middleware that will get the authenticated user for
		// the request. If authentication was successful, the user's id can be
		// retrieved from the context value ContextValueUserId. Otherwise, this
		// middleware returns an unauthenticated result to the client
		UseAuthentication(next http.Handler) http.Handler

		HashPassword(password string) (string, error)
		VerifyHashedPassword(password string, encodedPassword *string) (HashValidationResult, error)
		SetAuthenticationCookie(w http.ResponseWriter, user *domain.User) error
	}

	service struct {
		jwtSecret string
	}
)

var (
	ErrInvalidHash         = errors.New("Hash provided in an incorrect format")
	ErrIncompatibleVersion = errors.New("Hash utilizes unsupported Argon2 algorithm or version")
)

func New() Service {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Environment variable JWT_SECRET was not set")
	}

	return &service{
		jwtSecret,
	}
}

func (s *service) UseAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil || cookie.Expires.After(time.Now()) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(
			cookie.Value,
			&appClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(s.jwtSecret), nil
			},
			jwt.WithIssuer(jwtIssuer),
		)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*appClaims)
		if !ok || len(claims.Subject) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextValueUserId, id)
		w.Header().Set("Cache-Control", "max-age=0,private,must-revalidate")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *service) generateRandomBytes(n uint8) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *service) encodeHash(hash []byte, salt []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		b64Salt,
		b64Hash,
	)
}

func (s *service) decodeHash(encodedHash string) (hash []byte, salt []byte, p *argon2idParams, err error) {
	p = &argon2idParams{}
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		err = ErrInvalidHash
	}

	if vals[1] != "argon2id" && err == nil {
		err = ErrIncompatibleVersion
	}

	var version int
	_, scanErr := fmt.Sscanf(vals[2], "v=%d", &version)
	if (version != argon2.Version || scanErr != nil) && err == nil {
		err = ErrIncompatibleVersion
	}

	_, scanErr = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if scanErr != nil && err == nil {
		err = scanErr
	}

	salt, decodeErr := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if decodeErr != nil && err == nil {
		err = decodeErr
	}
	p.saltLength = uint8(len(salt))

	hash, decodeErr = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if decodeErr != nil && err == nil {
		err = decodeErr
	}
	p.keyLength = uint32(len(hash))

	return hash, salt, p, err
}

func (s *service) HashPassword(password string) (string, error) {
	salt, err := s.generateRandomBytes(argonSaltLength)
	if err != nil {
		return "", err
	}

	var hash []byte
	hash = argon2.IDKey(
		[]byte(password),
		salt,
		argonIterations,
		argonMemory,
		argonParallelism,
		argonKeyLength,
	)

	encodedHash := s.encodeHash(hash, salt)

	return encodedHash, nil
}

func (s *service) VerifyHashedPassword(password string, encoded *string) (HashValidationResult, error) {
	encodedPassword := "$argon2id$v=19$m=12288,t=3,p=1$RUhxczVSVE5SQV4z$T0blM/Jzk2V6LQ/TRNqfm5Mine3F6wP2564aq7Uxr+o"

	if encoded != nil {
		encodedPassword = *encoded
	}

	hash, salt, p, decodeErr := s.decodeHash(encodedPassword)
	passwordHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	compareResult := subtle.ConstantTimeCompare(hash, passwordHash)

	if compareResult == 0 || decodeErr != nil {
		return Invalid, nil
	}

	if p.memory != argonMemory ||
		p.iterations != argonIterations ||
		p.parallelism != argonParallelism ||
		p.keyLength != argonKeyLength ||
		p.saltLength != argonSaltLength {
		return ValidRehashNeeded, nil
	}

	return Valid, nil
}

func (s *service) SetAuthenticationCookie(w http.ResponseWriter, user *domain.User) error {
	now := time.Now()
	expiresAt := now.UTC().Add(72 * time.Hour)

	claims := appClaims{
		user.Username,
		jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   fmt.Sprintf("%d", user.Id),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tokenString,
		Expires:  expiresAt,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}
