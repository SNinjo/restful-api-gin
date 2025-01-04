package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	group := r.Group("/users")
	{
		group.POST("", createUserHandler)
		group.GET("/:id", getUserHandler)
		group.PATCH("/:id", updateUserHandler)
		group.PUT("/:id", replaceUserHandler)
		group.DELETE("/:id", deleteUserHandler)
	}
}
