package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/hotsock/voker"
	"github.com/hotsock/voker/vokercfn"
	"github.com/hotsock/voker/vokerslog"
	"github.com/samber/lo"
)

var KMS issuer.KMSAPI

// keyInfoLoaderProperties is the CloudFormation custom resource's
// ResourceProperties. The key info loader does not take any input properties.
type keyInfoLoaderProperties struct{}

// keyInfoLoaderData is the CloudFormation custom resource's Data, exposed to
// the stack through Fn::GetAtt.
type keyInfoLoaderData struct {
	KeyArn             string `json:"KeyArn"`
	KeyID              string `json:"KeyID"`
	PublicKeyPEMBase64 string `json:"PublicKeyPEMBase64"`
	SigningMethod      string `json:"SigningMethod"`
}

func main() {
	logger := slog.New(vokerslog.NewHandler(os.Stdout))
	slog.SetDefault(logger)

	baseConfig, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	KMS = kms.NewFromConfig(baseConfig)

	vokercfn.Start(handler, voker.WithLogger(logger))
}

func handler(ctx context.Context, event vokercfn.Event[keyInfoLoaderProperties]) (result vokercfn.Result[keyInfoLoaderData], err error) {
	defer issuer.LogWithTiming(ctx, slog.LevelInfo, "key_info_loader.handler", "event", event)()

	result.PhysicalResourceID = "KeyInfoLoader"

	publicKeyOutput, err := KMS.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: new(os.Getenv("SIGNING_KEY_ARN")),
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

	result.Data = keyInfoLoaderData{
		KeyArn:             keyArn,
		KeyID:              keyID,
		PublicKeyPEMBase64: publicKeyPEMBase64,
		SigningMethod:      "ES256",
	}

	return
}
