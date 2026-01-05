package token

import (
	"testing"
	"time"

	"github.com/codercollo/simple_bank/util"
	"github.com/stretchr/testify/require"
)

// TestPasetoMaker verifies successful PASETO Token creation and validation
func TestPasetoMaker(t *testing.T) {
	//Create token maker
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	//Token inputs
	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	//Create token
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	//Verify token
	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	//Validate Payload
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssueAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

// TestExpiredPasetoToken verifies that expired tokens are rejected
func TestExpiredPasetoToken(t *testing.T) {
	//Create token maker
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	//Create expired token
	token, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	//Verify token fails
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)

}

// func TestPasetoWrongTokenType(t *testing.T){
// 	maker, err := NewPasetoMaker(util.RandomString(32))
// 	require.NoError(t, err)

// 	token, payload, err := maker.CreateToken(util.RandomOwner(), util.DepositorRole, time.Minute, TokenTypeAccessToken)

// }
