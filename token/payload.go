package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrExpiredToken indicates the token has passed its expiration time
// ErrInvalidToken indicates the token is malformed or invalid
var (
	ErrExpiredToken = errors.New("token has exprired")
	ErrInvalidToken = errors.New("token has expired")
)

type TokenType byte

const (
	TokenTypeAccessToken  = 1
	TokenTypeRefreshToken = 2
)

// Payload defines the JWT payload structure
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssueAt   time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a unique ID and expiry
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	//Generate unique token ID
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	//Initialize payload timestamps
	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssueAt:   time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Valid validates the payload by checking token expiration
func (payload *Payload) Valid() error {
	//Reject token if expired
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
