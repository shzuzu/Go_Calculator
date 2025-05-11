package repo_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/shzuzu/Go_Calculator/internal/database/repo"
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

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS expressions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		t.Fatalf("Failed to create expressions table: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", "testuser", "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func TestExpressionRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repo.NewRepository(db)

	var userID int64
	err := db.QueryRow("SELECT id FROM users WHERE login = ?", "testuser").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to get test user ID: %v", err)
	}
	id, err := repo.Create(userID, "2+2")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}
	if id <= 0 {
		t.Fatal("Expression ID should be positive")
	}

	expr, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get expression: %v", err)
	}
	if expr == nil {
		t.Fatal("Expression should not be nil")
	}
	if expr.Expression != "2+2" {
		t.Fatalf("Expected expression '2+2', got '%s'", expr.Expression)
	}
	if expr.Status != "pending" {
		t.Fatalf("Expected status 'pending', got '%s'", expr.Status)
	}
	if expr.Result != nil {
		t.Fatal("Result should be nil")
	}

	result := 4.0
	err = repo.UpdateStatus(id, "done", &result)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	expr, err = repo.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get updated expression: %v", err)
	}
	if expr.Status != "done" {
		t.Fatalf("Expected status 'done', got '%s'", expr.Status)
	}
	if expr.Result == nil {
		t.Fatal("Result should not be nil")
	}
	if *expr.Result != 4.0 {
		t.Fatalf("Expected result 4.0, got %f", *expr.Result)
	}

	expressions, err := repo.GetByUserID(userID)
	if err != nil {
		t.Fatalf("Failed to get expressions by user ID: %v", err)
	}
	if len(expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(expressions))
	}
	if expressions[0].Expression != "2+2" {
		t.Fatalf("Expected expression '2+2', got '%s'", expressions[0].Expression)
	}
}
