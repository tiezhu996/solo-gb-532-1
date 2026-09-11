package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type Claims struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repository *repository.SupportRepository
	secret     []byte
	ttl        time.Duration
}

func NewAuthService(repository *repository.SupportRepository, secret string) *AuthService {
	return &AuthService{repository: repository, secret: []byte(secret), ttl: 8 * time.Hour}
}

func (s *AuthService) Login(request dto.LoginRequest) (dto.LoginResponse, error) {
	user, err := s.repository.UserByUsername(request.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.LoginResponse{}, api.Unauthorized("用户名或密码错误")
		}
		return dto.LoginResponse{}, fmt.Errorf("load login user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return dto.LoginResponse{}, api.Unauthorized("用户名或密码错误")
	}
	now := time.Now().UTC()
	claims := Claims{UserID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{Subject: fmt.Sprint(user.ID), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)), Issuer: "sonar-survey-coverage-planner"}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return dto.LoginResponse{}, fmt.Errorf("sign authentication token: %w", err)
	}
	return dto.LoginResponse{Token: token, User: dto.UserIdentity{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Role: user.Role}}, nil
}

func (s *AuthService) Parse(raw string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing algorithm %s", token.Method.Alg())
		}
		return s.secret, nil
	}, jwt.WithIssuer("sonar-survey-coverage-planner"), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return Claims{}, api.Unauthorized("登录状态已失效")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return Claims{}, api.Unauthorized("登录凭据格式无效")
	}
	user, err := s.repository.UserByID(claims.UserID)
	if err != nil {
		return Claims{}, api.Unauthorized("账号不存在或已停用")
	}
	if user.Role != claims.Role {
		return Claims{}, api.Unauthorized("账号权限已变更，请重新登录")
	}
	return *claims, nil
}
