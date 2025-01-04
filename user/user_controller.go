package user

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateAndBind(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	if err := validate.Struct(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

// PatchUserHandler godoc
// @Router /users [post]
// @Accept json
// @Produce json
// @Param user body EditUserDto true " "
// @Success 201 {object} UserDto "user"
func createUserHandler(c *gin.Context) {
	var user User
	if !ValidateAndBind(c, &user) {
		return
	}

	if err := createUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create user: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// PatchUserHandler godoc
// @Router /users/{id} [get]
// @Accept json
// @Produce json
// @Param id path string true " "
// @Success 200 {object} UserDto "user"
func getUserHandler(c *gin.Context) {
	id := c.Param("id")
	user, err := getUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get user: %v", err),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUserHandler godoc
// @Router /users/{id} [patch]
// @Accept json
// @Produce json
// @Param id path string true " "
// @Param user body EditUserDto true " "
// @Success 200 {object} UserDto "user"
func updateUserHandler(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := updateUser(id, updates)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUserHandler godoc
// @Router /users/{id} [put]
// @Accept json
// @Produce json
// @Param id path string true " "
// @Param user body EditUserDto true " "
// @Success 200 {object} UserDto "user"
func replaceUserHandler(c *gin.Context) {
	id := c.Param("id")
	user, err := getUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if !ValidateAndBind(c, user) {
		return
	}

	if err := replaceUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUserHandler godoc
// @Router /users/{id} [delete]
// @Accept json
// @Produce json
// @Param id path string true " "
// @Success 200 {object} UserDto "user"
func deleteUserHandler(c *gin.Context) {
	id := c.Param("id")
	user, err := getUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err := deleteUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
