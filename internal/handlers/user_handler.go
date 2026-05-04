package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"go-api/internal/repositories"
	"go-api/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

type CreateUserRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	RoleIDs  []string `json:"role_ids"`
}

type UpdateUserRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password"`
	RoleIDs  []string `json:"role_ids"`
}

type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	users, total, err := h.userService.ListUsers(search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"search":    search,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	user, err := h.userService.CreateUser(req.Email, req.Password, req.RoleIDs)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user input"})
		case errors.Is(err, services.ErrEmailInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	user, err := h.userService.UpdateUser(id, req.Email, req.Password, req.RoleIDs)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user input"})
		case errors.Is(err, repositories.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, services.ErrEmailInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.userService.DeleteUser(id); err != nil {
		switch {
		case errors.Is(err, services.ErrUserInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		case errors.Is(err, repositories.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	roles, err := h.userService.GetUserRoles(userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		case errors.Is(err, repositories.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch user roles"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *UserHandler) AssignUserRoles(c *gin.Context) {
	userID := c.Param("id")
	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	roles, err := h.userService.AssignRoles(userID, req.RoleIDs)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user input"})
		case errors.Is(err, repositories.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not assign roles"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *UserHandler) RemoveUserRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleId")
	roles, err := h.userService.RemoveRole(userID, roleID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user input"})
		case errors.Is(err, repositories.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, repositories.ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not remove role"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}
