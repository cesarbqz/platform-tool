package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func UpsertSecretInSecretsManager(assumeRoleArn, profile, region, secretName, secretValue string) error {
	client := secretsmanager.NewFromConfig(LoadAWSConfig(profile, region, assumeRoleArn))

	_, err := client.DescribeSecret(context.TODO(), &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			_, err = client.CreateSecret(context.TODO(), &secretsmanager.CreateSecretInput{
				Name:         aws.String(secretName),
				SecretString: aws.String(secretValue),
			})
			if err != nil {
				return fmt.Errorf("failed to create secret: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to describe secret: %w", err)
	}

	_, err = client.UpdateSecret(context.TODO(), &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(secretValue),
	})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

func DownloadFromSecretsManager(assumeRoleArn, profile, region, secretName string) (string, error) {
	client := secretsmanager.NewFromConfig(LoadAWSConfig(profile, region, assumeRoleArn))

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("Error obtaining secret %s: %w", secretName, err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	return "", fmt.Errorf("Secret %s does not contain a string", secretName)
}
