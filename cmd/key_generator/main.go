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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/hotsock/voker"
	"github.com/hotsock/voker/vokercfn"
	"github.com/hotsock/voker/vokerslog"
)

var SSM issuer.SSMAPI

// keyGeneratorProperties is the CloudFormation custom resource's
// ResourceProperties. The key generator does not take any input properties.
type keyGeneratorProperties struct{}

// keyGeneratorData is the CloudFormation custom resource's Data, exposed to
// the stack through Fn::GetAtt.
type keyGeneratorData struct {
	KeyArn             string `json:"KeyArn"`
	KeyID              string `json:"KeyID"`
	PublicKeyPEMBase64 string `json:"PublicKeyPEMBase64"`
	SigningMethod      string `json:"SigningMethod"`
}

func main() {
	logger := slog.New(vokerslog.NewHandler(os.Stdout))
	slog.SetDefault(logger)

	baseConfig, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	SSM = ssm.NewFromConfig(baseConfig)

	vokercfn.Start(handler, voker.WithLogger(logger))
}

func handler(ctx context.Context, event vokercfn.Event[keyGeneratorProperties]) (result vokercfn.Result[keyGeneratorData], err error) {
	defer issuer.LogWithTiming(ctx, slog.LevelInfo, "key_generator.handler", "event", event)()

	result.PhysicalResourceID = "KeyGenerator"

	switch event.RequestType {
	case vokercfn.RequestCreate:
		privateKeyPEM, publicKeyPEM := generateKeyPair()
		err = createParameters(ctx, privateKeyPEM, publicKeyPEM)
		result.Data = keyGeneratorData{
			KeyArn:             "",
			KeyID:              issuer.ParameterStoreKeyID(),
			PublicKeyPEMBase64: base64.StdEncoding.EncodeToString(publicKeyPEM),
			SigningMethod:      "ES256",
		}
		return
	case vokercfn.RequestUpdate:
		// no-op
		return
	case vokercfn.RequestDelete:
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
		DataType:    new("text"),
		Description: new("JWT Issuer Private Key"),
		Name:        new(issuer.PrivateKeyParameterName()),
		Overwrite:   new(false),
		Type:        ssmtypes.ParameterTypeSecureString,
		Value:       new(string(privateKeyPEM)),
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
		DataType:    new("text"),
		Description: new("JWT Issuer Public Key"),
		Name:        new(issuer.PublicKeyParameterName()),
		Overwrite:   new(false),
		Type:        ssmtypes.ParameterTypeSecureString,
		Value:       new(string(publicKeyPEM)),
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
