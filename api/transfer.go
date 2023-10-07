package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/XiaozhouCui/go-bank/db/sqlc"
	"github.com/XiaozhouCui/go-bank/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	// binding is for validation
	// "currency" validator is registered in server.go, to replace binding "oneof=USD EUR CAD"
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest // incoming request
	// validate the incoming request JSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// validate currency for FromAccount
	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if fromAccount.Owner != authPayload.Username {
		err := fmt.Errorf("from account [%d] is not owned by the current user", fromAccount.ID)
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// validate currency for ToAccount
	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	// if req is valid, create transfer in db
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// create money transfer transaction
	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// return the result
	ctx.JSON(http.StatusOK, result)
}

// validate transfer currency against account currency
func (server *Server) validAccount(ctx *gin.Context, accountId int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountId)
	if err != nil {
		// 2 error scenarios: 404 and 500
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	// validate currency
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
