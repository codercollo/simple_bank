package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Request body for account creation
type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

// createAccount handles HTTP requests to creare a new bank account
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest

	//Validate input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//Get authenticated user
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	//Prepare DB params
	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	//Execute DB insert account
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		//Handle constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//Success response
	ctx.JSON(http.StatusOK, account)

}

// URI params for get account
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getAccount gets account by ID
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	//Bind URI params
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//Get account
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//Check ownership
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	//Success response
	ctx.JSON(http.StatusOK, account)

}

// Query params for listing accounts
type ListAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// List accounts with pagination
func (server *Server) listAccount(ctx *gin.Context) {
	var req ListAccountRequest

	//Bind query params
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//Get authenticated user
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	//Prepare DB params
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	//Fetch accounts
	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//Return accounts
	ctx.JSON(http.StatusOK, accounts)
}

// // Update account request
// type updateAccountRequest struct {
// 	Balance int64 `json:"balance" binding:"required"`
// }

// // Update account balance
// func (server *Server) updateAccount(ctx *gin.Context) {
// 	//Parse and validate account ID
// 	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
// 	if err != nil || id < 1 {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
// 		return
// 	}

// 	//Bind JSON body
// 	var req updateAccountRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	//Update account
// 	account, err := server.store.UpdateAccount(ctx, db.UpdateAccountParams{
// 		ID:      id,
// 		Balance: req.Balance,
// 	})
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			ctx.JSON(http.StatusNotFound, errorResponse(err))
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	//Return updated account
// 	ctx.JSON(http.StatusOK, account)
// }

// // deleteAccount deletes an account
// func (server *Server) deleteAccount(ctx *gin.Context) {
// 	//Parse account ID  from URL
// 	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
// 	if err != nil || id < 1 {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
// 		return
// 	}

// 	//Delete account in DB
// 	err = server.store.DeleteAccount(ctx, id)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			ctx.JSON(http.StatusNotFound, errorResponse(err))
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	//Success response
// 	ctx.JSON(http.StatusOK, gin.H{"mesage": "account deleted"})
// }
