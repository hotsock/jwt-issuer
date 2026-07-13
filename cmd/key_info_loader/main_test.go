package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hotsock/jwt-issuer/internal/mocks"
	"github.com/hotsock/voker/vokercfn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//go:embed cloudformation-input.json
var cloudformationInput []byte

const kmsPublicKeyResponse = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/qZBAS8rW1+QG5BRpMdF/hf+ZsB/QZ0/EOwD5UM2l2Kxxv76RbOTtx3H1ZRP6ppxt/oC5Pvy0p+g+a0WoF4GXQ=="

func Test_handler(t *testing.T) {
	var event vokercfn.Event[keyInfoLoaderProperties]
	require.NoError(t, json.Unmarshal(cloudformationInput, &event))

	mockKMS := mocks.KMSAPI{}
	kmsPublicKey, _ := base64.StdEncoding.DecodeString(kmsPublicKeyResponse)
	kmsOutput := &kms.GetPublicKeyOutput{
		KeyId:     new("arn:aws:kms:us-east-1:111111111111:key/c662cc14-a835-4e28-b6c1-0c77126d98b9"),
		PublicKey: []byte(kmsPublicKey),
	}
	mockKMS.On("GetPublicKey", mock.Anything, mock.Anything).Return(kmsOutput, nil)
	KMS = &mockKMS

	result, err := handler(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, "KeyInfoLoader", result.PhysicalResourceID)
	assert.Equal(t, "arn:aws:kms:us-east-1:111111111111:key/c662cc14-a835-4e28-b6c1-0c77126d98b9", result.Data.KeyArn)
	assert.Equal(t, "c662cc14-a835-4e28-b6c1-0c77126d98b9", result.Data.KeyID)
	assert.Equal(t, "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFL3FaQkFTOHJXMStRRzVCUnBNZEYvaGYrWnNCLwpRWjAvRU93RDVVTTJsMkt4eHY3NlJiT1R0eDNIMVpSUDZwcHh0L29DNVB2eTBwK2crYTBXb0Y0R1hRPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==", result.Data.PublicKeyPEMBase64)
	assert.Equal(t, "ES256", result.Data.SigningMethod)
}
