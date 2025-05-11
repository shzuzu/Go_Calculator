package auth

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"-"`
}

type AuthService struct {
	db        *sql.DB
	jwtSecret string
}

type Claims struct {
	ID int64 `json:"id"`
	jwt.RegisteredClaims
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: os.Getenv("JWT_SECRET"), // Читаем секрет из окружения
	}
}

func (s *AuthService) RegisterUser(login, password string) error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE login = ?", login).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 { // Исправлено с count > 1
		return ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO users (login, password) VALUES(?, ?)", login, passwordHash)
	return err
}

func (s *AuthService) LoginUser(login, password string) (string, error) {
	var user User
	err := s.db.QueryRow("SELECT id, login, password FROM users WHERE login = ?", login).Scan(
		&user.ID, &user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidCreds
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCreds
	}

	now := time.Now()
	claims := &Claims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "calculator-api",
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now), // Убрали задержку в 5 секунд
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	return claims.ID, nil
}
