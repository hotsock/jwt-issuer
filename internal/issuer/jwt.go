package issuer

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type JWTIssuerFunctionInput struct {
	// Whether or not to apply an issued at "iat" claim with the current time.
	// Overrides "iat" in Claims, if true.
	SetIat *bool `json:"setIat,omitempty"`

	// Whether or not to generate an apply a `jti` claim with a generated UUID.
	// Overrides "jti" in Claims, if true.
	SetJti *bool `json:"setJti,omitempty"`

	// Optional number of seconds until the token will expire. Overrides "exp" in
	// Claims, if provided.
	TTL *int64 `json:"ttl,omitempty"`

	// All the claims for the token.
	Claims jwt.MapClaims `json:"claims,omitempty"`
}

type JWTIssuerFunctionOutput struct {
	// The signed JWT.
	Token string `json:"token"`
}

func PrepareToken(input JWTIssuerFunctionInput, keyID string) *jwt.Token {
	if input.Claims == nil {
		input.Claims = jwt.MapClaims{}
	}

	now := time.Now()
	if lo.FromPtr(input.SetIat) {
		input.Claims["iat"] = jwt.NewNumericDate(now)
	}

	if input.TTL != nil {
		input.Claims["exp"] = jwt.NewNumericDate(now.Add(time.Second * time.Duration(lo.FromPtr(input.TTL))))
	}

	if lo.FromPtr(input.SetJti) {
		input.Claims["jti"] = uuid.New().String()
	}

	slog.Debug("issuer.PrepareToken/claims", "claims", input.Claims)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, input.Claims)
	if keyID != "" {
		token.Header["kid"] = keyID
	}

	return token
}
