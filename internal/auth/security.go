package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the provided password using bcrypt default cost.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword compares a bcrypt hashed password with its possible plaintext equivalent.
func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Claims defines the JWT payload used for authentication.
type Claims struct {
	Email string `json:"email"`
	Plan  string `json:"plan"`
	jwt.RegisteredClaims
}

// MakeJWT creates a signed JWT for the provided user.
func MakeJWT(cfg Config, user User) (string, error) {
	if cfg.JWTSecret == "" {
		return "", errors.New("jwt secret is required")
	}

	ttl := cfg.JWTTTLHours
	if ttl <= 0 {
		ttl = 24
	}

	now := time.Now()
	claims := Claims{
		Email: user.Email,
		Plan:  user.PlanID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    "estimaGO",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseJWT validates the provided token string and returns the claims.
func ParseJWT(tokenString string, secret string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
