package handlers

import (
	"net/http"

	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "register"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	user, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case services.ErrEmailInUse:
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
	})
}

// Login godoc
// @Summary Login and obtain JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body LoginRequest true "login"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	token, roleIDs, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidCredential:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not login"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"role_ids":     roleIDs,
	})
}
