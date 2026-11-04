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

func CreateTerraformPlan(c *gin.Context) {
	success := false
	defer func() {
		terraformPlanMetrics.Record(success)
	}()

	body, err := io.ReadAll(c.Request.Body)

	if err != nil {
		zap.L().Error("Failed to read request body", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, responses.GenerateError("Failed to process request", err))
		return
	}

	var terraformInput model.TerraformPlan
	err = jsonapi.Unmarshal(body, &terraformInput)

	if err != nil {
		zap.L().Error("Failed to unmarshal body as JSON API", zap.Error(err))
		c.IndentedJSON(http.StatusBadRequest, responses.GenerateError("Failed to process request", err))
		return
	}

	server, token, apiKey, err := getServerTokenApiKey(c)

	response, err, output := handler.CreateTerraformPlan(server, token, apiKey, terraformInput)

	logging.SaveEnhancedLogs("plan", output, server, err == nil)
	logging.SaveEnhanced(terraformInput.Configuration, server, err == nil)

	if err != nil {
		zap.L().Error("Failed to create Terraform plan", zap.Error(err))
		c.IndentedJSON(http.StatusInternalServerError, responses.GenerateError("Failed to process request", err))
		return
	}

	responseJSON, err := jsonapi.Marshal(response)

	if err != nil {
		zap.L().Error("Failed to read request body", zap.Error(err))
		c.IndentedJSON(http.StatusInternalServerError, responses.GenerateError("Failed to process request", err))
		return
	}

	c.Header("Content-Type", "application/vnd.api+json")
	success = true
	c.String(http.StatusCreated, string(responseJSON))
}
