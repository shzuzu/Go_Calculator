package auth

import (
	"database/sql"
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
	db *sql.DB
}
type Claims struct {
	ID int64 `json:"id"`
	jwt.MapClaims
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) RegisterUser(login, password string) error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE login = ?", login).Scan(&count)

	if err != nil {
		return err
	}
	if count > 1 {
		return ErrUserAlreadyExists
	}

	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (login, password) VALUES(?, ?)", login, password_hash)

	if err != nil {
		return err
	}

	return nil
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

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return "", ErrInvalidCreds
		}
		return "", err
	}

	now := time.Now()
	claims := &Claims{
		ID: user.ID,
		MapClaims: jwt.MapClaims{
			"name": user.ID,
			"nbf":  now.Add(5 * time.Second).Unix(),
			"exp":  now.Add(24 * time.Hour).Unix(),
			"iat":  now.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret-from-env"))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return "secret-from-env", nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, ErrInvalidToken
	}

	return claims.ID, nil
}
