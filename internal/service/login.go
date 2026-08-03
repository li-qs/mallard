package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mallard/internal/model"
	"mallard/internal/repository"
	"mallard/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	Username string `json:"username"`
}

var (
	ErrInvalidCredentials  = fmt.Errorf("invalid credentials")
	ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
)

type Login struct {
	UserRepo                  *repository.User
	RefreshTokenRepo          *repository.RefreshToken
	JWTSecret                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
}

func (s *Login) AuthUser(ctx context.Context, username, password string) (*model.User, bool, error) {
	user, err := s.UserRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, ErrInvalidCredentials
		}
		return nil, false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, false, ErrInvalidCredentials
	}
	return user, true, nil
}

func (s *Login) GenerateTokens(ctx context.Context, user *model.User) (string, string, int, error) {
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
	if err := s.RefreshTokenRepo.Create(ctx, user.ID, hash, expiresAt); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshRaw, s.AccessTokenExpireSeconds, nil
}

func (s *Login) RefreshTokens(ctx context.Context, refreshRaw string) (string, string, int, error) {
	hash := hashToken(refreshRaw)
	rt, err := s.RefreshTokenRepo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", 0, ErrInvalidRefreshToken
		}
		return "", "", 0, fmt.Errorf("find refresh token: %w", err)
	}
	if time.Now().After(rt.ExpiresAt) {
		s.RefreshTokenRepo.Delete(ctx, rt.ID)
		return "", "", 0, ErrInvalidRefreshToken
	}

	s.RefreshTokenRepo.Delete(ctx, rt.ID)

	user, err := s.UserRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", 0, err
	}

	return s.GenerateTokens(ctx, user)
}

func (s *Login) Logout(ctx context.Context, refreshRaw string) error {
	hash := hashToken(refreshRaw)
	rt, err := s.RefreshTokenRepo.FindByHash(ctx, hash)
	if err != nil || rt.ID == bson.NilObjectID {
		return nil
	}
	return s.RefreshTokenRepo.Delete(ctx, rt.ID)
}

func (s *Login) RevokeAllUserTokens(ctx context.Context, userID bson.ObjectID) error {
	return s.RefreshTokenRepo.DeleteAllByUserID(ctx, userID)
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
	return utils.SHA256Hex(raw)
}
