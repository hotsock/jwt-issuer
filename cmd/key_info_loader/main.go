package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/samber/lo"
)

var KMS issuer.KMSAPI

func main() {
	baseConfig, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	KMS = kms.NewFromConfig(baseConfig)

	lambda.Start(cfn.LambdaWrap(issuer.CloudFormationHandlerWithLambdaLogging(handler)))
}

func handler(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]any, err error) {
	defer issuer.LogWithTiming(ctx, slog.LevelInfo, "key_info_loader.handler", "event", event)()

	physicalResourceID = "KeyInfoLoader"

	publicKeyOutput, err := KMS.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: lo.ToPtr(os.Getenv("SIGNING_KEY_ARN")),
	})

	if err != nil {
		return
	}

	keyArn := lo.FromPtr(publicKeyOutput.KeyId)
	arnParts := strings.Split(keyArn, "/")
	keyID := arnParts[1]
	publicKey, _ := x509.ParsePKIXPublicKey(publicKeyOutput.PublicKey)
	x509Public, _ := x509.MarshalPKIXPublicKey(publicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509Public})
	publicKeyPEMBase64 := base64.StdEncoding.EncodeToString(publicKeyPEM)

	data = map[string]any{
		"KeyArn":             keyArn,
		"KeyID":              keyID,
		"PublicKeyPEMBase64": publicKeyPEMBase64,
		"SigningMethod":      "ES256",
	}

	return
}
