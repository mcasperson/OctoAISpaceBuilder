package domain

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OctopusDeploy/go-octopusdeploy/v2/pkg/client"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/customerrors"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/files"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/handler"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/logging"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/model"
	"github.com/OctopusSolutionsEngineering/OctopusTerraformTestFramework/v2/test"
)

// TestPopulateSpaceWithK8sProject creates a space and populates it via the domain level handlers.
func TestPopulateSpaceWithK8sProject(t *testing.T) {
	logging.ConfigureZapLogger()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Setenv("AzureWebJobsStorage", "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"); err != nil {
		t.Fatalf("Failed to set AzureWebJobsStorage: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_OPA_PATH", "opa"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_PATH: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_TOFU_PATH", "tofu"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_TOFU_PATH: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", "true"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG: %v", err)
	}

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
			return err
		}

		configuration, err := os.ReadFile("../../terraform/k8s-example/example.tf")

		if err != nil {
			return err
		}

		plan, err, _ := handler.CreateTerraformPlan(container.URI, "", test.ApiKey, model.TerraformPlan{
			ID:               "",
			PlanBinaryBase64: nil,
			PlanText:         nil,
			Server:           "",
			Created:          time.Time{},
			SpaceId:          spaceId,
			Configuration:    string(configuration),
		})

		if plan != nil {
			t.Log(*plan.PlanText)
		}

		if err != nil {
			return err
		}

		apply, err, _ := handler.CreateTerraformApply(container.URI, "", test.ApiKey, model.TerraformApply{
			ID:        "",
			PlanId:    plan.ID,
			Server:    "",
			ApplyText: nil,
		})

		if apply != nil {
			t.Log(*apply.ApplyText)
		}

		if err != nil {
			return err
		}

		return nil
	})
}

// TestInvalidCustomSecretsProject attempts to create a space and populate it with a project that has custom secrets.
// This must fail, as we do not support sending sensitive values.
func TestInvalidCustomSecretsProject(t *testing.T) {
	logging.ConfigureZapLogger()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Setenv("AzureWebJobsStorage", "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;"); err != nil {
		t.Fatalf("Failed to set AzureWebJobsStorage: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_OPA_PATH", "opa"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_PATH: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_TOFU_PATH", "tofu"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_TOFU_PATH: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", "true"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG: %v", err)
	}

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
			return err
		}

		configuration, err := os.ReadFile("../../terraform/custom-secrets/example.tf")

		if err != nil {
			return err
		}

		plan, err, _ := handler.CreateTerraformPlan(container.URI, "", test.ApiKey, model.TerraformPlan{
			ID:               "",
			PlanBinaryBase64: nil,
			PlanText:         nil,
			Server:           "",
			Created:          time.Time{},
			SpaceId:          spaceId,
			Configuration:    string(configuration),
		})

		if plan != nil {
			t.Log(*plan.PlanText)
		}

		if err == nil {
			t.Fatalf("CreateTerraformPlan should have failed")
		}

		cusErr := &customerrors.OpaValidationFailed{}
		if !errors.As(err, cusErr) {
			t.Fatalf("CreateTerraformPlan should have failed with OpaValidationFailed")
		}

		return nil
	})
}

// TestInvalidProviderProject attempts to create resources using prohibited providers.
func TestInvalidProviderProject(t *testing.T) {
	logging.ConfigureZapLogger()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_OPA_PATH", "opa"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_PATH: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_TOFU_PATH", "tofu"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_TOFU_PATH: %v", err)
	}

	if err := os.Setenv("SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG", "true"); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG: %v", err)
	}

	policyPath := filepath.Join(cwd, "../../functions/policy/")
	if err := os.Setenv("SPACEBUILDER_OPA_POLICY_PATH", policyPath); err != nil {
		t.Fatalf("Failed to set SPACEBUILDER_OPA_POLICY_PATH: %v", err)
	}

	configuration, err := os.ReadFile("../../terraform/localfile-example/example.tf")

	if err != nil {
		t.Fatalf("Failed to read terraform configuration: %v", err)
	}

	plan, err, _ := handler.CreateTerraformPlan("http://example.org", "", test.ApiKey, model.TerraformPlan{
		ID:               "",
		PlanBinaryBase64: nil,
		PlanText:         nil,
		Server:           "",
		Created:          time.Time{},
		SpaceId:          "Spaces-1",
		Configuration:    string(configuration),
	})

	if plan != nil {
		t.Log(*plan.PlanText)
	}

	if err == nil {
		t.Fatalf("CreateTerraformPlan should have failed")
	}

	cusErr := &customerrors.OpaValidationFailed{}
	if !errors.As(err, cusErr) {
		t.Fatalf("CreateTerraformPlan should have failed with OpaValidationFailed")
	}
}
