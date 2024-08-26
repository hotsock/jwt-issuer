package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	_ "embed"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/hotsock/jwt-issuer/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//go:embed ec256-private.pem
var privateKeyPEM []byte

//go:embed ec256-public.pem
var publicKeyPEM []byte

func Test_handler(t *testing.T) {
	signingKeyArn = "arn:aws:kms:us-east-1:111111111111:key/4a2c1b37-e4c8-466a-b873-11aaf144b01b"
	keyID = "4a2c1b37-e4c8-466a-b873-11aaf144b01b"

	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	privateKeyObj, err := jwt.ParseECPrivateKeyFromPEM(privateKeyPEM)
	require.NoError(t, err)

	publicKeyObj, err := jwt.ParseECPublicKeyFromPEM(publicKeyPEM)
	require.NoError(t, err)

	h := crypto.SHA256.New()
	signingString, _ := token.SigningString()
	h.Write([]byte(signingString))
	signature, _ := ecdsa.SignASN1(rand.Reader, privateKeyObj, h.Sum(nil))

	mockKMS := mocks.KMSAPI{}
	mockKMS.On("Sign", mock.Anything, mock.Anything).Return(&kms.SignOutput{Signature: signature}, nil)
	KMS = &mockKMS

	output, err := handler(context.Background(), issuer.JWTIssuerFunctionInput{Claims: claims})
	require.NoError(t, err)

	_, err = jwt.Parse(output.Token, func(t *jwt.Token) (any, error) {
		return publicKeyObj, nil
	}, jwt.WithValidMethods([]string{"ES256"}))

	require.NoError(t, err)
}
