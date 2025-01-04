package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kamva/mgm/v3"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())
	RegisterRoutes(ginEngine)
	return ginEngine
}

func setupTestMongoDB() func() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:5.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp"),
	}
	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start MongoDB container: %v", err)
	}

	host, err := mongoContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get MongoDB host: %v", err)
	}
	port, err := mongoContainer.MappedPort(ctx, "27017")
	if err != nil {
		log.Fatalf("Failed to get MongoDB port: %v", err)
	}
	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())
	if err := mgm.SetDefaultConfig(nil, "test", options.Client().ApplyURI(uri)); err != nil {
		log.Fatalf("Failed to set mgm config: %v", err)
	}

	return func() {
		_ = mongoContainer.Terminate(ctx)
	}
}

type UserStub struct {
	id        string
	name      string
	age       int
	createdAt string
	updatedAt string
}

func mockUser(option UserStub) *User {
	id, _ := primitive.ObjectIDFromHex(option.id)
	createdAt, _ := time.Parse(time.RFC3339, option.createdAt)
	updatedAt, _ := time.Parse(time.RFC3339, option.updatedAt)
	user := &User{
		DefaultModel: mgm.DefaultModel{
			IDField: mgm.IDField{ID: id},
			DateFields: mgm.DateFields{
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		},
		Name: option.name,
		Age:  option.age,
	}
	return user
}
func insertUserDirectly(user *User) error {
	collection := mgm.Coll(user).Collection
	_, err := collection.InsertOne(mgm.Ctx(), user)
	return err
}
func assertInDatabase(t *testing.T, expected *User) {
	var results []User
	err := mgm.Coll(&User{}).SimpleFind(&results, bson.D{})
	assert.NoError(t, err, "Failed to query database")

	found := false
	for _, result := range results {
		if result.ID == expected.ID || result.Name == expected.Name || result.Age == expected.Age {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected object not found in database")
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Age       int    `json:"age"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func TestCreateUserRoute(t *testing.T) {
	gin := setupGin()
	closeMongoDB := setupTestMongoDB()
	defer closeMongoDB()

	t.Run("empty body", func(t *testing.T) {
		payload := map[string]interface{}{}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonPayload))
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("normal", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Jo Liao",
			"age":  22,
		}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonPayload))
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)

		var user UserResponse
		err = json.Unmarshal(res.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Regexp(t, `^[0-9a-f]{24}$`, user.ID)
		assert.Equal(t, "Jo Liao", user.Name)
		assert.Equal(t, 22, user.Age)
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, user.CreatedAt)
		createdAt, err := time.Parse(time.RFC3339, user.CreatedAt)
		assert.NoError(t, err)
		assert.True(t, time.Since(createdAt) < 5*time.Second)
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, user.UpdatedAt)
		updatedAt, err := time.Parse(time.RFC3339, user.UpdatedAt)
		assert.NoError(t, err)
		assert.True(t, time.Since(updatedAt) < 5*time.Second)

		users := []User{}
		err = mgm.Coll(&User{}).SimpleFind(&users, bson.D{})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assertInDatabase(t, mockUser(UserStub{
			id:        user.ID,
			name:      "Jo Liao",
			age:       22,
			createdAt: user.CreatedAt,
			updatedAt: user.UpdatedAt,
		}))
	})
}

func TestGetUserRoute(t *testing.T) {
	gin := setupGin()
	closeMongoDB := setupTestMongoDB()
	defer closeMongoDB()
	err := insertUserDirectly(mockUser(UserStub{
		id:        "123456789012345678901234",
		name:      "Jo Liao",
		age:       22,
		createdAt: "1970-01-01T00:00:00Z",
		updatedAt: "1970-01-01T00:00:00Z",
	}))
	assert.NoError(t, err)

	t.Run("user not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/users/123456789012345678901230", nil)
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("normal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/users/123456789012345678901234", nil)
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		var user UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, "123456789012345678901234", user.ID)
		assert.Equal(t, "Jo Liao", user.Name)
		assert.Equal(t, 22, user.Age)
		assert.Equal(t, "1970-01-01T00:00:00Z", user.CreatedAt)
		assert.Equal(t, "1970-01-01T00:00:00Z", user.UpdatedAt)
	})
}

func TestUpdateUserRoute(t *testing.T) {
	gin := setupGin()
	closeMongoDB := setupTestMongoDB()
	defer closeMongoDB()
	err := insertUserDirectly(mockUser(UserStub{
		id:        "123456789012345678901234",
		name:      "Jo Liao",
		age:       22,
		createdAt: "1970-01-01T00:00:00Z",
		updatedAt: "1970-01-01T00:00:00Z",
	}))
	assert.NoError(t, err)

	t.Run("user not found", func(t *testing.T) {
		payload := map[string]interface{}{}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, _ := http.NewRequest("PATCH", "/users/123456789012345678901230", bytes.NewBuffer(jsonPayload))
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("invalid id format", func(t *testing.T) {
		payload := map[string]interface{}{}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, _ := http.NewRequest("PATCH", "/users/123", bytes.NewBuffer(jsonPayload))
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	t.Run("normal", func(t *testing.T) {
		payload := map[string]interface{}{
			"age": 25,
		}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		res := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/123456789012345678901234", bytes.NewBuffer(jsonPayload))
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		var user UserResponse
		err = json.Unmarshal(res.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Regexp(t, `^[0-9a-f]{24}$`, user.ID)
		assert.Equal(t, "Jo Liao", user.Name)
		assert.Equal(t, 25, user.Age)
		assert.Equal(t, "1970-01-01T00:00:00Z", user.CreatedAt)
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, user.UpdatedAt)
		updatedAt, err := time.Parse(time.RFC3339, user.UpdatedAt)
		assert.NoError(t, err)
		assert.True(t, time.Since(updatedAt) < 5*time.Second)

		users := []User{}
		err = mgm.Coll(&User{}).SimpleFind(&users, bson.D{})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assertInDatabase(t, mockUser(UserStub{
			id:        user.ID,
			name:      "Jo Liao",
			age:       25,
			createdAt: user.CreatedAt,
			updatedAt: user.UpdatedAt,
		}))
	})
}

func TestReplaceUserRoute(t *testing.T) {
	gin := setupGin()
	closeMongoDB := setupTestMongoDB()
	defer closeMongoDB()
	err := insertUserDirectly(mockUser(UserStub{
		id:        "123456789012345678901234",
		name:      "Jo Liao",
		age:       22,
		createdAt: "1970-01-01T00:00:00Z",
		updatedAt: "1970-01-01T00:00:00Z",
	}))
	assert.NoError(t, err)

	t.Run("user not found", func(t *testing.T) {
		payload := map[string]interface{}{}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, _ := http.NewRequest("PUT", "/users/123456789012345678901230", bytes.NewBuffer(jsonPayload))
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("normal", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Alan Lin",
			"age":  25,
		}
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		req, _ := http.NewRequest("PUT", "/users/123456789012345678901234", bytes.NewBuffer(jsonPayload))
		res := httptest.NewRecorder()
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		var user UserResponse
		err = json.Unmarshal(res.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Regexp(t, `^[0-9a-f]{24}$`, user.ID)
		assert.Equal(t, "Alan Lin", user.Name)
		assert.Equal(t, 25, user.Age)
		assert.Equal(t, "1970-01-01T00:00:00Z", user.CreatedAt)
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, user.UpdatedAt)
		updatedAt, err := time.Parse(time.RFC3339, user.UpdatedAt)
		assert.NoError(t, err)
		assert.True(t, time.Since(updatedAt) < 5*time.Second)

		users := []User{}
		err = mgm.Coll(&User{}).SimpleFind(&users, bson.D{})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assertInDatabase(t, mockUser(UserStub{
			id:        user.ID,
			name:      "Alan Lin",
			age:       25,
			createdAt: user.CreatedAt,
			updatedAt: user.UpdatedAt,
		}))
	})
}

func TestDeleteUserRoute(t *testing.T) {
	gin := setupGin()
	closeMongoDB := setupTestMongoDB()
	defer closeMongoDB()
	err := insertUserDirectly(mockUser(UserStub{
		id:        "123456789012345678901234",
		name:      "Jo Liao",
		age:       22,
		createdAt: "1970-01-01T00:00:00Z",
		updatedAt: "1970-01-01T00:00:00Z",
	}))
	assert.NoError(t, err)

	t.Run("user not found", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/users/123456789012345678901230", nil)
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("normal", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/users/123456789012345678901234", nil)
		gin.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		users := []User{}
		err := mgm.Coll(&User{}).SimpleFind(&users, bson.D{})
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}
