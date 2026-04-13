package tfcloud

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"platform-tool/internal/tfvars"

	tfe "github.com/hashicorp/go-tfe"
)

type TFCClient struct {
	client *tfe.Client
}

func NewClient(token string) (*TFCClient, error) {
	config := &tfe.Config{
		Token: token,
	}
	client, err := tfe.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFC client: %w", err)
	}
	return &TFCClient{client: client}, nil
}

type WorkspaceInfo struct {
	ProjectID string
	Project   string
	Workspace string
	HasVars   bool
}

func (c *TFCClient) ListWorkspacesWithVars(org string) ([]WorkspaceInfo, error) {
	ctx := context.Background()

	var allWorkspaces []*tfe.Workspace
	page := 1

	for {
		workspaces, err := c.client.Workspaces.List(ctx, org, &tfe.WorkspaceListOptions{
			ListOptions: tfe.ListOptions{
				PageSize:   100,
				PageNumber: page,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list workspaces (page %d): %w", page, err)
		}

		allWorkspaces = append(allWorkspaces, workspaces.Items...)

		if len(workspaces.Items) < 100 {
			break // reached last page
		}
		page++
	}

	projectCache := make(map[string]string)
	var results []WorkspaceInfo

	for _, ws := range allWorkspaces {
		vars, err := c.client.Variables.List(ctx, ws.ID, &tfe.VariableListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error fetching vars for workspace %s: %w", ws.Name, err)
		}

		hasVars := len(vars.Items) > 0

		projectID := ""
		projectName := "unknown"
		if ws.Project != nil && ws.Project.ID != "" {
			projectID = ws.Project.ID
			if cached, ok := projectCache[projectID]; ok {
				projectName = cached
			} else {
				project, err := c.client.Projects.Read(ctx, projectID)
				if err != nil {
					projectName = "error-loading-project"
				} else {
					projectName = project.Name
					projectCache[projectID] = projectName
				}
			}
		}

		results = append(results, WorkspaceInfo{
			ProjectID: projectID,
			Project:   projectName,
			Workspace: ws.Name,
			HasVars:   hasVars,
		})
	}

	return results, nil
}

func UploadTfvars(tfcloudOrg, filePath, workspaceName, tfCloudToken string) error {
	client, err := tfe.NewClient(&tfe.Config{
		Token: tfCloudToken,
	})
	if err != nil {
		return fmt.Errorf("error creating Terraform Cloud client: %w", err)
	}

	workspace, err := client.Workspaces.Read(context.Background(), tfcloudOrg, workspaceName)
	if err != nil {
		return fmt.Errorf("error getting workspace '%s': %w", workspaceName, err)
	}

	variables, err := tfvars.ParseTfvars(filePath)
	if err != nil {
		return fmt.Errorf("error parsing tfvars file: %w", err)
	}

	existingVars, err := client.Variables.List(context.Background(), workspace.ID, nil)
	if err != nil {
		return fmt.Errorf("error getting existing variables: %w", err)
	}

	existingVarsMap := make(map[string]*tfe.Variable)
	for _, variable := range existingVars.Items {
		existingVarsMap[variable.Key] = variable
	}

	for key, value := range variables {
		formattedValue, err := formatToTfvars(value)
		category := tfe.CategoryTerraform
		if err != nil {
			return fmt.Errorf("error formatting variable value'%s': %w", key, err)
		}

		if existingVar, exists := existingVarsMap[key]; exists {
			if existingVar.Value != formattedValue {
				fmt.Printf("Updating %s variable\n", key)
				_, err := client.Variables.Update(context.Background(), workspace.ID, existingVar.ID, tfe.VariableUpdateOptions{
					Value:     tfe.String(formattedValue),
					Category:  &category,
					HCL:       tfe.Bool(true),
					Sensitive: tfe.Bool(false),
				})
				if err != nil {
					return fmt.Errorf("error updating variable '%s': %w", key, err)
				}
			}
		} else {
			fmt.Printf("Creating %s variable\n", key)
			_, err := client.Variables.Create(context.Background(), workspace.ID, tfe.VariableCreateOptions{
				Key:       tfe.String(key),
				Value:     tfe.String(formattedValue),
				Category:  &category,
				HCL:       tfe.Bool(true),
				Sensitive: tfe.Bool(false),
			})
			if err != nil {
				return fmt.Errorf("error creating variable '%s': %w", key, err)
			}
		}
	}

	for key, existingVar := range existingVarsMap {
		if _, exists := variables[key]; !exists {
			fmt.Printf("Deleting %s variable\n", key)
			err := client.Variables.Delete(context.Background(), workspace.ID, existingVar.ID)
			if err != nil {
				return fmt.Errorf("error deleting variable '%s': %w", key, err)
			}
		}
	}

	return nil
}

func formatToTfvars(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v), nil
	case int, float64:
		return fmt.Sprintf(`%v`, v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case []interface{}:
		var elements []string
		for _, elem := range v {
			formattedElem, err := formatToTfvars(elem)
			if err != nil {
				return "", err
			}
			elements = append(elements, formattedElem)
		}
		return fmt.Sprintf("[%s]", strings.Join(elements, ", ")), nil
	case map[string]interface{}:
		var buffer bytes.Buffer
		buffer.WriteString("{")
		for key, val := range v {
			formattedVal, err := formatToTfvars(val)
			if err != nil {
				return "", err
			}
			buffer.WriteString(fmt.Sprintf(`"%s" = %s, `, key, formattedVal))
		}
		result := buffer.String()
		return strings.TrimSuffix(result, ", ") + "}", nil
	default:
		return "", fmt.Errorf("data type not supported: %T", v)
	}
}

func DownloadTfvars(tfcloudOrg, workspaceName, tfCloudToken string, createFile bool) error {
	client, err := tfe.NewClient(&tfe.Config{
		Token: tfCloudToken,
	})
	if err != nil {
		return fmt.Errorf("error creating Terraform Cloud client: %w", err)
	}

	workspace, err := client.Workspaces.Read(context.Background(), tfcloudOrg, workspaceName)
	if err != nil {
		return fmt.Errorf("error getting workspace '%s': %w", workspaceName, err)
	}

	variables, err := client.Variables.List(context.Background(), workspace.ID, nil)
	if err != nil {
		return fmt.Errorf("error getting workspace variables: %w", err)
	}

	if createFile {
		fileName := workspaceName + ".auto.tfvars"
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("error creating tfvars file %s: %w", fileName, err)
		}
		defer file.Close()
		for _, variable := range variables.Items {
			if variable.Category == tfe.CategoryTerraform {
				var line string
				if variable.HCL {
					line = fmt.Sprintf("%s = %s\n", variable.Key, variable.Value)
				} else {
					line = fmt.Sprintf("%s = \"%s\"\n", variable.Key, escapeQuotes(variable.Value))
				}

				_, err := file.WriteString(line)
				if err != nil {
					return fmt.Errorf("error writing tfvars file: %w", err)
				}
			}
		}
	} else {
		for _, variable := range variables.Items {
			if variable.Category == tfe.CategoryTerraform {
				if variable.HCL {
					fmt.Printf("%s = %s\n", variable.Key, variable.Value)
				} else {
					fmt.Printf("%s = \"%s\"\n", variable.Key, escapeQuotes(variable.Value))
				}
			}
		}
	}

	return nil
}

func escapeQuotes(value string) string {
	return strings.ReplaceAll(value, "\"", "\\\"")
}
