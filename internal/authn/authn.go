package authn

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenUnknownType = errors.New("unknown token type")
)

type TokenService struct {
	Conf *config.ClientToken
}

func NewTokenService(conf *config.ClientToken) *TokenService {
	return &TokenService{Conf: conf}
}

type ClientAPIToken struct {
	jwt.RegisteredClaims
}

func (ts *TokenService) IssueToken() (string, error) {
	now := time.Now().UTC()
	conf := ts.Conf
	expiresAt := now.Add(conf.TTL)
	issuedAt := now
	issuer := conf.Issuer
	tokenID := uuid.New()
	claims := &ClientAPIToken{
		jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ID:        tokenID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := conf.Secret

	return token.SignedString([]byte(secret))
}

type TokenValidator interface {
	ValidateToken(clientJWT string) (*ClientAPIToken, error)
}

func (ts *TokenService) ValidateToken(clientJWT string) (*ClientAPIToken, error) {
	conf := ts.Conf
	token, err := jwt.ParseWithClaims(clientJWT, &ClientAPIToken{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parse failed: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*ClientAPIToken)
	if !ok {
		return nil, ErrTokenUnknownType
	}

	return claims, nil
}
