package application

import (
	"io"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/application/responses"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/handler"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/logging"
	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateTerraformApply(c *gin.Context) {
	success := false
	defer func() {
		terraformApplyMetrics.Record(success)
	}()

	body, err := io.ReadAll(c.Request.Body)

	if err != nil {
		zap.L().Error("Failed to read request body", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, responses.GenerateError("Failed to read request body", err))
		return
	}

	var terraform model.TerraformApply
	err = jsonapi.Unmarshal(body, &terraform)

	if err != nil {
		zap.L().Error("Failed to unmarshal JSON API body", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, responses.GenerateError("Failed to unmarshal JSON API body", err))
		return
	}

	server, token, apiKey, err := getServerTokenApiKey(c)

	if err != nil {
		zap.L().Error("Failed to get the Octopus details", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, responses.GenerateError("Failed to get the Octopus details", err))
		return
	}

	response, err := createTerraformApply(server, token, apiKey, terraform, 0)

	if err != nil {
		zap.L().Error("Failed to perform Terraform apply", zap.Error(err))
		c.IndentedJSON(http.StatusInternalServerError, responses.GenerateError("Failed to perform Terraform apply", err))
		return
	}

	responseJSON, err := jsonapi.Marshal(response)

	if err != nil {
		zap.L().Error("Failed to marshal JSON API response", zap.Error(err))
		c.IndentedJSON(http.StatusInternalServerError, responses.GenerateError("Failed to marshal JSON API response", err))
		return
	}

	c.Header("Content-Type", "application/vnd.api+json")
	success = true
	c.String(http.StatusCreated, string(responseJSON))
}

func createTerraformApply(server, token, apiKey string, terraform model.TerraformApply, retry int) (*model.TerraformApply, error) {
	response, err, output := handler.CreateTerraformApply(server, token, apiKey, terraform)

	// Retry if there was a connection failure
	if err != nil && handler.IsFlakyNetworkError(output) && retry < 2 {
		return createTerraformApply(server, token, apiKey, terraform, retry+1)
	}

	logging.SaveEnhancedLogs("apply", output, server, err == nil)

	return response, err
}
