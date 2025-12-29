package db

import (
	"context"
	"testing"
	"time"

	"github.com/codercollo/simple_bank/util"
	"github.com/stretchr/testify/require"
)

// createRandomUser inserts a random user into the database and validates it
func createRandomUser(t *testing.T) User {
	//Hash a plain password
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	//Input params
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	//Create user
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	//Field validation
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	//Timestamps
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

// TestCreateUser ensures user creation works
func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

// TestGetUser ensures a user can be fetched by username
func TestGetUser(t *testing.T) {
	//Create user
	user1 := createRandomUser(t)

	//Fetch user
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	//Field validation
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)

	//Timestamp comparison
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}
