package issuer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParameterStoreKeyID(t *testing.T) {
	os.Setenv(StackArnEnvVar, "arn:aws:cloudformation:us-east-1:111111111111:stack/JWTIssuer/d0a511e0-531d-11ee-8080-0a1f08df5697")
	assert.Equal(t, "d0a511e0-531d-11ee-8080-0a1f08df5697", ParameterStoreKeyID())
}

func Test_PrivateKeyParameterName(t *testing.T) {
	os.Setenv(StackArnEnvVar, "arn:aws:cloudformation:us-east-1:111111111111:stack/JWTIssuer/d0a511e0-531d-11ee-8080-0a1f08df5697")
	name := PrivateKeyParameterName()
	assert.Equal(t, "/jwt-issuer/stack/JWTIssuer/d0a511e0-531d-11ee-8080-0a1f08df5697/private-key", name)
}

func Test_PublicKeyParameterName(t *testing.T) {
	os.Setenv(StackArnEnvVar, "arn:aws:cloudformation:us-east-1:111111111111:stack/JWTIssuer/d0a511e0-531d-11ee-8080-0a1f08df5697")
	name := PublicKeyParameterName()
	assert.Equal(t, "/jwt-issuer/stack/JWTIssuer/d0a511e0-531d-11ee-8080-0a1f08df5697/public-key", name)
}
