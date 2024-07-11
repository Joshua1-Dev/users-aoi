package controllers

import (
	"hng/models"
	"hng/utils"
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}


func Register(c *gin.Context) {

	db := c.MustGet("db").(*gorm.DB)

	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": utils.ValidationErrors(err)})
		return
	}


	// hashPassword(input.Password)
	input.UserID = utils.GenerateUUID()
	organisation := models.Organisation{
		OrgID:       utils.GenerateUUID(),
		Name:        input.FirstName + "'s Organisation",
		Description: "Default organisation for " + input.FirstName,
		Users:       []models.User{input},
	}

	if err := db.Create(&organisation).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Bad request", "message": "Registration unsuccessful. email exist", "statusCode": 400})
		return
	}

	token, _ := utils.GenerateToken(input.Email)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Registration successful", "data": gin.H{"accessToken": token, "user": input}})
}


func Login(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": utils.ValidationErrors(err)})
		return
	}

	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Bad request", "message": "Authentication failed", "statusCode": 401})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Bad request", "message": "Authentication failed", "statusCode": 401})
		return
	}

	token, _ := utils.GenerateToken(user.Email)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Login successful", "data": gin.H{"accessToken": token, "user": user}})
}