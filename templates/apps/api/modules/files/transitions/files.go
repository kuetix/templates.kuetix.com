package transitions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

type filesTransitions struct {
	workflow.BaseServiceTransition
}

func NewFilesTransitions() interfaces.ServiceTransitions {
	return &filesTransitions{}
}

// ReadFile reads a file's content
func (t *filesTransitions) ReadFile(filePath string) (r domain.FlowStepResult) {
	// Security check - prevent directory traversal
	if err := validatePath(filePath); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("invalid file path: %w", err)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to read file: %w", err)
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"path":    filePath,
		"content": string(content),
	}
	return
}

// WriteFile writes content to a file
func (t *filesTransitions) WriteFile(filePath, content string) (r domain.FlowStepResult) {
	// Security check - prevent directory traversal
	if err := validatePath(filePath); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("invalid file path: %w", err)
		return
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to create directory: %w", err)
		return
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to write file: %w", err)
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"message": "File written successfully",
		"path":    filePath,
	}
	return
}

// validatePath ensures the path is safe and doesn't contain directory traversal
func validatePath(path string) error {
	// Clean the path to resolve any . or .. elements
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if the path is relative to current directory or explicitly allowed paths
	relPath, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	// If relative path starts with .., it's trying to escape
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path attempts to escape working directory")
	}

	return nil
}
