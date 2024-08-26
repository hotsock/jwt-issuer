package issuer

import (
	"context"
	"encoding/asn1"
	"encoding/base64"
	"log/slog"
	"math/big"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
)

type KMSAPI interface {
	GetPublicKey(context.Context, *kms.GetPublicKeyInput, ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(context.Context, *kms.SignInput, ...func(*kms.Options)) (*kms.SignOutput, error)
}

// SignJWTWithKMS signs a JWT using a private key that is known only to KMS
func SignJWTWithKMS(ctx context.Context, kmsClient KMSAPI, token *jwt.Token, kmsKeyArn string) (string, error) {
	defer LogWithTiming(ctx, slog.LevelDebug, "issuer.SignJWTWithKMS", "token", token, "kmsKeyArn", kmsKeyArn)()

	sstr, err := token.SigningString()
	if err != nil {
		return "", err
	}

	signInput := &kms.SignInput{
		KeyId:            lo.ToPtr(kmsKeyArn),
		Message:          []byte(sstr),
		MessageType:      kmstypes.MessageTypeRaw,
		SigningAlgorithm: kmstypes.SigningAlgorithmSpecEcdsaSha256,
	}

	signOutput, err := kmsClient.Sign(ctx, signInput)
	if err != nil {
		return "", err
	}

	// KMS returns a DER-encoded object as defined by ANS X9.62–2005
	// and RFC 3279 Section 2.2.3 (https://tools.ietf.org/html/rfc3279#section-2.2.3).
	//
	// We need to convert it to the JWT r || s format before applying
	// it as the signature.
	// https://stackoverflow.com/questions/66170120/aws-kms-signature-returns-invalid-signature-for-my-jwt
	// https://stackoverflow.com/questions/48423188/verifying-a-ecdsa-signature-with-a-provided-public-key
	var esig struct {
		R, S *big.Int
	}
	asn1.Unmarshal(signOutput.Signature, &esig)

	fullSignature := []byte{}
	fullSignature = append(fullSignature, esig.R.Bytes()...)
	fullSignature = append(fullSignature, esig.S.Bytes()...)

	sig := strings.TrimRight(base64.URLEncoding.EncodeToString(fullSignature), "=")

	return strings.Join([]string{sstr, sig}, "."), nil
}
