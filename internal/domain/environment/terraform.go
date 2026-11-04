package environment

import (
	"os"
	"strings"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/files"
)

/*
This version must be updated:
* In the build.yaml file
* In the Dockerfile
* In the functions/provider/registry.opentofu.org/octopusdeploy/octopusdeploy directory
*/
const terraformProviderVersion = "1.18.2"

func GetDisableTerraformCliConfig() bool {
	// Check if the environment variable is set to "true"
	if strings.ToLower(os.Getenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG")) == "true" {
		return true
	}

	// Default to false if the environment variable is not set or not "true"
	return false
}

func GetTerraformProviderVersion() string {
	if os.Getenv("SPACEBUILDER_TERRAFORM_PROVIDER_VERSION") != "" {
		return os.Getenv("SPACEBUILDER_TERRAFORM_PROVIDER_VERSION")
	}

	return terraformProviderVersion
}

func GetTerraformProvidersPath() string {
	if os.Getenv("SPACEBUILDER_TERRAFORM_PROVIDERS") != "" {
		return os.Getenv("SPACEBUILDER_TERRAFORM_PROVIDERS")
	}

	return "provider"
}

func GetCombinedTerraformProvidersPath() (string, error) {
	installDir, err := GetInstallationDirectory()

	if err != nil {
		return "", err
	}

	providerPath := GetTerraformProvidersPath()

	return files.GetAbsoluteOrRelativePath(providerPath, installDir), nil
}
