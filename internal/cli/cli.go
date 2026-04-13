package cli

import (
	"fmt"
	"os"

	"platform-tool/internal/aws"
	"platform-tool/internal/config"
	"platform-tool/internal/log"
	"platform-tool/internal/tfcloud"
	"platform-tool/internal/tfvars"
)

func GetSecretName(region, repoName, workspace string) string {
	secretName := fmt.Sprintf("/terraform/%s/%s/%s/config", region, repoName, workspace)
	return secretName
}

func GetAssumeRoleArnFromFile(filePath, variableName, mapKey string) string {
	val, err := tfvars.LoadValueFromMap(filePath, variableName, mapKey)
	if err != nil {
		return ""
	}
	return val
}

func UploadSecret(cfg *config.Config, filePath, repoName, workspace, assumeRoleArn, profile, region string) {
	region, profile = getDefaultsIfNotExists(cfg, region, profile)
	log.Printf("Uploading %s to AWS Secrets Manager...", filePath)

	if assumeRoleArn == "" {
		assumeRoleArn = GetAssumeRoleArnFromFile(filePath, cfg.DefaultAssumeRoleHCLVarName, workspace)
		if assumeRoleArn != "" {
			log.Printf("Using assume role ARN from tfvars file: %s", assumeRoleArn)
		}
	}
	secretValue, err := tfvars.LoadTfvarsFile(filePath)
	if err != nil {
		log.FatalfColored("%s", err)
	}

	secretName := GetSecretName(region, repoName, workspace)

	result := aws.UpsertSecretInSecretsManager(assumeRoleArn, profile, region, secretName, secretValue)

	if result != nil {
		log.FatalfColored("Error creating Secret %s: %s", secretName, result)
	}

	log.PrintfColored("Secret created: %s", secretName)
}

func RetrieveSecret(cfg *config.Config, repoName, workspace, assumeRoleArn, profile, region string, createFile bool) {
	region, profile = getDefaultsIfNotExists(cfg, region, profile)
	log.Printf("Retrieving %s.auto.tfvars from AWS Secrets Manager...", workspace)

	secretName := GetSecretName(region, repoName, workspace)

	result, err := aws.DownloadFromSecretsManager(assumeRoleArn, profile, region, secretName)
	if err != nil {
		log.FatalfColored("Error getting Secret %s: %s", secretName, err)
	}

	// Maybe needs refactor. In what layer the file should be created or not
	if createFile {
		fileName := workspace + ".auto.tfvars"
		file, err := os.Create(fileName)
		if err != nil {
			log.FatalfColored("Error creating %s file: %s", fileName, err)
		}

		_, err = file.WriteString(result)
		if err != nil {
			log.FatalfColored("Error writing file %s: %s", fileName, err)
		}

		log.PrintfColored("File %s created", file.Name())

	} else {
		log.Print(result)
	}
}

func UploadToTFCloud(TFCloudOrg, filePath, workspace string) {
	tfCloudToken, err := tfcloud.GetTFCloudToken()
	if err != nil {
		log.Fatalf("Error getting Terraform Cloud token: %v", err)
	}

	log.Printf("Uploading %s to Terraform Cloud. Workspace: %s ", filePath, workspace)

	errUpload := tfcloud.UploadTfvars(TFCloudOrg, filePath, workspace, tfCloudToken)
	if errUpload != nil {
		log.FatalfColored("Failed to upload variables to Terraform Cloud: %v", err)
	}

	log.PrintlnColored("Variables uploaded successfully to Terraform Cloud")
}

func RetrieveFromTFCloud(TFCloudOrg, workspace string, createFile bool) {
	tfCloudToken, err := tfcloud.GetTFCloudToken()
	if err != nil {
		log.Fatalf("Error getting Terraform Cloud token: %v", err)
	}

	log.Printf("Retrieving %s.auto.tfvars from Terraform Cloud. Workspace: %s", workspace, workspace)

	// Maybe needs refactor. In what layer the file should be created or not
	errRetrieve := tfcloud.DownloadTfvars(TFCloudOrg, workspace, tfCloudToken, createFile)
	if errRetrieve != nil {
		log.FatalfColored("Failed to download variables to Terraform Cloud: %v", err)
	}

	log.PrintlnColored("\nVariables downloaded successfully from Terraform Cloud")
}

func getDefaultsIfNotExists(cfg *config.Config, region, profile string) (string, string) {
	if region == "" {
		log.Printf("No region provided, using default: %s", cfg.DefaultAWSRegion)
		region = cfg.DefaultAWSRegion
	}

	if profile == "" {
		log.Printf("No profile provided, using default: %s", cfg.DefaultAWSProfile)
		profile = cfg.DefaultAWSProfile
	}
	return region, profile
}
