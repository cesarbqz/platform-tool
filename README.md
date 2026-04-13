# platform-tool

`platform-tool` is a CLI tool for managing Terraform variables using AWS Secrets Manager (ASM) and/or Terraform Cloud.

## Requirements

The repository publishes [pre-built binaries for all platforms](https://github.com/cesarbqz/platform-tool/releases) — no additional dependencies required.

On every run, the CLI auto-updates to the latest release.

## Installation

The repository includes an `install.sh` at the root that detects the target platform and installs the correct binary.

### In pipelines

```bash
curl -s -H "Authorization: token $GH_TOKEN" \
  -H "Accept: application/vnd.github.v3.raw" \
  https://api.github.com/repos/cesarbqz/platform-tool/contents/install.sh?ref=main \
  -o install.sh

bash install.sh
```

### On desktops

Clone the repository and run `install.sh`.

## Commands

### AWS Secrets Manager

#### `upload-asm`

Uploads a `.tfvars` file to AWS Secrets Manager.

```bash
platform-tool upload-asm \
  --file <path-to-file.tfvars> \
  --repo-name <iac-repo-name> \
  --workspace <workspace-name> \
  [--assumerole-arn <arn>] \
  [--profile <aws-profile>] \
  [--aws-region <region>]
```

If `--assumerole-arn` is not provided:

1. The tool attempts to extract it from the `.tfvars` file — it looks for a map variable whose name matches `hcl_assumerole_varname` (from config), using the workspace name as the map key.
2. If the variable is not found in the file, it proceeds without assume role and uses the `--profile` directly.

Example `.tfvars` with ARNs per workspace (with `hcl_assumerole_varname = "workspace_iam_roles"` and `--workspace dev`):

```hcl
workspace_iam_roles = {
  dev  = "arn:aws:iam::123456789012:role/TerraformRole"
  prod = "arn:aws:iam::987654321098:role/TerraformRole"
}
```

The secret is stored at: `/terraform/<region>/<repo-name>/<workspace>/config`

#### `retrieve-asm`

Retrieves variables from AWS Secrets Manager.

```bash
platform-tool retrieve-asm \
  --repo-name <iac-repo-name> \
  --workspace <workspace-name> \
  [--assumerole-arn <arn>] \
  [--profile <aws-profile>] \
  [--aws-region <region>] \
  [--create-file]
```

With `--create-file`, the output is written to `<workspace>.auto.tfvars`. Without it, the content is printed to stdout.

### Terraform Cloud

#### `upload-tfcloud`

Uploads variables from a `.tfvars` file to a Terraform Cloud workspace.

```bash
platform-tool upload-tfcloud \
  --file <path-to-file.tfvars> \
  --workspace <workspace-name>
```

Requires `terraform login` to have been run beforehand (uses the token from `~/.terraform.d/credentials.tfrc.json`).

#### `retrieve-tfcloud`

Retrieves variables from a Terraform Cloud workspace.

```bash
platform-tool retrieve-tfcloud \
  --workspace <workspace-name> \
  [--create-file]
```

With `--create-file`, the output is written to `<workspace>.auto.tfvars`. Without it, the content is printed to stdout.

## Development

- **Go 1.23+**
- **JSON config**: a `config.json` file under `./internal/config/` with the CLI default values.

### Configuration

Create `./internal/config/config.json`:

```json
{
    "version": "v0.0.0",
    "github_repo_owner": "cesarbqz",
    "github_repo_name": "platform-tool",
    "hcl_assumerole_varname": "workspace_iam_roles",
    "default_aws_region": "us-east-1",
    "default_aws_profile": "default",
    "tf_cloud_org": "your-tfc-org"
}
```

| Field                    | Description                                                                |
| ------------------------ | -------------------------------------------------------------------------- |
| `version`                | Current binary version                                                     |
| `github_repo_owner`      | GitHub repository owner (used by the auto-updater)                         |
| `github_repo_name`       | GitHub repository name (used by the auto-updater)                          |
| `hcl_assumerole_varname` | Name of the HCL map variable that holds the assume role ARNs per workspace |
| `default_aws_region`     | Default AWS region (can be overridden with `--aws-region`)                 |
| `default_aws_profile`    | Default AWS profile (can be overridden with `--profile`)                   |
| `tf_cloud_org`           | Terraform Cloud organization                                               |
