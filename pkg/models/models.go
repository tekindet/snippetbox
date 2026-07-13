package models

import (
	"errors"
	"time"
)

var (
	ErrNoRecord           = errors.New("error: no record found")
	ErrInvalidCredentials = errors.New("models : invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

type Snippet struct {
	ID      int       `json:"id" gorm:"Column:id"`
	Title   string    `json:"title" gorm:"Column:title"`
	Content string    `json:"content" gorm:"Column:content"`
	Created time.Time `json:"created_at" gorm:"Column:created"`
	Expires time.Time `json:"expires_at" gorm:"Column:expires"`
}

type User struct {
	ID             int    `json:"id" gorm:"Column:id"`
	Name           string `json:"name" gorm:"Column:name"`
	Email          string `json:"email" gorm:"Column:email"`
	HashedPassword []byte `gorm:"Column:hashed_password"`
	Created        time.Time
}
