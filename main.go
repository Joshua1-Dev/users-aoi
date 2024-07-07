package main

import (
	"fmt"
	"os"

	"hng/models"
	"hng/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



func main() {

	// Initialize Database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	db.AutoMigrate(&models.User{}, &models.Organisation{})

	r := gin.Default()

	routes.AuthRoutes(r, db)
	routes.UserRoutes(r, db)
	routes.OrganisationRoutes(r, db)

	r.Run(":10000")
}
