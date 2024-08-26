package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hotsock/jwt-issuer/internal/issuer"
	"github.com/hotsock/jwt-issuer/internal/mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//go:embed cloudformation-input.json
var cloudformationInput []byte

func Test_handler(t *testing.T) {
	os.Setenv(issuer.StackArnEnvVar, "arn:aws:cloudformation:us-east-1:111111111111:stack/JWTIssuer/d9385410-50ee-11ee-b05b-0a236ebfa8d3")

	var event cfn.Event
	json.Unmarshal(cloudformationInput, &event)

	t.Run("update requests no-op", func(t *testing.T) {
		mockSSM := mockedSSM()
		SSM = mockSSM

		event.RequestType = cfn.RequestUpdate
		handler(context.Background(), event)
		mockSSM.AssertNotCalled(t, "PutParameter", mock.Anything, mock.Anything)
	})

	t.Run("delete requests delete parameters", func(t *testing.T) {
		mockSSM := mockedSSM()
		SSM = mockSSM

		event.RequestType = cfn.RequestDelete
		handler(context.Background(), event)
		mockSSM.AssertCalled(t, "DeleteParameters", mock.Anything, mock.Anything)
	})

	t.Run("create requests write key parameters", func(t *testing.T) {
		mockSSM := mockedSSM()
		SSM = mockSSM

		event.RequestType = cfn.RequestCreate
		handler(context.Background(), event)
		mockSSM.AssertNumberOfCalls(t, "PutParameter", 2)

		call1 := mockSSM.Calls[0].Arguments[1].(*ssm.PutParameterInput)
		call2 := mockSSM.Calls[1].Arguments[1].(*ssm.PutParameterInput)

		assert.Equal(t, issuer.PrivateKeyParameterName(), lo.FromPtr(call1.Name))
		assert.Len(t, lo.FromPtr(call1.Value), 247)
		assert.Equal(t, issuer.PublicKeyParameterName(), lo.FromPtr(call2.Name))
		assert.Len(t, lo.FromPtr(call2.Value), 178)
		assert.NotEqual(t, lo.FromPtr(call1.Value), lo.FromPtr(call2.Value))
	})

	t.Run("create requests no-op parameter store write if parameters already exist", func(t *testing.T) {
		mockSSM := mocks.SSMAPI{}
		mockSSM.On("PutParameter", mock.Anything, mock.Anything).Return(nil, &ssmtypes.ParameterAlreadyExists{Message: lo.ToPtr("parameter already exists")})
		SSM = &mockSSM

		event.RequestType = cfn.RequestCreate
		handler(context.Background(), event)
		mockSSM.AssertNumberOfCalls(t, "PutParameter", 2)
	})
}

func mockedSSM() *mocks.SSMAPI {
	mockSSM := mocks.SSMAPI{}
	mockSSM.On("PutParameter", mock.Anything, mock.Anything).Return(nil, nil)
	mockSSM.On("DeleteParameters", mock.Anything, mock.Anything).Return(nil, nil)
	return &mockSSM
}
