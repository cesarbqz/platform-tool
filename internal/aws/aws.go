package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func LoadAWSConfig(profile, region, assumeRoleArn string) aws.Config {
	var cfg aws.Config
	var err error

	if profile != "noprofile" {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(profile),
			config.WithRegion(region),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
	}

	if err != nil {
		log.Fatalf("Error loading AWS config: %v", err)
	}

	if assumeRoleArn != "" {
		stsClient := sts.NewFromConfig(cfg)
		assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, assumeRoleArn)
		cfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
	}

	return cfg
}
