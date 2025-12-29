package api

import (
	"net/http"
	"time"

	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Request body for creating a user
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	Fullname string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// Response body after user creation
type createUserResponse struct {
	Username          string    `json:"username"`
	HashedPassword    string    `json:"hashed_password"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// createUser registers a new user and returns the created record
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest

	//Bind and validate request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//Hash the plain-text password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//Buils DB parameters
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.Fullname,
		Email:          req.Email,
	}

	//Insert user into database
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		//Handle duplicate username/email
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Build response without sensitive data
	rsp := createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}

	//Respond with success
	ctx.JSON(http.StatusOK, rsp)
}
