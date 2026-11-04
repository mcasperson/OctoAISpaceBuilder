package environment

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDisableTerraformCliConfig(t *testing.T) {
	// Save original environment variable to restore later
	originalValue := os.Getenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG")

	// Clean up after the test
	defer func() {
		assert.NoError(t, os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", originalValue))
	}()

	tests := []struct {
		name           string
		envValue       string
		expectedResult bool
	}{
		{
			name:           "returns true when environment variable is 'true'",
			envValue:       "true",
			expectedResult: true,
		},
		{
			name:           "returns false when environment variable is 'false'",
			envValue:       "false",
			expectedResult: false,
		},
		{
			name:           "returns false when environment variable is empty",
			envValue:       "",
			expectedResult: false,
		},
		{
			name:           "returns false when environment variable has other value",
			envValue:       "1",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for this test case
			assert.NoError(t, os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", tt.envValue))

			// Call the function under test
			result := GetDisableTerraformCliConfig()

			// Assert
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetTerraformProviderVersion(t *testing.T) {
	assert.Equal(t, "1.18.2", GetTerraformProviderVersion())
}
