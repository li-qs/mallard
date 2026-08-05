package user

import (
	"context"
	"errors"
	"fmt"
	"mallard/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo  UserRepoImpl
	tokenRepo TokenRepoImpl
	options   *ServiceOptions
}

type ServiceOptions struct {
	JWTSecret                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	Username string `json:"username"`
}

func NewService(userRepo UserRepoImpl, tokenRepo TokenRepoImpl, opts *ServiceOptions) *Service {
	return &Service{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		options:   opts,
	}
}

func (s *Service) AuthUser(ctx context.Context, username, password string) (*UserEntity, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) GenerateTokens(ctx context.Context, user *UserEntity) (string, string, int, error) {
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
	expiresAt := now.Add(time.Duration(s.options.RefreshTokenExpireSeconds) * time.Second)
	if err := s.tokenRepo.Create(ctx, user.ID, hash, expiresAt); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshRaw, s.options.AccessTokenExpireSeconds, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshRaw string) (string, string, int, error) {
	tokenHash := hashToken(refreshRaw)
	rt, err := s.tokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", 0, ErrInvalidRefreshToken
		}
		return "", "", 0, fmt.Errorf("find refresh token: %w", err)
	}
	if time.Now().After(rt.ExpiresAt) {
		s.tokenRepo.Delete(ctx, rt.ID)
		return "", "", 0, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", 0, err
	}

	accessToken, newRefreshRaw, expiresIn, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return "", "", 0, err
	}

	_ = s.tokenRepo.Delete(ctx, rt.ID)

	return accessToken, newRefreshRaw, expiresIn, nil
}

func (s *Service) Logout(ctx context.Context, refreshRaw string) error {
	tokenHash := hashToken(refreshRaw)
	rt, err := s.tokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return err
	}
	return s.tokenRepo.Delete(ctx, rt.ID)
}

func (s *Service) RevokeAllUserTokens(ctx context.Context, userID bson.ObjectID) error {
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}

func (s *Service) generateJWT(user *UserEntity, now time.Time) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.options.AccessTokenExpireSeconds) * time.Second)),
		},
		UserID:   user.ID.Hex(),
		Username: user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.options.JWTSecret))
}

func hashToken(raw string) string {
	// TODO: 加盐
	return utils.SHA256Hex(raw)
}

func (s *Service) GetByID(ctx context.Context, userID bson.ObjectID) (*UserEntity, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) GetByUsername(ctx context.Context, username string) (*UserEntity, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *Service) UpdatePassword(ctx context.Context, userID bson.ObjectID, pwd string) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePasswordHash(ctx, userID, string(pwdHash))
}
