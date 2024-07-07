package controllers

import (
	"hng/models"
	"hng/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterInput struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	Phone     string `json:"phone"`
}

var validate *validator.Validate

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": utils.ValidationErrors(err),
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// Check if email already exists
	var existingUser models.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": []utils.ValidationError{
				{Field: "email", Message: "Email already exists"},
			},
		})
		return
	}

	user := models.User{
		UserID:    utils.GenerateUUID(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  hashPassword(input.Password),
		Phone:     input.Phone,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a default organisation for the user
	organisation := models.Organisation{
		OrgID:       utils.GenerateUUID(),
		Name:        input.FirstName + "'s Organisation",
		Description: input.FirstName + "'s personal organisation",
	}
	db.Create(&organisation)

	// Generate JWT token
	token, err := utils.GenerateToken(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registration successful",
		"data": gin.H{
			"accessToken": token,
			"user": gin.H{
				"userId":    user.UserID,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"email":     user.Email,
				"phone":     user.Phone,
			},
		},
	})
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
