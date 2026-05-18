# platform-tool

`platform-tool` is a CLI tool for managing Terraform variables using AWS Secrets Manager (ASM) and/or Terraform Cloud.

## Requirements

The repository publishes [pre-built binaries for all platforms](https://github.com/cesarbqz/platform-tool/releases) — no additional dependencies required.

## Installation

The repository includes an `install.sh` at the root that detects the target platform and installs the correct binary.

There are three ways to use `platform-tool` depending on your setup:

---

### Option A — Use directly (no fork needed)

If you just want to use the tool without customizing it, install directly from this repo.

**On your local machine:**

```bash
git clone https://github.com/cesarbqz/platform-tool
cd platform-tool
bash install.sh
```

**In a CI/CD pipeline:**

```bash
curl -sfL \
  https://raw.githubusercontent.com/cesarbqz/platform-tool/main/install.sh \
  -o install.sh

bash install.sh
```

No token needed — the repo is public and the binaries are available in the releases.

---

### Option B — Fork and keep it private

If you want to customize the tool (change defaults, add commands, etc.) and keep it in a private repo:

**1. Fork the repository on GitHub and set it to private.**

**2. Update `./internal/config/config.json` with your repo details:**

```json
{
    "version": "v0.0.0",
    "github_repo_owner": "your-org",
    "github_repo_name": "platform-tool",
    "hcl_assumerole_varname": "workspace_iam_roles",
    "default_aws_region": "us-east-1",
    "default_aws_profile": "default",
    "tf_cloud_org": "your-tfc-org"
}
```

**3. Build and release your own binaries** using the same GitHub Actions release workflow.

**4. Install from your private repo** using a GitHub token with `repo` scope:

```bash
# On your local machine
GH_TOKEN=your_github_token bash install.sh
```

```bash
# In a CI/CD pipeline
curl -s -H "Authorization: token $GH_TOKEN" \
  -H "Accept: application/vnd.github.v3.raw" \
  https://api.github.com/repos/your-org/platform-tool/contents/install.sh?ref=main \
  -o install.sh

GH_TOKEN=$GH_TOKEN bash install.sh
```

> **Why a token?** GitHub requires authentication to access private repo releases. The `GH_TOKEN` is used only to download the binary — it is never stored or sent anywhere else by the tool itself.

---

### Option C — Customize and use locally (no releases)

If you only need the tool locally and don't want to set up releases:

```bash
git clone https://github.com/cesarbqz/platform-tool
cd platform-tool

# Edit ./internal/config/config.json with your values
# then build the binary directly
go build -o platform-tool ./cmd/platform-tool/.

# Move to a directory in your PATH
mv platform-tool ~/.local/bin/
```

---

## Updating

Updates are explicit and opt-in. To update the binary to the latest release:

```bash
platform-tool update
```

> Note: `platform-tool update` only works if you installed from a repo with GitHub releases configured (Options A or B). For Option C, rebuild manually with `go build`.

---

## Commands

### Core

#### `version`

Prints the current version of the binary.

```bash
platform-tool version
```

#### `update`

Updates the binary to the latest release from the configured GitHub repo.

```bash
platform-tool update
```

---

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
  [--aws-region <region>] \
  [--json]
```

The secret is stored at: `/terraform/<region>/<repo-name>/<workspace>/config`

**Flags:**

| Flag               | Required | Description                                                                         |
| ------------------ | -------- | ----------------------------------------------------------------------------------- |
| `--file`           | Yes      | Path to local `.tfvars` file                                                        |
| `--repo-name`      | Yes      | IaC repository name                                                                 |
| `--workspace`      | Yes      | Workspace name                                                                      |
| `--assumerole-arn` | No       | ARN of the role to assume                                                           |
| `--profile`        | No       | AWS profile (default: `default`)                                                    |
| `--aws-region`     | No       | AWS region (default: `us-east-1`)                                                   |
| `--json`           | No       | Store variables as JSON — enables ephemeral resource consumption in Terraform 1.10+ |

**Default mode** — stores the `.tfvars` file as-is. Use `retrieve-asm --create-file` to restore it before running Terraform.

**JSON mode (`--json`)** — parses the `.tfvars` file and stores variables as a JSON object. This enables consuming the secret directly via Terraform ephemeral resources (Terraform 1.10+) without generating a local file and without persisting values in the state file:

```hcl
ephemeral "aws_secretsmanager_secret_version" "tfvars" {
  secret_id = "/terraform/us-east-1/my-repo/dev/config"
}

locals {
  vars = jsondecode(ephemeral.aws_secretsmanager_secret_version.tfvars.secret_string)
}
```

**Assume role support** — if `--assumerole-arn` is not provided, the tool attempts to extract it from the `.tfvars` file. It looks for a map variable named `workspace_iam_roles` (overridable via config), using the workspace name as the map key:

```hcl
workspace_iam_roles = {
  dev  = "arn:aws:iam::123456789012:role/TerraformRole"
  prod = "arn:aws:iam::987654321098:role/TerraformRole"
}
```

---

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

With `--create-file`, the output is written to `<workspace>.auto.tfvars` — picked up automatically by Terraform. Without it, the content is printed to stdout.

---

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

---

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

All fields are optional — the binary uses sensible defaults if not provided. Fields only required for specific features:

| Field                    | Default               | Required for                          |
| ------------------------ | --------------------- | ------------------------------------- |
| `version`                | `""`                  | Auto-updater (`platform-tool update`) |
| `github_repo_owner`      | `""`                  | Auto-updater                          |
| `github_repo_name`       | `""`                  | Auto-updater                          |
| `hcl_assumerole_varname` | `workspace_iam_roles` | Assume role auto-detection            |
| `default_aws_region`     | `us-east-1`           | ASM commands                          |
| `default_aws_profile`    | `default`             | ASM commands                          |
| `tf_cloud_org`           | `""`                  | Terraform Cloud commands              |
