package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Transfer request payload
type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// createTransfer handles money transfer between accounts
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest

	//Validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//Validate source and destination accounts
	if !server.validAccount(ctx, req.FromAccountID, req.Currency) {
		return
	}

	if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	//Execute transfer transaction
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//Success response
	ctx.JSON(http.StatusOK, result)
}

// validAccount verifies account existence and currency consistency
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {

	//Fetch account by ID
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		//Account not found
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		//Database error
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	//Validate currency match
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}
