package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/customerrors"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/execute"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/files"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/logging"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/model"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/sha"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/terraform"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/validation"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/infrastructure"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func CreateTerraformPlan(server string, token string, apiKey string, terraformInput model.TerraformPlan) (*model.TerraformPlan, error, string) {

	logging.LogEnhanced(terraformInput.Configuration, server)

	if err := validation.ValidateTerraformPlanRequest(terraformInput); err != nil {
		return nil, err, ""
	}

	tempDir, err := files.CreateTempDir()

	if err != nil {
		return nil, err, ""
	}

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			zap.L().Error("Failed to remove temporary directory", zap.Error(err))
		}
	}()

	if err := os.WriteFile(filepath.Join(tempDir, "terraformInput.tf"), []byte(terraformInput.Configuration), 0644); err != nil {
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

	lockFile, err := initTofu(cliConfigFile, tempDir)

	if err != nil {
		return nil, err, ""
	}

	planFile, planBinary, err, logs := generatePlanLoop(cliConfigFile, tempDir, token, apiKey, server, terraformInput.SpaceId, 0)

	if err != nil {
		return nil, err, logs
	}

	response := model.TerraformPlan{
		ID:      uuid.New().String(),
		Created: time.Now(),
		Server:  sha.GetSha256Hash(server),
		SpaceId: terraformInput.SpaceId,
	}

	if err := infrastructure.CreatePlanAzureStorageBlob(response.ID, planBinary, lockFile, []byte(terraformInput.Configuration)); err != nil {
		return nil, err, logs
	}

	planJson, err, _ := generatePlanJson(tempDir, planFile)

	if err != nil {
		return nil, err, logs
	}

	logging.LogEnhanced(planJson, server)

	if err := checkPlan(planJson, server); err != nil {
		return nil, err, logs
	}

	planText, err := generatePlanText(tempDir, planFile)

	if err != nil {
		return nil, err, logs
	}

	logging.LogEnhanced(planText, server)

	response.PlanText = &planText

	return &response, nil, logs
}

func initTofu(cliConfigFile string, tempDir string) ([]byte, error) {
	zap.L().Info("Init tofu")

	tofu, err := environment.GetTofuExecutable()

	if err != nil {
		return nil, err
	}

	stdOut, stdErr, _, err := execute.Execute(
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
		zap.L().Error("Failed to init: "+stdErr+" "+stdOut, zap.Error(err))
		return nil, errors.New("Failed to init: " + stdErr + " " + stdOut + " " + err.Error())
	}

	lockFile, err := os.ReadFile(filepath.Join(tempDir, ".terraform.lock.hcl"))

	if err != nil {
		zap.L().Error("Failed to get lock file: "+stdErr, zap.Error(err))
		return nil, errors.New("Failed to get lock file: " + stdErr)
	}

	return lockFile, nil
}

func generatePlanLoop(cliConfigFile string, tempDir string, token string, apiKey string, aud string, spaceId string, retry int) (string, []byte, error, string) {
	planFile, planBinary, err, logs := generatePlan(cliConfigFile, tempDir, token, apiKey, aud, spaceId)

	if err != nil && retry < 2 && IsFlakyNetworkError(logs) {
		return generatePlanLoop(cliConfigFile, tempDir, token, apiKey, aud, spaceId, retry+1)
	}

	return planFile, planBinary, err, logs
}

