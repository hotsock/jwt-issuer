package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/samber/lo"
)

var SSM issuer.SSMAPI

func main() {
	baseConfig, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	SSM = ssm.NewFromConfig(baseConfig)

	lambda.Start(cfn.LambdaWrap(issuer.CloudFormationHandlerWithLambdaLogging(handler)))
}

func handler(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]any, err error) {
	defer issuer.LogWithTiming(ctx, slog.LevelInfo, "key_generator.handler", "event", event)()

	physicalResourceID = "KeyGenerator"

	switch event.RequestType {
	case cfn.RequestCreate:
		privateKeyPEM, publicKeyPEM := generateKeyPair()
		err = createParameters(ctx, privateKeyPEM, publicKeyPEM)
		data = map[string]any{
			"KeyArn":             "",
			"KeyID":              issuer.ParameterStoreKeyID(),
			"PublicKeyPEMBase64": base64.StdEncoding.EncodeToString(publicKeyPEM),
			"SigningMethod":      "ES256",
		}
		return
	case cfn.RequestUpdate:
		// no-op
		return
	case cfn.RequestDelete:
		deleteParameters(ctx)
		return
	}
	return
}

func generateKeyPair() (privateKeyPEM []byte, publicKeyPEM []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	x509Private, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	privateKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Private})
	x509Public, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	publicKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509Public})

	return privateKeyPEM, publicKeyPEM
}

func createParameters(ctx context.Context, privateKeyPEM []byte, publicKeyPEM []byte) error {
	permittedError := false

	_, err := SSM.PutParameter(ctx, &ssm.PutParameterInput{
		DataType:    lo.ToPtr("text"),
		Description: lo.ToPtr("JWT Issuer Private Key"),
		Name:        lo.ToPtr(issuer.PrivateKeyParameterName()),
		Overwrite:   lo.ToPtr(false),
		Type:        ssmtypes.ParameterTypeSecureString,
		Value:       lo.ToPtr(string(privateKeyPEM)),
	})

	if err != nil {
		var alreadyExists *ssmtypes.AlreadyExistsException
		if !errors.As(err, &alreadyExists) {
			permittedError = true
		}

		if !permittedError {
			return err
		}
	}

	_, err = SSM.PutParameter(ctx, &ssm.PutParameterInput{
		DataType:    lo.ToPtr("text"),
		Description: lo.ToPtr("JWT Issuer Public Key"),
		Name:        lo.ToPtr(issuer.PublicKeyParameterName()),
		Overwrite:   lo.ToPtr(false),
		Type:        ssmtypes.ParameterTypeSecureString,
		Value:       lo.ToPtr(string(publicKeyPEM)),
	})

	if err != nil {
		var alreadyExists *ssmtypes.AlreadyExistsException
		if !errors.As(err, &alreadyExists) {
			permittedError = true
		}
	}

	if !permittedError {
		return err
	}

	return nil
}

func deleteParameters(ctx context.Context) error {
	_, err := SSM.DeleteParameters(ctx, &ssm.DeleteParametersInput{
		Names: []string{
			issuer.PrivateKeyParameterName(),
			issuer.PublicKeyParameterName(),
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	return err
}
