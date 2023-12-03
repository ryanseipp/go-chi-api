package database

import (
	"context"
	"database/sql"
	"fmt"
	"go-chi-api/internal/domain"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type Service interface {
	Health() (bool, string)
	GetUserById(ctx context.Context, id int64) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)

	// Creates a user in the database, filling the Id field of the user on
	// success, or returning an error if the email is not unique
	CreateUser(ctx context.Context, user *domain.User) error
}

type service struct {
	db *sql.DB
}

var (
	database = os.Getenv("DB_DATABASE")
	password = os.Getenv("DB_PASSWORD")
	username = os.Getenv("DB_USERNAME")
	port     = os.Getenv("DB_PORT")
	host     = os.Getenv("DB_HOST")
)

const (
	HealthyMessage   = "Healthy"
	UnhealthyMessage = "Unhealthy"
)

func New() Service {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	s := &service{db: db}
	return s
}

func (s *service) Health() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
	if err != nil {
		return false, UnhealthyMessage
	}

	return true, HealthyMessage
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, username, status, password_hash, created_at, updated_at, deleted_at
FROM goapi.users
WHERE username = $1
LIMIT 1`,
		username,
	)

	var user domain.User
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime
	var status string
	err := row.Scan(
		&user.Id,
		&user.Username,
		&status,
		&user.PasswordHash,
		&user.CreatedAtTimestamp,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		return nil, err
	}

	switch status {
	case domain.ActiveStr:
		user.Status = domain.Active
	case domain.DeletedStr:
		user.Status = domain.Deleted
	}

	if updatedAt.Valid {
		user.UpdatedAtTimestamp = &updatedAt.Time
	} else {
		user.UpdatedAtTimestamp = nil
	}

	if deletedAt.Valid {
		user.DeletedAtTimestamp = &deletedAt.Time
	} else {
		user.DeletedAtTimestamp = nil
	}

	return &user, nil
}

func (s *service) GetUserById(ctx context.Context, id int64) (*domain.User, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, username, status, password_hash, created_at, updated_at, deleted_at
FROM goapi.users
WHERE id = $1
LIMIT 1`,
		id,
	)

	var user domain.User
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime
	var status string
	err := row.Scan(
		&user.Id,
		&user.Username,
		&status,
		&user.PasswordHash,
		&user.CreatedAtTimestamp,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		return nil, err
	}

	switch status {
	case domain.ActiveStr:
		user.Status = domain.Active
	case domain.DeletedStr:
		user.Status = domain.Deleted
	}

	if updatedAt.Valid {
		user.UpdatedAtTimestamp = &updatedAt.Time
	} else {
		user.UpdatedAtTimestamp = nil
	}

	if deletedAt.Valid {
		user.DeletedAtTimestamp = &deletedAt.Time
	} else {
		user.DeletedAtTimestamp = nil
	}

	return &user, nil
}

func (s *service) CreateUser(ctx context.Context, user *domain.User) error {
	row := s.db.QueryRowContext(ctx, `
INSERT INTO goapi.users (username, status, password_hash, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id`,
		user.Username,
		user.Status.ToString(),
		user.PasswordHash,
		user.CreatedAtTimestamp,
	)

	err := row.Scan(&user.Id)
	return err
}