func generatePlan(cliConfigFile string, tempDir string, token string, apiKey string, aud string, spaceId string) (string, []byte, error, string) {
	zap.L().Info("Generating plan for " + aud)

	planFile := filepath.Join(tempDir, "tfplan")

	tofu, err := environment.GetTofuExecutable()

	if err != nil {
		return "", nil, err, ""
	}

	stdOut, stdErr, _, err := execute.Execute(
		tofu,
		[]string{
			"-chdir=" + tempDir,
			"plan",
			"-no-color",
			"-out",
			planFile,
			"-var=octopus_space_id=" + spaceId},
		map[string]string{
			"OCTOPUS_ACCESS_TOKEN":        token,
			"OCTOPUS_API_KEY":             apiKey,
			"OCTOPUS_URL":                 aud,
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

	logging.LogEnhanced(stdErr, aud)
	logging.LogEnhanced(stdOut, aud)

	if err != nil {
		return "", nil, errors.New("Failed to generate plan: " + stdErr + " " + stdOut + " " + err.Error()), stdOut + "\n" + stdErr
	}

	plan, err := os.ReadFile(planFile)

	if err != nil {
		return "", nil, err, ""
	}

	return planFile, plan, nil, stdOut + "\n" + stdErr
}

func generatePlanJson(tempDir string, planFile string) (string, error, string) {
	zap.L().Info("Generating plan JSON")

	tofu, err := environment.GetTofuExecutable()

	if err != nil {
		return "", err, ""
	}

	planJsonStdOut, stdErr, _, err := execute.Execute(
		tofu,
		[]string{
			"-chdir=" + tempDir,
			"show",
			"-json",
			"-no-color",
			planFile},
		map[string]string{
			"TF_INPUT":              "0",
			"TF_REATTACH_PROVIDERS": os.Getenv("TF_REATTACH_PROVIDERS"),
		})

	if err != nil {
		zap.L().Error("Failed to generate plan json: "+stdErr+" "+planJsonStdOut, zap.Error(err))
		return "", errors.New("Failed to generate plan json: " + stdErr + " " + planJsonStdOut + " " + err.Error()), planJsonStdOut + "\n" + stdErr
	}

	return planJsonStdOut, nil, planJsonStdOut + "\n" + stdErr
}

func generatePlanText(tempDir string, planFile string) (string, error) {
	zap.L().Info("Generating plan text")

	tofu, err := environment.GetTofuExecutable()

	if err != nil {
		return "", err
	}

	planStdOut, stdErr, _, err := execute.Execute(
		tofu,
		[]string{
			"-chdir=" + tempDir,
			"show",
			"-no-color",
			planFile},
		map[string]string{
			"TF_INPUT":              "0",
			"TF_REATTACH_PROVIDERS": os.Getenv("TF_REATTACH_PROVIDERS"),
		})

	if err != nil {
		zap.L().Error("Failed to generate plan text: "+stdErr+" "+planStdOut, zap.Error(err))
		return "", errors.New("Failed to generate plan text: " + stdErr + " " + planStdOut + " " + err.Error())
	}

	return planStdOut, nil
}

func checkPlan(planJson string, server string) error {
	zap.L().Info("Checking plan with OPA")

	tempDir, err := files.CreateTempDir()

	if err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			zap.L().Error("Failed to remove temporary directory", zap.Error(err))
		}
	}()

	if err := os.WriteFile(filepath.Join(tempDir, "plan.json"), []byte(planJson), 0644); err != nil {
		return err
	}

	policyPath, err := environment.GetCombinedOpaPolicyPath()

	if err != nil {
		return err
	}

	opaPath, err := environment.GetOpaExecutable()

	if err != nil {
		return err
	}

	checkStdOut, _, exitCode, err := execute.Execute(
		opaPath,
		[]string{
			"exec",
			"--fail",
			"--decision",
			"terraform/analysis/allow",
			"--bundle",
			policyPath,
			filepath.Join(tempDir, "plan.json")},
		map[string]string{})

	if err != nil {
		zap.L().Error("Failed to check the plan", zap.Error(err))
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("OPA check failed with exit code %d: %s", exitCode, checkStdOut)
	}

	// Parse the OPA JSON output
	var opaResponse model.OpaResult
	if err := json.Unmarshal([]byte(checkStdOut), &opaResponse); err != nil {
		return fmt.Errorf("failed to parse OPA response: %w", err)
	}

	// Check the result from the parsed JSON
	for _, result := range opaResponse.Result {
		if !result.Result {
			logging.LogEnhanced(planJson, server)
			logging.LogEnhanced(checkStdOut, server)
			return customerrors.OpaValidationFailed{
				ExitCode:   exitCode,
				DecisionID: result.DecisionID,
				Path:       result.Path,
			}
		}
	}

	zap.L().Info(checkStdOut)

	return nil
}
