package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hotsock/jwt-issuer/internal/mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//go:embed cloudformation-input.json
var cloudformationInput []byte

const kmsPublicKeyResponse = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/qZBAS8rW1+QG5BRpMdF/hf+ZsB/QZ0/EOwD5UM2l2Kxxv76RbOTtx3H1ZRP6ppxt/oC5Pvy0p+g+a0WoF4GXQ=="

func Test_handler(t *testing.T) {
	var event cfn.Event
	require.NoError(t, json.Unmarshal(cloudformationInput, &event))

	mockKMS := mocks.KMSAPI{}
	kmsPublicKey, _ := base64.StdEncoding.DecodeString(kmsPublicKeyResponse)
	kmsOutput := &kms.GetPublicKeyOutput{
		KeyId:     lo.ToPtr(event.ResourceProperties["KeyArn"].(string)),
		PublicKey: []byte(kmsPublicKey),
	}
	mockKMS.On("GetPublicKey", mock.Anything, mock.Anything).Return(kmsOutput, nil)
	KMS = &mockKMS

	physicalResourceID, data, err := handler(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, "KeyInfoLoader", physicalResourceID)
	assert.Equal(t, "arn:aws:kms:us-east-1:111111111111:key/c662cc14-a835-4e28-b6c1-0c77126d98b9", data["KeyArn"])
	assert.Equal(t, "c662cc14-a835-4e28-b6c1-0c77126d98b9", data["KeyID"])
	assert.Equal(t, "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFL3FaQkFTOHJXMStRRzVCUnBNZEYvaGYrWnNCLwpRWjAvRU93RDVVTTJsMkt4eHY3NlJiT1R0eDNIMVpSUDZwcHh0L29DNVB2eTBwK2crYTBXb0Y0R1hRPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==", data["PublicKeyPEMBase64"])
	assert.Equal(t, "ES256", data["SigningMethod"])
}
