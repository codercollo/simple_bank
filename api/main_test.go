package api

import (
	"os"
	"testing"
	"time"

	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// newTestServer creates a test server with mock store and config
func newTestServer(t *testing.T, store db.Store) *Server {
	//Test configuration
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	//Initialize server
	server, err := NewServer(store, config)
	require.NoError(t, err)
	return server

}

// TestMain sets up global test configuration
func TestMain(m *testing.M) {

	//Use Gin test mode
	gin.SetMode(gin.TestMode)

	//Run tests
	os.Exit(m.Run())
}
