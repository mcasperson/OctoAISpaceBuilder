// We don't support provisioners like local-exec, even on Octopus resources
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

resource "octopusdeploy_environment" "environment_dev" {
  name                         = "Dev"
  description                  = ""
  allow_dynamic_infrastructure = true
  use_guided_failure           = false

  provisioner "local-exec" {
    command = "echo ${var.octopus_space_id}"
  }
}
