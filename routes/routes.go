package routes

import (
	"hng/utils"
	"hng/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(r *gin.Engine, db *gorm.DB) {
	auth := r.Group("/auth")
	auth.Use(DbMiddleware(db))
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
}

func UserRoutes(r *gin.Engine, db *gorm.DB) {
	user := r.Group("/api/users")
	user.Use(DbMiddleware(db), authMiddleware())
	{
		user.GET("/:id", controllers.GetUser)
	}
}

func OrganisationRoutes(r *gin.Engine, db *gorm.DB) {
	po := r.Group("/api")
	po.Use(DbMiddleware(db))
	{
		po.POST("organisations/:orgId/users", controllers.AddUserToOrganisation)
	}

	org := r.Group("/api/organisations")
	
	org.Use(DbMiddleware(db), authMiddleware())
	{
		org.GET("/", controllers.GetOrganisations)
		org.GET("/:orgId", controllers.GetOrganisation)
		org.POST("/", controllers.CreateOrganisation)
		
	}
}

func DbMiddleware(db *gorm.DB) gin.HandlerFunc {
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
