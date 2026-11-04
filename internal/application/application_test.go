package application

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/files"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/logging"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/model"
	"github.com/OctopusSolutionsEngineering/OctopusTerraformTestFramework/v2/test"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestTerraformPlanAndApplyEndpoint plans and then applies a terraform configuration via the application
// level Gin interface.
func TestTerraformPlanAndApplyEndpoint(t *testing.T) {
	logging.ConfigureZapLogger()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Setenv("AzureWebJobsStorage", "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"); err != nil {
		t.Fatalf("Failed to set AzureWebJobsStorage: %v", err)
	}

	// Use the system installation of OPA
	if err := os.Setenv("SPACEBUILDER_OPA_PATH", "opa"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_PATH: %v", err)
	}

	// Use the system installation of Tofu
	if err := os.Setenv("SPACEBUILDER_TOFU_PATH", "tofu"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_TOFU_PATH: %v", err)
	}

	// Disable the Terraform CLI config to allow provider downloads
	if err := os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", "true"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG: %v", err)
	}

	// Set the OPA policy path to the local policy directory
	policyPath := filepath.Join(cwd, "../../functions/policy/")
	if err := os.Setenv("SPACEBUILDER_OPA_POLICY_PATH", policyPath); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_POLICY_PATH: %v", err)
	}

	base, err := files.CopyDir("../../terraform")

	if err != nil {
		t.Fatalf("Failed to copy Terraform files: %v", err)
	}

	testFramework := test.OctopusContainerTest{}
	testFramework.ArrangeTest(t, func(t *testing.T, container *test.OctopusContainer, client *client.Client) error {
		spaceId, err := testFramework.Act(t, container, base, "2-localsetup", []string{})

		if err != nil {
			t.Fatalf("Failed to create space: %v", err)
		}

		// Set up Gin in test mode
		gin.SetMode(gin.TestMode)

		// Create a test router with the endpoint
		router := gin.Default()

		// Register the endpoint with mocked middleware
		router.POST("/api/terraformplan", CreateTerraformPlan)
		router.POST("/api/terraformapply", CreateTerraformApply)

		reponse, err := func() (*model.TerraformPlan, error) {
			configuration, err := os.ReadFile("../../terraform/k8s-example/example.tf")

			if err != nil {
				return nil, err
			}

			// Apply the changes
			body := model.TerraformPlan{
				ID:            "unused",
				SpaceId:       spaceId,
				Configuration: string(configuration),
			}

			jsonBody, err := jsonapi.Marshal(body)

			if err != nil {
				return nil, err
			}

			// Create test request
			req, err := http.NewRequest(http.MethodPost, "/api/terraformplan", bytes.NewBuffer(jsonBody))

			if err != nil {
				return nil, err
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Octopus-Url", container.URI)
			req.Header.Set("X-Octopus-ApiKey", test.ApiKey)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, 201, w.Code, w.Body.String())

			var response model.TerraformPlan
			err = jsonapi.Unmarshal(w.Body.Bytes(), &response)

			if err != nil {
				return nil, err
			}

			assert.NotEmpty(t, response.ID)
			assert.Equal(t, spaceId, response.SpaceId)

			return &response, nil
		}()

		if err != nil {
			return err
		}

		// Assert response
		err = func(response *model.TerraformPlan) error {
			// Plan the changes
			applyBody := model.TerraformApply{
				ID:        "unused",
				PlanId:    response.ID,
				Server:    "",
				ApplyText: nil,
			}

			applyJsonBody, err := jsonapi.Marshal(applyBody)

			if err != nil {
				return err
			}

			// Create test request
			applyReq, err := http.NewRequest(http.MethodPost, "/api/terraformapply", bytes.NewBuffer(applyJsonBody))

			if err != nil {
				return err
			}

			applyReq.Header.Set("Content-Type", "application/json")
			applyReq.Header.Set("X-Octopus-Url", container.URI)
			applyReq.Header.Set("X-Octopus-ApiKey", test.ApiKey)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, applyReq)

			// Assert status code
			assert.Equal(t, 201, w.Code, w.Body.String())

			// Assert response
			var applyResponse model.TerraformApply
			err = jsonapi.Unmarshal(w.Body.Bytes(), &applyResponse)

			if err != nil {
				fmt.Println("Terraform apply response was " + w.Body.String())
				return err
			}

			assert.NotEmpty(t, applyResponse.ID)

			return nil
		}(reponse)

		return err
	})
}

