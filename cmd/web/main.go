package main

import (
	"fmt"
	"os"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/application"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/logging"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/validation"
	"go.uber.org/zap"
)

func main() {
	logging.ConfigureZapLogger()

	cwd, err := os.Getwd()
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	installDir, err := environment.GetInstallationDirectory()
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	tofu, _ := environment.GetTofuExecutable()
	opa, _ := environment.GetOpaExecutable()

	zap.L().Info("Current working directory: " + cwd)
	zap.L().Info("Home directory: " + homeDir)
	zap.L().Info("Install directory: " + installDir)
	zap.L().Info("Disable validation: " + fmt.Sprint(environment.DisableValidation()))
	zap.L().Info("OPA executable: " + fmt.Sprint(opa))
	zap.L().Info("OPA policy path: " + fmt.Sprint(environment.GetOpaPolicyPath()))
	zap.L().Info("Tofu providers path: " + fmt.Sprint(environment.GetTerraformProvidersPath()))
	zap.L().Info("Tofu executable: " + fmt.Sprint(tofu))
	zap.L().Info("Tofu provider version: " + environment.GetTerraformProviderVersion())
	zap.L().Info("Disable setting binary execution flag: " + fmt.Sprint(environment.DisableMakeBinariesExecutable()))
	zap.L().Info("Disable Terraform config: " + fmt.Sprint(environment.GetDisableTerraformCliConfig()))
	zap.L().Info("Enhanced logging instance: " + fmt.Sprint(environment.GetEnhancedLoggingInstances()))
	zap.L().Info("Force the redirector: " + fmt.Sprint(environment.GetRedirectionForce()))
	zap.L().Info("Redirector bypass: " + fmt.Sprint(environment.GetRedirectionBypass()))
	zap.L().Info("Port: " + fmt.Sprint(environment.GetPort()))
	zap.L().Info("Persist Logs: " + fmt.Sprint(environment.GetPersistEnhancedLogs()))

	/*
		Validate that the required files are present. This is important because missing files lead to strange behaviour.
		For example, missing policy files leads to OPA hanging indefinitely.
		This helps people running the service locally to ensure they have the correct files in place.
	*/
	if !environment.IsInAzureFunctions() {
		if err := validation.TestOpaPolicyInstallation(); err != nil {
			zap.L().Error(err.Error())
			return
		}

		if !environment.GetDisableTerraformCliConfig() {
			if err := validation.TestFileSystemProviderInstallation(); err != nil {
				zap.L().Error(err.Error())
				return
			}
		}
	}

	if err := application.StartServer(); err != nil {
		zap.L().Error(err.Error())
	}
}
