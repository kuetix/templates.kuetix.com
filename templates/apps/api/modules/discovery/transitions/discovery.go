package transitions

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

type discoveryTransitions struct {
	workflow.BaseServiceTransition
}

func NewDiscoveryTransitions() interfaces.ServiceTransitions {
	return &discoveryTransitions{}
}

// TransitionInfo represents information about a transition
type TransitionInfo struct {
	Module      string         `json:"module"`
	Name        string         `json:"name"`
	Path        string         `json:"path"`
	Functions   []FunctionInfo `json:"functions"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
}

// FunctionInfo represents a function in a transition
type FunctionInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters"`
}

// WorkflowInfo represents information about a workflow
type WorkflowInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Module      string   `json:"module"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Type        string   `json:"type"` // workflow, feature, solution
	Imports     []string `json:"imports"`
}

// ListTransitions scans and returns all available transitions using modules.json
func (t *discoveryTransitions) ListTransitions(modulesJsonPath string) (r domain.FlowStepResult) {
	data, err := os.ReadFile(modulesJsonPath)
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to read modules.json: %w", err)
		return
	}

	var modules map[string]struct {
		Info struct {
			Namespace   string `json:"namespace"`
			Class       string `json:"class"`
			Label       string `json:"label"`
			Description string `json:"description"`
		} `json:"info"`
		Methods []struct {
			Value       string `json:"value"`
			Label       string `json:"label"`
			Description string `json:"description"`
		} `json:"methods"`
	}

	if err := json.Unmarshal(data, &modules); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to parse modules.json: %w", err)
		return
	}

	transitions := []TransitionInfo{}
	for _, m := range modules {
		functions := []FunctionInfo{}
		seenMethods := make(map[string]bool)

		for _, method := range m.Methods {
			if seenMethods[method.Value] {
				continue
			}
			seenMethods[method.Value] = true

			// Extract parameters from label like "ChargeCard(amount, retries, accountId)"
			params := []string{}
			if strings.Contains(method.Label, "(") && strings.Contains(method.Label, ")") {
				paramStr := method.Label[strings.Index(method.Label, "(")+1 : strings.LastIndex(method.Label, ")")]
				if paramStr != "" {
					pParts := strings.Split(paramStr, ",")
					for _, p := range pParts {
						params = append(params, strings.TrimSpace(p))
					}
				}
			}

			functions = append(functions, FunctionInfo{
				Name:        method.Value,
				Description: method.Description,
				Parameters:  params,
			})
		}

		// Reconstruct relPath for the transition file
		// e.g. "billing/payment/payment" -> "billing/payment/transitions/payment.go"
		relPath := filepath.Join(m.Info.Namespace, "transitions", m.Info.Class+".go")

		transitions = append(transitions, TransitionInfo{
			Module:      m.Info.Namespace,
			Name:        m.Info.Class,
			Path:        relPath,
			Category:    extractCategory(strings.Split(m.Info.Namespace, "/")),
			Description: m.Info.Description,
			Functions:   functions,
		})
	}

	r.Success = true
	r.Response = transitions
	return
}

// ListWorkflows scans and returns all available workflows
func (t *discoveryTransitions) ListWorkflows(workflowsPath string, filterType string) (r domain.FlowStepResult) {
	workflows := []WorkflowInfo{}

	err := filepath.WalkDir(workflowsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".wsl") {
			return nil
		}

		relPath, _ := filepath.Rel(workflowsPath, path)
		dir := filepath.Dir(relPath)
		name := strings.TrimSuffix(d.Name(), ".wsl")

		workflowType := detectWorkflowType(path)

		// Apply filter if specified
		if filterType != "" && workflowType != filterType {
			return nil
		}

		workflowInfo := WorkflowInfo{
			Name:        name,
			Path:        relPath,
			Module:      dir,
			Type:        workflowType,
			Description: "Workflow: " + name,
			Imports:     extractImports(path),
		}

		workflows = append(workflows, workflowInfo)

		return nil
	})

	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to scan workflows: %w", err)
		return
	}

	r.Success = true
	r.Response = workflows
	return
}

// extractCategory extracts the category from path parts
func extractCategory(parts []string) string {
	if len(parts) >= 1 {
		return parts[0]
	}
	return "general"
}

// detectWorkflowType detects if a workflow is a workflow, feature, or solution
func detectWorkflowType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "workflow"
	}

	text := string(content)
	if strings.Contains(text, "feature ") {
		return "feature"
	}
	if strings.Contains(text, "solution ") {
		return "solution"
	}
	return "workflow"
}

// extractImports extracts import statements from a workflow file
func extractImports(filePath string) []string {
	imports := []string{}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return imports
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			importPath := strings.TrimPrefix(line, "import ")
			importPath = strings.TrimSpace(importPath)
			imports = append(imports, importPath)
		}
	}

	return imports
}
