package auth_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/shzuzu/Go_Calculator/internal/auth"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func TestRegisterAndLoginUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	authService := auth.NewAuthService(db)

	err := authService.RegisterUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	err = authService.RegisterUser("testuser", "password123")
	if err != auth.ErrUserAlreadyExists {
		t.Fatalf("Expected ErrUserExists, got: %v", err)
	}

	token, err := authService.LoginUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	if token == "" {
		t.Fatal("Token should not be empty")
	}

	_, err = authService.LoginUser("testuser", "wrongpassword")
	if err != auth.ErrInvalidCreds {
		t.Fatalf("Expected ErrInvalidCredentials, got: %v", err)
	}

	_, err = authService.LoginUser("nonexistent", "password123")
	if err != auth.ErrInvalidCreds {
		t.Fatalf("Expected ErrInvalidCredentials, got: %v", err)
	}

	userID, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if userID <= 0 {
		t.Fatal("User ID should be positive")
	}
}
