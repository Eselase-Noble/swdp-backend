package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

//
// Create User
//

// CreateUser godoc
// @Summary      Create user
// @Description  Creates a new user
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      User  true  "User"
// @Success      201   {object}  User
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/add [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := h.Service.AddUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

//
// Get User by ID
//

// GetUserByID godoc
// @Summary      Get user
// @Description  Retrieves a user by ID
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  User
// @Failure      404  {object}  map[string]string
// @Router       /users/get/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	id := (c.Param("id"))
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
	//	return
	//}

	user, err := h.Service.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

//
// List Users
//

// ListUsers godoc
// @Summary      List users
// @Description  Retrieves all users
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   User
// @Router       /users/all [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.Service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

//
// Update User
//

// UpdateUser godoc
// @Summary      Update user
// @Description  Updates user fields
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                 true  "User ID"
// @Param        body  body      map[string]interface{} true  "Fields to update"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Router       /users/update/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
	//	return
	//}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := h.Service.UpdateUser(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.Status(http.StatusNoContent)
}

//
// Delete User
//

// DeleteUser godoc
// @Summary      Delete user
// @Description  Soft deletes a user
// @Tags         User
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Router       /users/delete/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
	//	return
	//}

	if err := h.Service.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}
