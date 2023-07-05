package api

import (
	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	store  *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add routes to router
	router.POST("/accounts", server.createAccount)

	server.router = router
	return server
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
