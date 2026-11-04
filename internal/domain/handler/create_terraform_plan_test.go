package handler

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/customerrors"
	"github.com/stretchr/testify/assert"
)

// setupOpaEnvironment configures the environment to use the system installation of OPA
// and the local policy directory.
func setupOpaEnvironment(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	t.Setenv("SPACEBUILDER_OPA_PATH", "opa")
	t.Setenv("SPACEBUILDER_OPA_POLICY_PATH", filepath.Join(cwd, "../../../functions/policy/"))
}

// buildPlan creates a minimal plan JSON document that passes the policy, with the
// supplied resource configurations placed in the root module.
func buildPlan(t *testing.T, rootModule map[string]any) string {
	plan := map[string]any{
		"resource_changes": []any{
			map[string]any{"type": "octopusdeploy_environment"},
		},
		"planned_values": map[string]any{
			"root_module": map[string]any{
				"resources": []any{},
			},
		},
		"configuration": map[string]any{
			"provider_config": map[string]any{
				"octopusdeploy": map[string]any{
					"full_name": "registry.opentofu.org/octopusdeploy/octopusdeploy",
				},
			},
			"root_module": rootModule,
		},
	}

	planJson, err := json.Marshal(plan)

	if err != nil {
		t.Fatalf("Failed to marshal plan: %v", err)
	}

	return string(planJson)
}

func provisioner(provisionerType string) map[string]any {
	return map[string]any{
		"type": provisionerType,
		"expressions": map[string]any{
			"command": map[string]any{"constant_value": "echo hello"},
		},
	}
}

func TestCheckPlanProvisioners(t *testing.T) {
	setupOpaEnvironment(t)

	tests := []struct {
		name       string
		rootModule map[string]any
		allowed    bool
	}{
		{
			name: "allows resources without provisioners",
			rootModule: map[string]any{
				"resources": []any{
					map[string]any{"address": "octopusdeploy_environment.test"},
				},
			},
			allowed: true,
		},
		{
			name: "blocks local-exec provisioner",
			rootModule: map[string]any{
				"resources": []any{
					map[string]any{
						"address":      "octopusdeploy_environment.test",
						"provisioners": []any{provisioner("local-exec")},
					},
				},
			},
			allowed: false,
		},
		{
			name: "blocks remote-exec provisioner",
			rootModule: map[string]any{
				"resources": []any{
					map[string]any{
						"address":      "octopusdeploy_environment.test",
						"provisioners": []any{provisioner("remote-exec")},
					},
				},
			},
			allowed: false,
		},
		{
			name: "blocks file provisioner",
			rootModule: map[string]any{
				"resources": []any{
					map[string]any{
						"address":      "octopusdeploy_environment.test",
						"provisioners": []any{provisioner("file")},
					},
				},
			},
			allowed: false,
		},
		{
			name: "blocks provisioner in nested module",
			rootModule: map[string]any{
				"module_calls": map[string]any{
					"child": map[string]any{
						"module": map[string]any{
							"module_calls": map[string]any{
								"grandchild": map[string]any{
									"module": map[string]any{
										"resources": []any{
											map[string]any{
												"address":      "octopusdeploy_environment.test",
												"provisioners": []any{provisioner("local-exec")},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			allowed: false,
		},
		{
			name: "blocks provisioner when other resources have none",
			rootModule: map[string]any{
				"resources": []any{
					map[string]any{"address": "octopusdeploy_environment.first"},
					map[string]any{
						"address":      "octopusdeploy_environment.second",
						"provisioners": []any{provisioner("local-exec")},
					},
				},
			},
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPlan(buildPlan(t, tt.rootModule), "")

			if tt.allowed {
				assert.NoError(t, err)
			} else {
				var opaErr customerrors.OpaValidationFailed
				assert.True(t, errors.As(err, &opaErr), "expected OpaValidationFailed, got %v", err)
			}
		})
	}
}