// TestTerraformAutoApplyEndpoint autoapplies a terraform configuration via the application
// level Gin interface.
func TestTerraformAutoApplyEndpoint(t *testing.T) {
	logging.ConfigureZapLogger()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Setenv("AzureWebJobsStorage", "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"); err != nil {
		t.Fatalf("Failed to set AzureWebJobsStorage: %v", err)
	}

	// Use the system installation of OPA
	if err := os.Setenv("SPACEBUILDER_OPA_PATH", "opa"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_PATH: %v", err)
	}

	// Use the system installation of Tofu
	if err := os.Setenv("SPACEBUILDER_TOFU_PATH", "tofu"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_TOFU_PATH: %v", err)
	}

	// Disable the Terraform CLI config to allow provider downloads
	if err := os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", "true"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG: %v", err)
	}

	// Set the OPA policy path to the local policy directory
	policyPath := filepath.Join(cwd, "../../functions/policy/")
	if err := os.Setenv("SPACEBUILDER_OPA_POLICY_PATH", policyPath); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_POLICY_PATH: %v", err)
	}

	base, err := files.CopyDir("../../terraform")

	if err != nil {
		t.Fatalf("Failed to copy Terraform files: %v", err)
	}

	testFramework := test.OctopusContainerTest{}
	testFramework.ArrangeTest(t, func(t *testing.T, container *test.OctopusContainer, client *client.Client) error {
		spaceId, err := testFramework.Act(t, container, base, "2-localsetup", []string{})

		if err != nil {
			t.Fatalf("Failed to create space: %v", err)
		}

		// Set up Gin in test mode
		gin.SetMode(gin.TestMode)

		// Create a test router with the endpoint
		router := gin.Default()

		// Register the endpoint with mocked middleware
		router.POST("/api/terraformautoapply", CreateTerraformAutoApply)

		err = func() error {
			configuration, err := os.ReadFile("../../terraform/k8s-example2/example.tf")

			if err != nil {
				return err
			}

			// Apply the changes
			body := model.TerraformPlan{
				ID:            "unused",
				SpaceId:       spaceId,
				Configuration: string(configuration),
			}

			jsonBody, err := jsonapi.Marshal(body)

			if err != nil {
				return err
			}

			// Create test request
			req, err := http.NewRequest(http.MethodPost, "/api/terraformautoapply", bytes.NewBuffer(jsonBody))

			if err != nil {
				return err
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Octopus-Url", container.URI)
			req.Header.Set("X-Octopus-ApiKey", test.ApiKey)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, 201, w.Code, w.Body.String())

			var response model.TerraformApply
			err = jsonapi.Unmarshal(w.Body.Bytes(), &response)

			if err != nil {
				return err
			}

			assert.NotEmpty(t, response.ID)

			return nil
		}()

		return err
	})
}

func TestHealthEndpoint(t *testing.T) {
	logging.ConfigureZapLogger()

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with the endpoint
	router := gin.Default()

	// Register the endpoint with mocked middleware
	router.GET("/api/health", Health)

	// Create test request
	req, err := http.NewRequest(http.MethodGet, "/api/health", nil)

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, 200, w.Code)

	var response model.Health
	err = jsonapi.Unmarshal(w.Body.Bytes(), &response)

	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	assert.NotEmpty(t, response.ID)
}

func TestRootEndpoint(t *testing.T) {
	logging.ConfigureZapLogger()

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with the endpoint
	router := gin.Default()

	// Register the endpoint with mocked middleware
	router.GET("/", Health)

	// Create test request
	req, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, 200, w.Code)

	var response model.Health
	err = jsonapi.Unmarshal(w.Body.Bytes(), &response)

	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	assert.NotEmpty(t, response.ID)
}
