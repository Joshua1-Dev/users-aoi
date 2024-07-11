package main

import (
	"fmt"
	"hng/models"
	"hng/routes"
	"os"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	// Form the connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: false})
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
