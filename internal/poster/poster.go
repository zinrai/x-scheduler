package poster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/zinrai/x-scheduler/pkg/logger"
)

// Posts content to X using xurl command
func Post(content string) error {
	// Create JSON payload using safe marshaling
	reqBody := struct {
		Text string `json:"text"`
	}{
		Text: content,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	cmd := exec.Command("xurl", "-X", "POST", "/2/tweets", "-d", string(jsonBytes))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Debug("Executing xurl command: %v", cmd.Args)
	logger.Debug("JSON payload: %s", string(jsonBytes))

	if err := cmd.Run(); err != nil {
		logger.Error("xurl command failed: %v", err)
		logger.Error("xurl stderr: %s", stderr.String())
		return fmt.Errorf("xurl failed: %w, stderr: %s", err, stderr.String())
	}

	logger.Debug("xurl stdout: %s", stdout.String())
	return nil
}

// Checks if xurl command is available and working
func Validate() error {
	// Check if xurl command exists
	if _, err := exec.LookPath("xurl"); err != nil {
		return fmt.Errorf("xurl command not found: %w", err)
	}

	logger.Debug("xurl command found")

	// Basic functionality check by running help command
	cmd := exec.Command("xurl", "--help")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Error("xurl command validation failed: %v", err)
		logger.Error("xurl stderr: %s", stderr.String())
		return fmt.Errorf("xurl command not working: %w", err)
	}

	logger.Info("xurl command validation successful")
	return nil
}
