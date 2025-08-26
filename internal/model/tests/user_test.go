package tests

import (
	"testing"
	"time"

	"github.com/ypxd99/yandex-gophkeeper/internal/model"
)

func TestUser_Creation(t *testing.T) {
	user := &model.User{
		ID:           "test-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Salt:         "salt",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastLoginAt:  time.Now(),
		IsActive:     true,
	}

	if user.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", user.ID)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", user.Email)
	}

	if !user.IsActive {
		t.Error("Expected user to be active")
	}
}

func TestUserRegistration_Validation(t *testing.T) {
	validReq := &model.UserRegistration{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "validpassword123",
	}

	if validReq.Username == "" {
		t.Error("Username should not be empty")
	}

	if validReq.Email == "" {
		t.Error("Email should not be empty")
	}

	if validReq.Password == "" {
		t.Error("Password should not be empty")
	}

	if len(validReq.Password) < 8 {
		t.Error("Password should be at least 8 characters")
	}
}

func TestUserLogin_Validation(t *testing.T) {
	validReq := &model.UserLogin{
		Username: "testuser",
		Password: "testpassword",
	}

	if validReq.Username == "" {
		t.Error("Username should not be empty")
	}

	if validReq.Password == "" {
		t.Error("Password should not be empty")
	}
}

func TestUserResponse_Fields(t *testing.T) {
	now := time.Now()
	response := &model.UserResponse{
		ID:          "test-id",
		Username:    "testuser",
		Email:       "test@example.com",
		CreatedAt:   now,
		LastLoginAt: now,
	}

	if response.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", response.ID)
	}

	if response.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", response.Username)
	}

	if response.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", response.Email)
	}

	if response.CreatedAt != now {
		t.Error("CreatedAt should match the provided time")
	}

	if response.LastLoginAt != now {
		t.Error("LastLoginAt should match the provided time")
	}
}
