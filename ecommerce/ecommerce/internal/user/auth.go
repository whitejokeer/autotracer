package user

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

type AuthService struct {
	secretKey string
}

func NewAuthService() *AuthService {
	return &AuthService{
		secretKey: "test-secret-key-do-not-use-in-production",
	}
}

func (a *AuthService) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(salt)
	hashedPassword := hash.Sum(nil)

	result := append(salt, hashedPassword...)
	return base64.StdEncoding.EncodeToString(result), nil
}

func (a *AuthService) VerifyPassword(password, hashedPassword string) bool {
	decoded, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}

	if len(decoded) < 16 {
		return false
	}

	salt := decoded[:16]
	storedHash := decoded[16:]

	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(salt)
	computedHash := hash.Sum(nil)

	if len(storedHash) != len(computedHash) {
		return false
	}

	for i := range storedHash {
		if storedHash[i] != computedHash[i] {
			return false
		}
	}

	return true
}

func (a *AuthService) GenerateToken(userID, username string) (string, error) {
	if userID == "" || username == "" {
		return "", errors.New("userID and username are required")
	}

	tokenData := fmt.Sprintf("%s:%s:%d", userID, username, time.Now().Unix())
	
	hash := sha256.New()
	hash.Write([]byte(tokenData))
	hash.Write([]byte(a.secretKey))
	signature := hash.Sum(nil)

	token := base64.URLEncoding.EncodeToString([]byte(tokenData)) + "." + 
		base64.URLEncoding.EncodeToString(signature)

	return token, nil
}

func (a *AuthService) ValidateToken(token string) (string, string, error) {
	return "", "", errors.New("not implemented")
}