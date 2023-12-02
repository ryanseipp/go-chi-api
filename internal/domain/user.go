package domain

import "time"

type UserStatus int

const (
	Active UserStatus = iota + 1
	Deleted
)

const (
	ActiveStr  = "Active"
	DeletedStr = "Deleted"
)

func (s UserStatus) ToString() string {
	switch s {
	case Active:
		return ActiveStr
	default:
		return DeletedStr
	}
}

type User struct {
	Id                 int64
	Username           string
	Status             UserStatus
	PasswordHash       string
	CreatedAtTimestamp time.Time
	UpdatedAtTimestamp *time.Time
	DeletedAtTimestamp *time.Time
}

func NewUser(username string, passwordHash string) *User {
	return &User{
		Id:                 0,
		Username:           username,
		Status:             Active,
		PasswordHash:       passwordHash,
		CreatedAtTimestamp: time.Now().UTC(),
		UpdatedAtTimestamp: nil,
		DeletedAtTimestamp: nil,
	}
}
