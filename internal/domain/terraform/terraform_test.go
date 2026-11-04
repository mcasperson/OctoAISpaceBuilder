package terraform

import (
	"strings"
	"testing"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTerraformRC(t *testing.T) {
	// Get current working directory to use in verification
	installDir, err := environment.GetInstallationDirectory()

	if err != nil {
		t.Fatal("Failed to get installation directory:", err)
	}

	providerPath := environment.GetTerraformProvidersPath()

	localPath := files.GetAbsoluteOrRelativePath(providerPath, installDir)

	// Call the function under test
	content, err := GenerateTerraformRC()

	// Verify no error occurred
	require.NoError(t, err, "GenerateTerraformRC should not return an error")

	// Verify the content
	assert.Contains(t, content, "provider_installation {", "Should contain provider_installation section")
	assert.Contains(t, content, "filesystem_mirror {", "Should contain filesystem_mirror section")
	assert.Contains(t, content, "path    = \""+localPath+"\"", "Should contain correct provider path")
	assert.Contains(t, content, "include = [\"*/*/*\"]", "Should include all providers")
	assert.Contains(t, content, "direct {exclude = [\"*/*/*\"]}", "Should contain direct provider section")

	// Verify overall structure
	lines := strings.Split(strings.TrimSpace(content), "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Should contain at least 5 lines")
}

func TestGenerateOverrides(t *testing.T) {
	// Call the function under test
	result := GenerateOverrides()
	providerVersion := environment.GetTerraformProviderVersion()

	// Verify the content
	assert.Contains(t, result, "required_providers {", "Should contain required_providers section")
	assert.Contains(t, result, "octopusdeploy = { source = \"OctopusDeploy/octopusdeploy\", version = \""+providerVersion+"\" }",
		"Should contain correct octopusdeploy provider definition")
	assert.Contains(t, result, "required_version = \">= 1.6.0\"", "Should contain minimum Terraform version")

	// Verify the structure
	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Should contain at least 5 lines")

	// Verify the provider version matches the constant
	assert.Contains(t, result, "version = \""+providerVersion+"\"",
		"Provider version should match environment.GetTerraformProviderVersion")
}

func TestGenerateStateFile(t *testing.T) {
	// Call the function under test
	result := GenerateStateFile()

	// Verify the content
	assert.Contains(t, result, "terraform {", "Should contain terraform block")
	assert.Contains(t, result, "backend \"local\" {", "Should contain local backend block")
	assert.Contains(t, result, "path = \"./.local-state\"", "Should define local state path correctly")

	// Verify structure
	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Should contain at least 5 lines")

	// Verify proper formatting
	assert.True(t, strings.Contains(result, "{") && strings.Contains(result, "}"),
		"Should contain properly formed blocks with braces")
}
