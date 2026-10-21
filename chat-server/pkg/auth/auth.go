package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "

	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	SubjectTypeUser     = "user"
	SubjectTypeEmployee = "employee"

	MaxTokenLen = 1024
)

var (
	ErrInvalidToken       = errors.New("无效的令牌")
	ErrExpiredToken       = errors.New("令牌已过期")
	ErrInvalidTokenType   = errors.New("令牌类型错误")
	ErrInvalidSubjectType = errors.New("令牌主体类型错误")
)

type Config struct {
	Secret            string        `mapstructure:"secret"`
	Expiration        time.Duration `mapstructure:"expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
	Issuer            string        `mapstructure:"issuer"`
}

func (c *Config) Validate() error {
	secret := strings.TrimSpace(c.Secret)
	if secret == "" {
		return fmt.Errorf("auth.secret 未配置")
	}
	if c.Expiration <= 0 {
		return fmt.Errorf("auth.expiration 未配置")
	}
	if c.RefreshExpiration <= 0 {
		return fmt.Errorf("auth.refresh_expiration 未配置")
	}
	if c.RefreshExpiration <= c.Expiration {
		return fmt.Errorf("auth.refresh_expiration 必须大于 auth.expiration")
	}
	if strings.TrimSpace(c.Issuer) == "" {
		return fmt.Errorf("auth.issuer 未配置")
	}
	return nil
}

type Claims struct {
	SubjectType string `json:"sub_type"`
	SubjectID   uint   `json:"sub_id"`
	TokenType   string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	RefreshJTI   string
}

type Auth struct {
	config *Config
}

func NewAuth(config *Config) *Auth {
	return &Auth{config: config}
}

func (a *Auth) RefreshTTL() time.Duration {
	return a.config.RefreshExpiration
}

func (a *Auth) GenerateAccessToken(subjectType string, subjectID uint) (string, error) {
	return a.sign(subjectType, subjectID, TokenTypeAccess, a.config.Expiration, "")
}

func (a *Auth) generateRefreshToken(subjectType string, subjectID uint) (string, string, error) {
	jti := uuid.NewString()
	token, err := a.sign(subjectType, subjectID, TokenTypeRefresh, a.config.RefreshExpiration, jti)
	return token, jti, err
}

func (a *Auth) GenerateTokenPair(subjectType string, subjectID uint) (*TokenPair, error) {
	access, err := a.GenerateAccessToken(subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	refresh, jti, err := a.generateRefreshToken(subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, RefreshJTI: jti}, nil
}

func (a *Auth) sign(subjectType string, subjectID uint, tokenType string, ttl time.Duration, jti string) (string, error) {
	now := time.Now()
	claims := Claims{
		SubjectType: subjectType,
		SubjectID:   subjectID,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.config.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.Secret))
}

func (a *Auth) ParseAccessToken(tokenString string) (*Claims, error) {
	return a.parseToken(tokenString, TokenTypeAccess)
}

func (a *Auth) ParseRefreshToken(tokenString string) (*Claims, error) {
	return a.parseToken(tokenString, TokenTypeRefresh)
}

func (a *Auth) parseToken(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(a.config.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}
	if claims.SubjectType != SubjectTypeUser && claims.SubjectType != SubjectTypeEmployee {
		return nil, ErrInvalidSubjectType
	}
	return claims, nil
}

func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > len(BearerPrefix) && authHeader[:len(BearerPrefix)] == BearerPrefix {
		return authHeader[len(BearerPrefix):]
	}
	return ""
}
