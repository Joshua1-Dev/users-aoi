package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hng/controllers"
	"hng/models"
	"hng/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupRouter initializes a Gin router with necessary middleware and routes
func setupRouter() *gin.Engine {

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
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", controllers.Register)
		authGroup.POST("/login", controllers.Login)
	}

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/users/:id", controllers.GetUser)
		apiGroup.GET("/organisations", controllers.GetOrganisations)
		apiGroup.GET("/organisations/:orgId", controllers.GetOrganisation)
		apiGroup.POST("/organisations", controllers.CreateOrganisation)
		apiGroup.POST("/organisations/:orgId/users", controllers.AddUserToOrganisation)
	}

	return r
}

func TestGenerateToken(t *testing.T) {
	email := "test@example.com"
	tokenString, err := utils.GenerateToken(email)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestValidateToken(t *testing.T) {
	email := "test@example.com"
	tokenString, _ := utils.GenerateToken(email)

	claims, err := utils.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, email, claims.Email)
}

func TestTokenExpiry(t *testing.T) {
	email := "test@example.com"
	tokenString, _ := utils.GenerateToken(email)

	claims, _ := utils.ValidateToken(tokenString)

	// Ensure token expires in 24 hours
	expirationTime := time.Unix(claims.ExpiresAt, 0)
	expectedExpirationTime := time.Now().Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExpirationTime, expirationTime, 5*time.Second)
}

func TestRegisterUserSuccess(t *testing.T) {
	router := setupRouter()

	input := map[string]string{
		"firstName": "John",
		"lastName":  "Doe",
		"email":     "john.doe@example.com",
		"password":  "password123",
		"phone":     "1234567890",
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Registration successful", response["message"])

	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["accessToken"])
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "John", user["firstName"])
	assert.Equal(t, "Doe", user["lastName"])
	assert.Equal(t, "john.doe@example.com", user["email"])
	assert.Equal(t, "1234567890", user["phone"])
}

func TestLoginUserSuccess(t *testing.T) {
	router := setupRouter()

	// First register a user
	registerInput := map[string]string{
		"firstName": "John",
		"lastName":  "Doe",
		"email":     "john.doe@example.com",
		"password":  "password123",
		"phone":     "1234567890",
	}
	body, _ := json.Marshal(registerInput)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Then login with the registered user's credentials
	loginInput := map[string]string{
		"email":    "john.doe@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginInput)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Login successful", response["message"])

	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["accessToken"])
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "John", user["firstName"])
	assert.Equal(t, "Doe", user["lastName"])
	assert.Equal(t, "john.doe@example.com", user["email"])
	assert.Equal(t, "1234567890", user["phone"])
}

func TestLoginUserFailure(t *testing.T) {
	router := setupRouter()

	loginInput := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(loginInput)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Bad request", response["status"])
	assert.Equal(t, "Authentication failed", response["message"])
	assert.Equal(t, float64(http.StatusUnauthorized), response["statusCode"].(float64))
}

func TestRegisterUserValidationErrors(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name: "missing firstName",
			input: map[string]string{
				"lastName": "Doe", "email": "john.doe@example.com", "password": "password123", "phone": "1234567890",
			},
			expected: "firstName",
		},
		{
			name: "missing lastName",
			input: map[string]string{
				"firstName": "John", "email": "john.doe@example.com", "password": "password123", "phone": "1234567890",
			},
			expected: "lastName",
		},
		{
			name: "missing email",
			input: map[string]string{
				"firstName": "John", "lastName": "Doe", "password": "password123", "phone": "1234567890",
			},
			expected: "email",
		},
		{
			name: "missing password",
			input: map[string]string{
				"firstName": "John", "lastName": "Doe", "email": "john.doe@example.com", "phone": "1234567890",
			},
			expected: "password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.input)
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			errors := response["errors"].([]interface{})
			assert.Len(t, errors, 1)
			errorItem := errors[0].(map[string]interface{})
			assert.Equal(t, tc.expected, errorItem["field"])
		})
	}
}

func TestRegisterUserDuplicateEmail(t *testing.T) {
	router := setupRouter()

	input := map[string]string{
		"firstName": "John",
		"lastName":  "Doe",
		"email":     "john.doe@example.com",
		"password":  "password123",
		"phone":     "1234567890",
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// First registration
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Second registration with the same email
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errors := response["errors"].([]interface{})
	assert.Len(t, errors, 1)
	errorItem := errors[0].(map[string]interface{})
	assert.Equal(t, "email", errorItem["field"])
	assert.Equal(t, "Email already exists", errorItem["message"])
}

func TestAccessOrganisationData(t *testing.T) {
	router := setupRouter()

	// Register first user
	user1 := map[string]string{
		"firstName": "John",
		"lastName":  "Doe",
		"email":     "john.doe@example.com",
		"password":  "password123",
		"phone":     "1234567890",
	}
	body1, _ := json.Marshal(user1)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response1 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response1)
	token1 := response1["data"].(map[string]interface{})["accessToken"].(string)

	// Register second user
	user2 := map[string]string{
		"firstName": "Jane",
		"lastName":  "Doe",
		"email":     "jane.doe@example.com",
		"password":  "password123",
		"phone":     "0987654321",
	}
	body2, _ := json.Marshal(user2)
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body2))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response2 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response2)
	token2 := response2["data"].(map[string]interface{})["accessToken"].(string)

	// Create an organisation with first user
	org := map[string]string{"name": "John's Organisation"}
	orgBody, _ := json.Marshal(org)
	req, _ = http.NewRequest("POST", "/api/organisations", bytes.NewBuffer(orgBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var orgResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &orgResponse)
	orgID := orgResponse["data"].(map[string]interface{})["id"].(float64)

	// Try to access the organisation with second user
	req, _ = http.NewRequest("GET", "/api/organisations/"+strconv.Itoa(int(orgID)), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var accessResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &accessResponse)
	assert.Equal(t, "forbidden", accessResponse["status"])
	assert.Equal(t, "You do not have access to this organisation", accessResponse["message"])
}
