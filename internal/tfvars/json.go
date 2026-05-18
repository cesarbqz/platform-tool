package tfvars

import (
	"encoding/json"
	"fmt"
)

// ConvertTfvarsToJSON parses a .tfvars file and returns its content as a JSON string.
// This enables storing variables in AWS Secrets Manager in a format compatible
// with Terraform ephemeral resources (Terraform 1.10+), allowing secrets to be
// consumed at runtime without being persisted in the state file.
func ConvertTfvarsToJSON(filePath string) (string, error) {
	variables, err := ParseTfvars(filePath)
	if err != nil {
		return "", fmt.Errorf("error parsing tfvars file: %w", err)
	}

	jsonBytes, err := json.Marshal(variables)
	if err != nil {
		return "", fmt.Errorf("error serializing variables to JSON: %w", err)
	}

	return string(jsonBytes), nil
}
