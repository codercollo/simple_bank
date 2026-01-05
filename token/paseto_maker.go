package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

// PasetoMaker creates and verifies PASETO tokens using symmetric encryption
type PasetoMaker struct {
	paseto      *paseto.V2
	symetrickey []byte
}

// NewPasetoMaker initializes a PasetoMaker with a valid symmmetric key
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	//Ensure key size matches ChaCha20-Poly1305 requirements
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:      paseto.NewV2(),
		symetrickey: []byte(symmetricKey),
	}

	return maker, nil
}

// CreateToken generates an encrypted PASETO token for a user
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	//Build token payload
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	//Encrypt payload into token
	return maker.paseto.Encrypt(maker.symetrickey, payload, nil)
}

// VerifyToken decrypts and validates a PASETO token
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {

	payload := &Payload{}

	//Decrypt token into payload
	err := maker.paseto.Decrypt(token, maker.symetrickey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	//Validate payload claims
	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
