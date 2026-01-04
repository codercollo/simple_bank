package token

import "time"

//Maker defines the interface for token creation and verification
type Maker interface {
	//CreateToken generates a signed token for a user with a given duration
	CreateToken(username string, duration time.Duration) (string, error)

	//VerifyToken validates a token and returns its payload
	VerifyToken(token string) (*Payload, error)
}
