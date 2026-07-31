package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"myapi/internal/model"
	"myapi/internal/repository"
	"myapi/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	Username string `json:"username"`
}

var errInvalidRefreshToken = fmt.Errorf("invalid refresh token")

type Login struct {
	UserRepo                  *repository.User
	RefreshTokenRepo          *repository.RefreshToken
	JWTSecret                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
}

func (s *Login) AuthUser(username, password string) (*model.User, bool, error) {
	user, err := s.UserRepo.GetByUsername(username)
	if err != nil {
		return nil, false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return user, err == nil, nil
}

func (s *Login) GenerateTokens(user *model.User) (string, string, int, error) {
	now := time.Now()
	accessToken, err := s.generateJWT(user, now)
	if err != nil {
		return "", "", 0, err
	}

	refreshRaw, err := utils.GenerateRandomString(32)
	if err != nil {
		return "", "", 0, err
	}

	hash := hashToken(refreshRaw)
	expiresAt := now.Add(time.Duration(s.RefreshTokenExpireSeconds) * time.Second)
	if err := s.RefreshTokenRepo.Create(user.ID, hash, expiresAt); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshRaw, s.AccessTokenExpireSeconds, nil
}

func (s *Login) RefreshTokens(refreshRaw string) (string, string, int, error) {
	hash := hashToken(refreshRaw)
	rt, err := s.RefreshTokenRepo.FindByHash(hash)
	if err != nil {
		return "", "", 0, fmt.Errorf("refresh token: %w", err)
	}
	if rt.ID == bson.NilObjectID {
		return "", "", 0, errInvalidRefreshToken
	}
	if time.Now().After(rt.ExpiresAt) {
		s.RefreshTokenRepo.Delete(rt.ID)
		return "", "", 0, errInvalidRefreshToken
	}

	s.RefreshTokenRepo.Delete(rt.ID)

	user, err := s.UserRepo.GetByID(rt.UserID)
	if err != nil {
		return "", "", 0, err
	}

	return s.GenerateTokens(user)
}

func (s *Login) Logout(refreshRaw string) error {
	hash := hashToken(refreshRaw)
	rt, err := s.RefreshTokenRepo.FindByHash(hash)
	if err != nil || rt.ID == bson.NilObjectID {
		return nil
	}
	return s.RefreshTokenRepo.Delete(rt.ID)
}

func (s *Login) RevokeAllUserTokens(userID bson.ObjectID) error {
	return s.RefreshTokenRepo.DeleteAllByUserID(userID)
}

func (s *Login) ParseJWT(tokenStr string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (s *Login) generateJWT(user *model.User, now time.Time) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.AccessTokenExpireSeconds) * time.Second)),
		},
		UserID:   user.ID.Hex(),
		Username: user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.JWTSecret))
}

func hashToken(raw string) string {
	b := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(b[:])
}
