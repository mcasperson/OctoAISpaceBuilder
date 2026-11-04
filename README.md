![Coverage](https://img.shields.io/badge/Coverage-49.8%25-yellow)

This microservice implements a web app, designed to be deployed as an Azure function, that generates a Terraform plan for an Octopus space with a call to `/api/terraformplan`, stores the plan, and then applies it with a call to `/api/terraformapply`. 

This services forces the use of a local backend, which is not persisted between calls. The Terraform configuration is expected to be stateless. See [Octoterra](https://github.com/OctopusSolutionsEngineering/OctopusTerraformExport) for more details on stateless Terraform configurations.

## Security Considerations

* The use of local state is enforced with overrides.
* Plugins must use a local mirror with the CLI configuration, so no providers are downloaded.
* OPA enforces rules that only allow Octopus resources to be created.
* OPA enforces rules that prohibit sensitive values being defined in the configuration.
* The client can only send Terraform configuration. The service defines all the file names (i.e. the service controls overrides).
* The client can not send variable values as separate files.
* Old plan records are deleted after 5 minutes.
* We use a custom build of OpenTofu which has a limited number of function.

## OpenTofu fork

The forked version of OpenTofu is available [here](https://github.com/OctopusSolutionsEngineering/opentofu).

Create a release of the fork with the workflow [here](https://github.com/OctopusSolutionsEngineering/opentofu/actions/workflows/manual_release.yaml). This builds the OpenTofu binary and saves it as a GitHub release.

The new version of OpenTofu is them referenced in the workflow [here](.github/workflows/buil.yaml).

## Terraform Provider

We use a forked version of the Octopus Terraform Provider to support routing requests to the Azure Function Router. The
provider is maintained in the branch [mattc/spacebuilder](https://github.com/OctopusDeploy/terraform-provider-octopusdeploy/pull/19).

This branch is pulled and built by the workflow.

## Testing

Start the Azurite emulator with the following command:

```bash
docker run -d -p 10000:10000 -p 10001:10001 -p 10002:10002 mcr.microsoft.com/azure-storage/azurite azurite --skipApiVersionCheck
```

This is the command in MacOs:

```bash
docker run -d --name azurite -p 10000:10000 -p 10001:10001 -p 10002:10002 \
  --restart unless-stopped \
  mcr.microsoft.com/azure-storage/azurite \
  azurite --skipApiVersionCheck --blobHost 0.0.0.0 --queueHost 0.0.0.0 --tableHost 0.0.0.0
```

Set the `AzureWebJobsStorage` environment variable to:

```
DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;
```

The following env vars allow the service to run locally:
```bash
export DISABLE_VALIDATION=true
export ENHANCED_LOGGING_INSTANCES='["yourinstance.octopus.app"]'
export SPACEBUILDER_FUNCTIONS_CUSTOMHANDLER_PORT=8084
export SPACEBUILDER_OPA_PATH=opa
export SPACEBUILDER_OPA_POLICY_PATH=functions/policy
export SPACEBUILDER_DISABLE_TERRAFORM_CLI_CONFIG=true
export SPACEBUILDER_TOFU_PATH=tofu
export DISABLE_BINARIES_EXECUTABLE=true
```

You will also need to install [Tofu](https://opentofu.org/docs/intro/install/) and [OPA](https://www.openpolicyagent.org/docs/latest/#running-opa) locally.