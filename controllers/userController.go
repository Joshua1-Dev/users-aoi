package controllers

import (
	"net/http"
	"hng/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.Param("id")

	var user models.User
	if err := db.First(&user, "user_id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Bad request", "message": "User not found", "statusCode": 404})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User found", "data": user})
}
