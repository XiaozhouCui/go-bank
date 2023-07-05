package api

import (
	"net/http"

	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createAccountRequest struct {
	// binding is for validation
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest // incoming request
	// validate the incoming request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// if req is valid, create account in db
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// return the account
	ctx.JSON(http.StatusOK, account)
}
