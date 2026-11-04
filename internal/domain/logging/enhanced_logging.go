package logging

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func getHostname(server string) string {
	parsedURL, err := url.Parse(server)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}

func IsEnhancedLoggingEnabled(server string) bool {
	servers := environment.GetEnhancedLoggingInstances()
	return lo.Contains(servers, getHostname(server))
}

// LogEnhanced will log a message only if the server that made the request is specifically configured
// for enhanced logging.
func LogEnhanced(log string, server string) {
	enhancedLogging := IsEnhancedLoggingEnabled(server)

	if enhancedLogging {
		// Azure will cut off long lines, so split and print each line separately
		for _, line := range strings.Split(log, "\n") {
			zap.L().Info(line)
		}
	}
}

func SaveEnhanced(content string, server string, success bool) {
	enhancedLogging := IsEnhancedLoggingEnabled(server)
	shouldPersist := environment.GetPersistEnhancedLogs()

	if !enhancedLogging || !shouldPersist {
		return
	}

	hostname := getHostname(server)
	if hostname == "" {
		hostname = "unknown"
	}

	successLabel := lo.Ternary(success, "success", "failure")

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.tf", hostname, timestamp, successLabel)

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		zap.L().Error("Failed to save enhanced log file", zap.String("filename", filename), zap.Error(err))
	}
}

func SaveEnhancedLogs(operation string, content string, server string, success bool) {
	enhancedLogging := IsEnhancedLoggingEnabled(server)
	shouldPersist := environment.GetPersistEnhancedLogs()

	if !enhancedLogging || !shouldPersist {
		return
	}

	hostname := getHostname(server)
	if hostname == "" {
		hostname = "unknown"
	}

	successLabel := lo.Ternary(success, "success", "failure")

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s_%s.log", operation, hostname, timestamp, successLabel)

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		zap.L().Error("Failed to save enhanced log file", zap.String("filename", filename), zap.Error(err))
	}
}
