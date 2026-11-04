package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"go.uber.org/zap"
)

func WriteOverrides(path string) error {
	if err := WriteBackendOverride(path); err != nil {
		return err
	}

	if err := WriteProviderOverrides(path); err != nil {
		return err
	}

	return nil
}

// WriteBackendOverride forces the use of local state. We never want to send state to the cloud.
func WriteBackendOverride(path string) error {
	backendOverride := GenerateStateFile()

	filePath := filepath.Join(path, "backend_override.tf")

	return os.WriteFile(filePath, []byte(backendOverride), 0644)
}

func GenerateStateFile() string {
	return `terraform {
	  backend "local" {
		path = "./.local-state"
	  }
	}`
}

// WriteProviderOverrides defines the provider block for the Octopus Deploy provider.
// It sets the version to the bundled provider and removes the optional providers that are not needed.
func WriteProviderOverrides(path string) error {
	providerOverrides := GenerateOverrides()

	filePath := filepath.Join(path, "provider_override.tf")

	return os.WriteFile(filePath, []byte(providerOverrides), 0644)
}

func GenerateOverrides() string {
	return `terraform {
	  required_providers {
		octopusdeploy = { source = "OctopusDeploy/octopusdeploy", version = "` + environment.GetTerraformProviderVersion() + `" }
	  }
	  required_version = ">= 1.6.0"
	}`
}

// createTerraformRcFile creates a CLI configuration file for Terraform.
// This file must be referecned using the TF_CLI_CONFIG_FILE environment variable.
// https://opentofu.org/docs/cli/config/config-file/
// The providers directory structure needs to be like:
// provider/registry.terraform.io/octopusdeploy/octopusdeploy/0.41.0/linux_amd64/terraform-provider-octopusdeploy_v0.41.0
func CreateTerraformRcFile() (string, error) {
	content, err := GenerateTerraformRC()

	if err != nil {
		return "", fmt.Errorf("failed to generate terraform rc file content: %w", err)
	}

	zap.L().Info("Generated terraform rc file: " + content)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user home directory: %w", err)
	}

	rcFilePath := filepath.Join(homeDir, "cliconfig.tfrc")

	// Check if file already exists
	if err := BackupRcFile(rcFilePath); err != nil {
		return "", err
	}

	// Write the new content
	if err := os.WriteFile(rcFilePath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write cliconfig.tfrc file: %w", err)
	}

	return rcFilePath, nil
}

// GenerateTerraformRC created a CLI config file prevents providers from being downloaded.
// See https://github.com/hashicorp/terraform/issues/33698
func GenerateTerraformRC() (string, error) {
	providerPath, err := environment.GetCombinedTerraformProvidersPath()

	if err != nil {
		return "", err
	}

	return `provider_installation {
  filesystem_mirror {
    path    = "` + providerPath + `"
    include = ["*/*/*"]
  }
  direct {exclude = ["*/*/*"]}
}`, nil

}

func BackupRcFile(rcFilePath string) error {
	if _, err := os.Stat(rcFilePath); err == nil {
		// Back up existing file
		backupPath := rcFilePath + ".backup." + time.Now().Format("20060102150405")
		if err := os.Rename(rcFilePath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing .terraformrc file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if .terraformrc file exists: %w", err)
	}

	return nil
}
