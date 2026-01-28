package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const minSecretKeySize = 32

// JWTMaker creates and verifies JWT tokens using HMAC
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker initializes a JWT maker with a minimum secret key length
func NewJWTMaker(secretKey string) (Maker, error) {
	//Enforce minimum secret key length for security
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

// CreateToken generates a signed JWT for a given username and duraion
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	//Create token payload with expiration
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	//Create JWT with HMAC-SHA265 signing method
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	//Sign token using secret key
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err

}

// VerifyToken validates the JWT and returns its payload
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {

	//Provide secret key and validate signing method
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		//Ensure token uses HMAC signing
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	//Parse and validate token claims
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		//Distinguish expired token from other errors
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	//Extract and assert payload type
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil

}
