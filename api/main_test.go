package api

import (
	"os"
	"testing"
	"time"

	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/XiaozhouCui/go-bank/db/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

// TestMain is the entry point for all tests in api.
func TestMain(m *testing.M) {
	// set gin to use TestMode instead of DebugMode
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
