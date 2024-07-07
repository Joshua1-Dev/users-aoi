package routes

import (
	"hng/utils"
	"hng/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(r *gin.Engine, db *gorm.DB) {
	auth := r.Group("/auth")
	auth.Use(dbMiddleware(db))
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
}

func UserRoutes(r *gin.Engine, db *gorm.DB) {
	user := r.Group("/api/users")
	user.Use(dbMiddleware(db), authMiddleware())
	{
		user.GET("/:id", controllers.GetUser)
	}
}

func OrganisationRoutes(r *gin.Engine, db *gorm.DB) {
	org := r.Group("/api/organisations")
	org.Use(dbMiddleware(db), authMiddleware())
	{
		org.GET("/", controllers.GetOrganisations)
		org.GET("/:orgId", controllers.GetOrganisation)
		org.POST("/", controllers.CreateOrganisation)
		org.POST("/:orgId/users", controllers.AddUserToOrganisation)
	}
}

func dbMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"status": "unauthorized", "message": "Missing authorization header"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"status": "unauthorized", "message": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Next()
	}
}
