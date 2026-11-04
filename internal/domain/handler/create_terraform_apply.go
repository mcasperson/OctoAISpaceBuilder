package handler

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/execute"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/files"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/logging"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/model"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/sha"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/terraform"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/infrastructure"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func CreateTerraformApply(server string, token string, apiKey string, terraformApply model.TerraformApply) (*model.TerraformApply, error, string) {

	terraformApply.Server = sha.GetSha256Hash(server)

	tempDir, err := files.CreateTempDir()

	if err != nil {
		return nil, err, ""
	}

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			zap.L().Error("Failed to remove temporary directory", zap.Error(err))
		}
	}()

	planFile := filepath.Join(tempDir, "tfplan")
	lockFileName := filepath.Join(tempDir, ".terraformApply.lock.hcl")
	configurationFileName := filepath.Join(tempDir, "terraformApply.tf")

	planContents, lockFile, configuration, err := infrastructure.ReadPlanAzureStorageBlob(terraformApply.PlanId)

	if err != nil {
		return nil, err, ""
	}

	if err := os.WriteFile(planFile, planContents, 0644); err != nil {
		return nil, err, ""
	}

	if err := os.WriteFile(lockFileName, lockFile, 0644); err != nil {
		return nil, err, ""
	}

	if err := os.WriteFile(configurationFileName, configuration, 0644); err != nil {
		return nil, err, ""
	}

	if err := terraform.WriteOverrides(tempDir); err != nil {
		return nil, err, ""
	}

	var cliConfigFile = ""
	if !environment.GetDisableTerraformCliConfig() {
		cliConfigFile, err = terraform.CreateTerraformRcFile()

		if err != nil {
			return nil, err, ""
		}
	}

	tofu, err := environment.GetTofuExecutable()

	if err != nil {
		return nil, err, ""
	}

	stdout, stderr, _, err := execute.Execute(
		tofu,
		[]string{
			"-chdir=" + tempDir,
			"init",
			"-input=false",
			"-no-color"},
		map[string]string{
			"TF_INPUT":           "0",
			"TF_LOG":             "INFO",
			"TF_CLI_CONFIG_FILE": cliConfigFile,
		})

	if err != nil {
		return nil, err, stdout + "\n" + stderr
	}

	stdout, stderr, _, err = execute.Execute(
		tofu,
		[]string{
			"-chdir=" + tempDir,
			"apply",
			"-auto-approve",
			"-input=false",
			"-no-color",
			planFile},
		map[string]string{
			"OCTOPUS_ACCESS_TOKEN":        token,
			"OCTOPUS_URL":                 server,
			"OCTOPUS_API_KEY":             apiKey,
			"TF_INPUT":                    "0",
			"TF_LOG":                      "INFO",
			"TF_CLI_CONFIG_FILE":          cliConfigFile,
			"TF_VAR_octopus_apikey":       "",
			"TF_VAR_octopus_server":       "",
			"REDIRECTION_SERVICE_API_KEY": os.Getenv("REDIRECTION_SERVICE_API_KEY"),
			"REDIRECTION_HOST":            os.Getenv("REDIRECTION_HOST"),
			"REDIRECTION_SERVICE_ENABLED": os.Getenv("REDIRECTION_SERVICE_ENABLED"),
			"TF_REATTACH_PROVIDERS":       os.Getenv("TF_REATTACH_PROVIDERS"),
		})

	logging.LogEnhanced(stdout, server)
	logging.LogEnhanced(stderr, server)

	if err != nil {
		return nil, errors.New("Failed to generate plan: " + stderr + " " + stdout + " " + err.Error()), stdout + "\n" + stderr
	}

	if err := infrastructure.DeletePlanAzureStorageBlob(terraformApply.PlanId); err != nil {
		// We're not going to fail here, but we'll log the error.
		// Any old plans will be cleaned up by the terraform plan cleanup job.
		zap.L().Error("Failed to delete terraform plan", zap.Error(err))
	}

	applyText := stdout + "\n" + stderr

	response := model.TerraformApply{
		ID:        uuid.New().String(),
		PlanId:    terraformApply.PlanId,
		Server:    terraformApply.Server,
		ApplyText: &applyText,
	}

	return &response, nil, ""
}
