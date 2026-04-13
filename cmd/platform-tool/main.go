package main

import (
	"platform-tool/internal/cli"
	"platform-tool/internal/config"
	"platform-tool/internal/log"
	"platform-tool/internal/updater"

	"github.com/spf13/cobra"
)

func main() {
	var (
		filePath      string
		repoName      string
		workspace     string
		profile       string
		region        string
		assumeRoleArn string
		createFile    bool
	)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	update, err := updater.Update(cfg.CurrentVersion)
	if err != nil {
		log.FatalfColored("Could not update the binary: %s", update)
	}

	rootCmd := &cobra.Command{
		Use:   "platform-tool",
		Short: "Platform CLI tool",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("platform-tool version %s", cfg.CurrentVersion)
		},
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "core commands:",
	})

	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "Show version information",
		GroupID: "core",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("platform-tool version %s", cfg.CurrentVersion)
		},
	}

	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Update CLI",
		GroupID: "core",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("platform-tool version %s", cfg.CurrentVersion)
		},
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    "tfvars",
		Title: "tfvars commands:",
	})

	uploadASMCmd := &cobra.Command{
		Use:     "upload-asm",
		Short:   "Upload tfvars file to AWS Secrets Manager",
		GroupID: "tfvars",
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" || repoName == "" || workspace == "" {
				log.Fatal("Required flags: --file, --repo-name, --workspace")
			}
			cli.UploadSecret(cfg, filePath, repoName, workspace, assumeRoleArn, profile, region)
		},
	}

	retrieveASMCmd := &cobra.Command{
		Use:     "retrieve-asm",
		Short:   "Retrieve tfvars file from AWS Secrets Manager",
		GroupID: "tfvars",
		Run: func(cmd *cobra.Command, args []string) {
			if repoName == "" || workspace == "" {
				log.Fatal("Required flags: --repo-name, --workspace")
			}
			cli.RetrieveSecret(cfg, repoName, workspace, assumeRoleArn, profile, region, createFile)
		},
	}

	uploadTFCCmd := &cobra.Command{
		Use:     "upload-tfcloud",
		Short:   "Upload tfvars file to Terraform Cloud",
		GroupID: "tfvars",
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" || workspace == "" {
				log.Fatal("Required flags: --file, --workspace")
			}
			cli.UploadToTFCloud(cfg.TFCloudOrg, filePath, workspace)
		},
	}

	retrieveTFCCmd := &cobra.Command{
		Use:     "retrieve-tfcloud",
		Short:   "Retrieve tfvars file from Terraform Cloud",
		GroupID: "tfvars",
		Run: func(cmd *cobra.Command, args []string) {
			if workspace == "" {
				log.Fatal("Required flag: --workspace")
			}
			cli.RetrieveFromTFCloud(cfg.TFCloudOrg, workspace, createFile)
		},
	}

	uploadASMCmd.Flags().StringVar(&filePath, "file", "", "Path to local tfvars file")
	uploadASMCmd.Flags().StringVar(&repoName, "repo-name", "", "IaC repository name")
	uploadASMCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace name")
	uploadASMCmd.Flags().StringVar(&assumeRoleArn, "assumerole-arn", "", "ARN of the role to assume")
	uploadASMCmd.Flags().StringVar(&profile, "profile", "", "AWS profile")
	uploadASMCmd.Flags().StringVar(&region, "aws-region", "", "AWS region")

	retrieveASMCmd.Flags().AddFlagSet(uploadASMCmd.Flags())
	retrieveASMCmd.Flags().BoolVar(&createFile, "create-file", false, "Write file instead of printing to stdout")
	uploadTFCCmd.Flags().StringVar(&filePath, "file", "", "Path to local tfvars file")
	uploadTFCCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace name")
	retrieveTFCCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace name")
	retrieveTFCCmd.Flags().BoolVar(&createFile, "create-file", false, "Write file instead of printing to stdout")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(uploadASMCmd)
	rootCmd.AddCommand(retrieveASMCmd)
	rootCmd.AddCommand(uploadTFCCmd)
	rootCmd.AddCommand(retrieveTFCCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
