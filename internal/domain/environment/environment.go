package environment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func GetPort() string {
	// Get the port from the environment variable
	port := os.Getenv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if port == "" {
		port = os.Getenv("SPACEBUILDER_FUNCTIONS_CUSTOMHANDLER_PORT")
		if port == "" {
			port = "8080" // Default port
		}
	}
	return port
}

func GetRedirectionBypass() []string {
	hostnames := []string{}
	hostnamesJson := os.Getenv("REDIRECTION_BYPASS")
	if hostnamesJson == "" {
		return []string{} // Default to empty slice if not set
	}

	err := json.Unmarshal([]byte(hostnamesJson), &hostnames)
	if err != nil {
		zap.L().Error("Error parsing JSON:", zap.Error(err))
		return []string{}
	}

	return hostnames
}

func GetRedirectionForce() bool {
	redirectionForce := os.Getenv("REDIRECTION_FORCE")
	return strings.ToLower(redirectionForce) == "true"
}

func IsInAzureFunctions() bool {
	return os.Getenv("FUNCTIONS_WORKER_RUNTIME") == "custom"
}

func GetInstallationDirectory() (string, error) {
	if os.Getenv("SPACEBUILDER_INSTALLATION_DIRECTORY") != "" {
		// If the environment variable is set, use it as the installation directory
		return os.Getenv("SPACEBUILDER_INSTALLATION_DIRECTORY"), nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if IsInAzureFunctions() {
		// In Azure Functions, all files are copied to the temp location
		return os.TempDir(), nil
	}

	// When running locally, the policies and other files are in the "functions" directory.
	// This assumes that most people will run this service locally with the working directory set to the root of the project.
	// You can also define absolute paths for the "policy" and "provider" directories in the environment variables
	// SPACEBUILDER_TERRAFORM_PROVIDERS and SPACEBUILDER_OPA_POLICY_PATH.
	// Note that tests will often use the current working directory of their test file. This means tests need to set
	// absolute paths for the policy and provider directories.
	return filepath.Join(cwd, "functions"), nil
}

func DisableValidation() bool {
	// Get the disable validation flag from the environment variable
	disableValidation := os.Getenv("DISABLE_VALIDATION")
	if disableValidation == "" {
		return false // Default to false if not set
	}
	return strings.ToLower(disableValidation) == "true"
}

// GetEnhancedLoggingInstances returns a list of instances for enhanced logging.
// This allows developers to debug their issues without logging prompt responses for anyone else.
func GetEnhancedLoggingInstances() []string {
	instances := []string{}
	instancesJson := os.Getenv("ENHANCED_LOGGING_INSTANCES")
	if instancesJson == "" {
		return []string{} // Default to empty slice if not set
	}

	err := json.Unmarshal([]byte(instancesJson), &instances)
	if err != nil {
		zap.L().Error("Error parsing JSON:", zap.Error(err))
		return []string{}
	}

	return instances
}

// GetPersistEnhancedLogs returns true if the service should save enhanced logs to disk, or false otherwise.
func GetPersistEnhancedLogs() bool {
	return strings.ToLower(os.Getenv("PERSIST_ENHANCED_LOGS")) == "true"
}

// GetLogFilePath returns the path to the log file, or an empty string if file logging is disabled.
// Set the SPACEBUILDER_LOG_FILE environment variable to enable file logging (e.g. /var/log/spacebuilder.log).
func GetLogFilePath() string {
	return os.Getenv("SPACEBUILDER_LOG_FILE")
}
