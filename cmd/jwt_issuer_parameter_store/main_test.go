package main

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed ec256-private.pem
var privateKeyPEM []byte

//go:embed ec256-public.pem
var publicKeyPEM []byte

func Test_handler(t *testing.T) {
	keyID = "4a2c1b37"

	claims := jwt.MapClaims{
		"foo": "bar",
	}

	privateKeyObj, err := jwt.ParseECPrivateKeyFromPEM(privateKeyPEM)
	require.NoError(t, err)
	privateKey = privateKeyObj

	publicKeyObj, err := jwt.ParseECPublicKeyFromPEM(publicKeyPEM)
	require.NoError(t, err)

	output, err := handler(context.Background(), issuer.JWTIssuerFunctionInput{Claims: claims, TTL: lo.ToPtr(time.Duration(60)), SetIat: lo.ToPtr(true), SetJti: lo.ToPtr(true)})
	require.NoError(t, err)

	generatedToken, err := jwt.Parse(output.Token, func(t *jwt.Token) (any, error) {
		return publicKeyObj, nil
	}, jwt.WithValidMethods([]string{"ES256"}))

	require.NoError(t, err)

	generatedClaims := generatedToken.Claims.(jwt.MapClaims)
	assert.Equal(t, keyID, generatedToken.Header["kid"])
	assert.Equal(t, "bar", generatedClaims["foo"])
	assert.Greater(t, generatedClaims["exp"], float64(time.Now().Unix()))
	assert.LessOrEqual(t, generatedClaims["iat"], float64(time.Now().Unix()))
	assert.Len(t, generatedClaims["jti"], 36)
}
