package main

import (
	"context"
	"crypto/ecdsa"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/samber/lo"
)

var SSM issuer.SSMAPI
var privateKey *ecdsa.PrivateKey
var keyID string

func main() {
	baseConfig, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	SSM = ssm.NewFromConfig(baseConfig)

	getParamResponse, err := SSM.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           lo.ToPtr(issuer.PrivateKeyParameterName()),
		WithDecryption: lo.ToPtr(true),
	})
	if err != nil {
		panic(err)
	}

	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(lo.FromPtr(getParamResponse.Parameter.Value)))
	if err != nil {
		panic(err)
	}

	privateKey = key
	keyID = issuer.ParameterStoreKeyID()

	lambda.StartHandlerFunc(issuer.HandlerWithLambdaLogging(handler))
}

func handler(ctx context.Context, input issuer.JWTIssuerFunctionInput) (issuer.JWTIssuerFunctionOutput, error) {
	defer issuer.LogWithTiming(ctx, slog.LevelDebug, "jwt_issuer_parameter_store.handler", "input", input)()

	token := issuer.PrepareToken(input, keyID)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return issuer.JWTIssuerFunctionOutput{}, err
	}

	return issuer.JWTIssuerFunctionOutput{
		Token: signedToken,
	}, nil
}
