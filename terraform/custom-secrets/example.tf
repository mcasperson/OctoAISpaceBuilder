provider "octopusdeploy" {
  space_id = "${trimspace(var.octopus_space_id)}"
}

terraform {

  required_providers {
    octopusdeploy = { source = "OctopusDeploy/octopusdeploy", version = "1.17.0" }
  }
  required_version = ">= 1.6.0"
}

variable "octopus_space_id" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The ID of the Octopus space to populate."
}

data "octopusdeploy_project_groups" "project_group_kubernetes" {
  ids          = null
  partial_name = "${var.project_group_kubernetes_name}"
  skip         = 0
  take         = 1
}

variable "project_group_kubernetes_name" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The name of the project group to lookup"
  default     = "Kubernetes"
}

resource "octopusdeploy_project_group" "project_group_kubernetes" {
  name  = "${var.project_group_kubernetes_name}"
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_environments" "environment_dev" {
  ids          = null
  partial_name = "Dev"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_environment" "environment_dev" {
  name                         = "Dev"
  description                  = ""
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  jira_extension_settings {
    environment_type = "development"
  }

  jira_service_management_extension_settings {
    is_enabled = true
  }

  servicenow_extension_settings {
    is_enabled = true
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_environments" "environment_development__security_" {
  ids          = null
  partial_name = "Development (Security)"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_environment" "environment_development__security_" {
  name                         = "Development (Security)"
  description                  = "Used to scan the development releases for security issues. This resource is created and managed by the [Octopus Terraform provider](https://registry.terraform.io/providers/OctopusDeploy/octopusdeploy/latest/docs). The Terraform files can be found in the [GitHub repo](https://github.com/mcasperson/AppBuilder-EKS)."
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  jira_extension_settings {
    environment_type = "unmapped"
  }

  jira_service_management_extension_settings {
    is_enabled = false
  }

  servicenow_extension_settings {
    is_enabled = false
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_environments" "environment_production__app_" {
  ids          = null
  partial_name = "Production (App)"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_environment" "environment_production__app_" {
  name                         = "Production (App)"
  description                  = "The production environment."
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  jira_extension_settings {
    environment_type = "unmapped"
  }

  jira_service_management_extension_settings {
    is_enabled = false
  }

  servicenow_extension_settings {
    is_enabled = false
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_environments" "environment_production__security_" {
  ids          = null
  partial_name = "Production (Security)"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_environment" "environment_production__security_" {
  name                         = "Production (Security)"
  description                  = "Used to scan the productions releases for security issues. This resource is created and managed by the [Octopus Terraform provider](https://registry.terraform.io/providers/OctopusDeploy/octopusdeploy/latest/docs). The Terraform files can be found in the [GitHub repo](https://github.com/mcasperson/AppBuilder-EKS)."
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  jira_extension_settings {
    environment_type = "unmapped"
  }

  jira_service_management_extension_settings {
    is_enabled = false
  }

  servicenow_extension_settings {
    is_enabled = false
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_lifecycles" "lifecycle_application_and_security" {
  ids          = null
  partial_name = "Application and Security"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_lifecycle" "lifecycle_application_and_security" {
  name        = "Application and Security"
  description = ""

  phase {
    automatic_deployment_targets          = []
    optional_deployment_targets           = [octopusdeploy_environment.environment_dev.id]
    name                                  = "Development"
    is_optional_phase                     = false
    minimum_environments_before_promotion = 0
  }
  phase {
    automatic_deployment_targets          = [octopusdeploy_environment.environment_development__security_.id]
    optional_deployment_targets           = []
    name                                  = "Development Security"
    is_optional_phase                     = false
    minimum_environments_before_promotion = 0
  }
  phase {
    automatic_deployment_targets          = []
    optional_deployment_targets           = [octopusdeploy_environment.environment_production__app_.id]
    name                                  = "Production"
    is_optional_phase                     = false
    minimum_environments_before_promotion = 0
  }
  phase {
    automatic_deployment_targets          = [octopusdeploy_environment.environment_production__security_.id]
    optional_deployment_targets           = []
    name                                  = "Production Security"
    is_optional_phase                     = false
    minimum_environments_before_promotion = 0
  }

  release_retention_with_strategy {
    strategy         = "Count"
    quantity_to_keep = 30
    unit             = "Days"
  }

  tentacle_retention_with_strategy {
    strategy         = "Count"
    quantity_to_keep = 30
    unit             = "Days"
  }

  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_environments" "environment_test" {
  ids          = null
  partial_name = "Test"
  skip         = 0
  take         = 1
}
resource "octopusdeploy_environment" "environment_test" {
  name                         = "Test"
  description                  = ""
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  jira_extension_settings {
    environment_type = "unmapped"
  }

  jira_service_management_extension_settings {
    is_enabled = false
  }

  servicenow_extension_settings {
    is_enabled = false
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_environments" "environment_production" {
  ids          = null
  partial_name = "Production"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_environment" "environment_production" {
  name                         = "Production"
  description                  = ""
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  jira_extension_settings {
    environment_type = "unmapped"
  }

  jira_service_management_extension_settings {
    is_enabled = false
  }

  servicenow_extension_settings {
    is_enabled = false
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "octopusdeploy_lifecycles" "lifecycle_default_lifecycle" {
  ids          = null
  partial_name = "Default Lifecycle"
  skip         = 0
  take         = 1
  lifecycle {
    postcondition {
      error_message = "Failed to resolve a lifecycle called \"Default Lifecycle\". This resource must exist in the space before this Terraform configuration is applied."
      condition     = length(self.lifecycles) != 0
    }
  }
}

data "octopusdeploy_channels" "channel_my_k8s_project_2_default" {
  ids          = []
  partial_name = "Default"
  skip         = 0
  take         = 1
}

data "octopusdeploy_feeds" "feed_octopus_server__built_in_" {
  feed_type    = "BuiltIn"
  ids          = null
  partial_name = ""
  skip         = 0
  take         = 1
  lifecycle {
    postcondition {
      error_message = "Failed to resolve a feed called \"BuiltIn\". This resource must exist in the space before this Terraform configuration is applied."
      condition     = length(self.feeds) != 0
    }
  }
}

data "octopusdeploy_feeds" "feed_github_container_registry" {
  feed_type    = "Docker"
  ids          = null
  partial_name = "GitHub Container Registry"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_docker_container_registry" "feed_github_container_registry" {
  name                                 = "GitHub Container Registry"
  registry_path                        = ""
  api_version                          = "v2"
  feed_uri                             = "https://ghcr.io"
  package_acquisition_location_options = ["ExecutionTarget", "NotAcquired"]
  lifecycle {
    ignore_changes  = [password]
    prevent_destroy = true
  }
}

data "octopusdeploy_worker_pools" "workerpool_default_worker_pool" {
  ids          = null
  partial_name = "Default Worker Pool"
  skip         = 0
  take         = 1
}

resource "octopusdeploy_process" "test" {
  project_id = octopusdeploy_project.project_my_k8s_project_2.id
  depends_on = []
}

resource "octopusdeploy_process_steps_order" "test" {
  process_id = "${octopusdeploy_process.test.id}"
  steps      = ["${octopusdeploy_process_step.process_step_get_mysql_host.id}"]
}

resource "octopusdeploy_process_step" "process_step_get_mysql_host" {
  name                  = "Get MySQL Host"
  type                  = "Octopus.KubernetesRunScript"
  process_id            = "${octopusdeploy_process.test.id}"
  channels              = null
  condition             = "Success"
  environments          = null
  excluded_environments = null
  package_requirement   = "LetOctopusDecide"
  slug                  = "get-mysql-host"
  start_trigger         = "StartAfterPrevious"
  tenant_tags           = null
  execution_properties  = {
    "Octopus.Action.Script.Syntax" = "PowerShell"
    "Octopus.Action.Script.ScriptBody" = "echo \"hi\""
    "Octopus.Action.RunOnServer" = "true"
    "Octopus.Action.Script.ScriptSource" = "Inline"
  }
  properties            = {
    "Octopus.Action.TargetRoles" = "eks"
  }
}

variable "variable_42a2de0c9187a3d773ca5ee10a26490b4e720a0edd470e2907fbe2a77f633531_sensitive_value" {
  type        = string
  nullable    = true
  sensitive   = true
  description = "The secret variable value associated with the variable SecretVariable"
  default     = "A custom secret value"
}
resource "octopusdeploy_variable" "my_k8s_project_2_secretvariable_1" {
  owner_id        = octopusdeploy_project.project_my_k8s_project_2.id
  name            = "SecretVariable"
  type            = "Sensitive"
  is_sensitive    = true
  sensitive_value = var.variable_42a2de0c9187a3d773ca5ee10a26490b4e720a0edd470e2907fbe2a77f633531_sensitive_value
  lifecycle {
    ignore_changes  = [sensitive_value]
    prevent_destroy = true
  }
  depends_on = []
}

variable "project_my_k8s_project_2_name" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The name of the project exported from My K8s Project 2"
  default     = "My K8s Project 2"
}
variable "project_my_k8s_project_2_description_prefix" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "An optional prefix to add to the project description for the project My K8s Project 2"
  default     = ""
}
variable "project_my_k8s_project_2_description_suffix" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "An optional suffix to add to the project description for the project My K8s Project 2"
  default     = ""
}
variable "project_my_k8s_project_2_description" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The description of the project exported from My K8s Project 2"
  default     = "This project provides an example Kubernetes deployment using YAML, Cloud Target Discovery, and SBOM scanning to an AWS EKS Kubernetes cluster."
}
variable "project_my_k8s_project_2_tenanted" {
  type        = string
  nullable    = false
  sensitive   = false
  description = "The tenanted setting for the project Untenanted"
  default     = "Untenanted"
}
data "octopusdeploy_projects" "project_my_k8s_project_2" {
  ids          = null
  partial_name = "${var.project_my_k8s_project_2_name}"
  skip         = 0
  take         = 1
}
resource "octopusdeploy_project" "project_my_k8s_project_2" {
  name                                 = "${var.project_my_k8s_project_2_name}"
  auto_create_release                  = false
  default_guided_failure_mode          = "EnvironmentDefault"
  default_to_skip_if_already_installed = false
  discrete_channel_release             = false
  is_disabled                          = false
  is_version_controlled                = false
  lifecycle_id                         = octopusdeploy_lifecycle.lifecycle_application_and_security.id
  project_group_id                     = octopusdeploy_project_group.project_group_kubernetes.id
  included_library_variable_sets       = []
  tenanted_deployment_participation    = "${var.project_my_k8s_project_2_tenanted}"

  connectivity_policy {
    allow_deployments_to_no_targets = true
    exclude_unhealthy_targets       = false
    skip_machine_behavior           = "None"
  }

  versioning_strategy {
    template = "#{Octopus.Version.LastMajor}.#{Octopus.Version.LastMinor}.#{Octopus.Version.NextPatch}"
  }
  description = "${var.project_my_k8s_project_2_description_prefix}${var.project_my_k8s_project_2_description}${var.project_my_k8s_project_2_description_suffix}"
  lifecycle {
    prevent_destroy = true
  }
}

