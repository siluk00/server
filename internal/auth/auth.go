package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	var claims jwt.RegisteredClaims
	claims.Issuer = "chirpy"
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiresIn))
	claims.Subject = userID.String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.UUID{}, err
	}

	uuidStr := claims.Subject
	return uuid.Parse(uuidStr)
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	return getAuthorization(auth, "Bearer")
}

func MakeRefreshToken() (string, error) {
	randomByte := make([]byte, 32)
	_, err := rand.Read(randomByte)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomByte), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	return getAuthorization(auth, "ApiKey")
}

func getAuthorization(auth, key string) (string, error) {
	if auth == "" {
		return "", fmt.Errorf("lacks authorization header")
	}

	authSlc := strings.SplitN(auth, " ", 2)

	if len(authSlc) != 2 || authSlc[0] != key {
		return "", fmt.Errorf("not authorized")
	}

	return authSlc[1], nil
}
