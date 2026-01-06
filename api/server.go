package api

import (
	"fmt"

	db "github.com/codercollo/simple_bank/db/sqlc"
	"github.com/codercollo/simple_bank/token"
	"github.com/codercollo/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	config     util.Config
}

// NewServer creates a new HTTP server and setup routing
func NewServer(store db.Store, config util.Config) (*Server, error) {

	//Create PASETO token maker using the symmetric key
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	//Initialize server with dependencies
	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	//Register custom currency validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	//Setup HTTP routes
	server.setupRouter()
	return server, nil

}

// setupRouter configures all API routes and middleware
func (server *Server) setupRouter() {
	///Create Gin router
	router := gin.Default()

	//Public user routes
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	//Auth-protected routes
	authRoutes := router.Group("/").Use(authMiddleware((server.tokenMaker)))

	//Account routes
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)
	// authRoutes.PATCH("/accounts/:id", server.updateAccount)
	// authRoutes.DELETE("/accounts/:id", server.deleteAccount)

	//Transfer routes
	authRoutes.POST("/transfers", server.createTransfer)

	//Assign router to server
	server.router = router

}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// errorResponse formats errors into a consistent JSON response
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
