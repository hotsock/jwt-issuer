package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/hotsock/voker"
	"github.com/hotsock/voker/vokerslog"
)

var KMS issuer.KMSAPI
var signingKeyArn string
var keyID string

func main() {
	logger := slog.New(vokerslog.NewHandler(os.Stdout))
	slog.SetDefault(logger)

	baseConfig, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	KMS = kms.NewFromConfig(baseConfig)

	signingKeyArn = os.Getenv("SIGNING_KEY_ARN")

	arnParts := strings.Split(signingKeyArn, "/")
	if len(arnParts) == 2 {
		keyID = arnParts[1]
	}

	voker.Start(handler, voker.WithLogger(logger))
}

func handler(ctx context.Context, input issuer.JWTIssuerFunctionInput) (issuer.JWTIssuerFunctionOutput, error) {
	defer issuer.LogWithTiming(ctx, slog.LevelDebug, "jwt_issuer_kms.handler", "input", input)()

	token := issuer.PrepareToken(input, keyID)

	signedToken, err := issuer.SignJWTWithKMS(ctx, KMS, token, signingKeyArn)
	if err != nil {
		return issuer.JWTIssuerFunctionOutput{}, err
	}

	return issuer.JWTIssuerFunctionOutput{
		Token: signedToken,
	}, nil
}
