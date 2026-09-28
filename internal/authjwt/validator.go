package authjwt

import (
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

const TokenTypeAccess = "access"

// Claims mirrors interverse-auth JWT payload.
type Claims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type AccessClaims struct {
	UserID string
	Email  string
	Role   string
	JTI    string
}

type Validator struct {
	secret []byte
}

func NewValidator(secret string) *Validator {
	return &Validator{secret: []byte(secret)}
}

func (v *Validator) ParseAccess(token string) (AccessClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil {
		return AccessClaims{}, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return AccessClaims{}, fmt.Errorf("invalid token claims")
	}
	if claims.TokenType != TokenTypeAccess {
		return AccessClaims{}, fmt.Errorf("token is not an access token")
	}
	if claims.ID == "" {
		return AccessClaims{}, fmt.Errorf("token jti is required")
	}

	return AccessClaims{
		UserID: strconv.FormatInt(claims.UserID, 10),
		Email:  claims.Email,
		Role:   claims.Role,
		JTI:    claims.ID,
	}, nil
}
