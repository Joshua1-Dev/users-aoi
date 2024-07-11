package controllers

import (
	"hng/models"
	"hng/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetOrganisations(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var organisations []models.Organisation
	if err := db.Find(&organisations).Error; err != nil {
		log.Printf("Error retrieving organisations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Internal server error", "message": "Could not retrieve organisations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Organisations found", "data": gin.H{"organisations": organisations}})
}

func GetOrganisation(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	orgID := c.Param("orgId")

	var organisation models.Organisation
	if err := db.First(&organisation, "org_id = ?", orgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Bad request", "message": "Organisation not found", "statusCode": 404})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Organisation found", "data": organisation})
}

func CreateOrganisation(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var input models.Organisation

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": utils.ValidationErrors(err)})
		return
	}

	input.OrgID = utils.GenerateUUID()
	userID := c.MustGet("userId").(string)

	var user models.User
	db.First(&user, "user_id = ?", userID)

	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Bad request", "message": "Organisation creation unsuccessful", "statusCode": 400})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Organisation created successfully", "data": input})
}

func AddUserToOrganisation(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	orgID := c.Param("orgId")

	var input struct {
		UserID string `json:"userId"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": utils.ValidationErrors(err)})
		return
	}

	var user models.User
	if err := db.First(&user, "user_id = ?", input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Bad request", "message": "User not found", "statusCode": 404})
		return
	}

	var organisation models.Organisation
	if err := db.First(&organisation, "org_id = ?", orgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Bad request", "message": "Organisation not found", "statusCode": 404})
		return
	}

	db.Model(&organisation).Association("Users").Append(&user)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User added to organisation successfully"})
}
