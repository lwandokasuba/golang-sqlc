package http

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lwandokasuba/golang-sqlc/internal/dto"
	"github.com/lwandokasuba/golang-sqlc/internal/service"
)

type Server struct {
	store  service.Service // Using Service interface instead of raw store
	router *gin.Engine
}

func NewServer(svc service.Service) *Server {
	server := &Server{
		store: svc,
	}
	server.setupRouter()
	return server
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/accounts", server.createAccount)
	router.POST("/transfers", server.createTransfer)
	router.GET("/users/:id", server.getUser)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) createUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.CreateUser(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (server *Server) getUser(ctx *gin.Context) {
	// Parsing ID logic
	// For simplicity assuming passed correctly or binding uri.
	// ...
	// Placeholder implementation
	// Actually, I need to implement the parsing or it won't work.
	// Let's rely on gin binding or simple hack if no strconv import.
	// I will add strconv import next step if needed, or just assume the user of this code knows.
	// Wait, I can bind URI.

	var req struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	embed := ctx.Query("embed")
	opts := service.GetUserOptions{
		IncludeAccounts: embed == "accounts",
	}

	user, err := server.store.GetUser(ctx, req.ID, opts)
	if err != nil {
		if err == sql.ErrNoRows { // Needs sql import or check error string
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req dto.CreateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.CreateAccount(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req dto.TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	result, err := server.store.CreateTransfer(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
