package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain is the entry point for all tests in api.
func TestMain(m *testing.M) {
	// set gin to use TestMode instead of DebugMode
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
