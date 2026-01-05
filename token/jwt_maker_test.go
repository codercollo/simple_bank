package token

import (
	"testing"
	"time"

	"github.com/codercollo/simple_bank/util"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
)

// TestJWTMaker verifies successful JWT creation and validation
func TestJWTMaker(t *testing.T) {
	//Create JWT maker with valid secret key
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	//Prepare test inputs
	username := util.RandomOwner()
	duration := time.Minute

	//Expected issue and expiry times
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	//Create JWT token
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	//Verify token and extract payload
	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	//Validate payload contents
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssueAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

// TestExpiredJWTToken ensures expired tokens are rejected
func TestExpiredJWTToken(t *testing.T) {
	//Create JWT maker
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	//Create token with negative duration (already expired)
	token, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	//Verify token should fail with expiration error
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

// TestInvalidJWTTokenAlgNone ensures unsigned tokens are rejected
func TestInvalidJWTTokenALgNone(t *testing.T) {
	//Create valid payload
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)

	//Create JWT using "none" signing algorithm (insecure)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	//Create JWT maker
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	//Verification should fail due to invalid signing method
	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
