package kms

import (
	"encoding/base64"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

var client = kms.New(session.Must(session.NewSession()))

func Decrypt(encrypted string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	response, err := client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: decodedBytes,
		EncryptionContext: aws.StringMap(map[string]string{
			"LambdaFunctionName": os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		}),
	})
	if err != nil {
		return "", err
	}
	return string(response.Plaintext[:]), nil
}
