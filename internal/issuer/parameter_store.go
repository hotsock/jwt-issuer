package issuer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const StackArnEnvVar = "STACK_ARN"

type SSMAPI interface {
	DeleteParameters(context.Context, *ssm.DeleteParametersInput, ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error)
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	PutParameter(context.Context, *ssm.PutParameterInput, ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
}

func ParameterStoreKeyID() string {
	stackArn, _ := arn.Parse(os.Getenv(StackArnEnvVar))
	return strings.Split(stackArn.Resource, "/")[2]
}

func PrivateKeyParameterName() string {
	return fmt.Sprintf("%s/private-key", parameterNamePrefix())
}

func PublicKeyParameterName() string {
	return fmt.Sprintf("%s/public-key", parameterNamePrefix())
}

func parameterNamePrefix() string {
	stackArn, _ := arn.Parse(os.Getenv(StackArnEnvVar))
	return fmt.Sprintf("/jwt-issuer/%s", stackArn.Resource)
}
