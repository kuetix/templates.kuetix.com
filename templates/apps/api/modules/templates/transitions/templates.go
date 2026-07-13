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

type templatesTransitions struct {
	workflow.BaseServiceTransition
}

func NewTemplatesTransitions() interfaces.ServiceTransitions {
	return &templatesTransitions{}
}

// ProjectTemplate represents a template project structure
type ProjectTemplate struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Files       []TemplateFile  `json:"files"`
	Directories []string        `json:"directories"`
	Metadata    ProjectMetadata `json:"metadata"`
}

// TemplateFile represents a file in a template
type TemplateFile struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	IsTemplate  bool   `json:"isTemplate"`
	Description string `json:"description"`
}

// ProjectMetadata contains metadata about a project template
type ProjectMetadata struct {
	Author     string   `json:"author"`
	Version    string   `json:"version"`
	Tags       []string `json:"tags"`
	License    string   `json:"license"`
	Repository string   `json:"repository"`
}

// ListTemplates returns all available project templates
func (t *templatesTransitions) ListTemplates() (r domain.FlowStepResult) {
	templates := []ProjectTemplate{
		{
			Name:        "basic-workflow",
			Description: "Basic workflow project with common transitions",
			Type:        "workflow",
			Directories: []string{"workflows", "modules", "runtime"},
			Files: []TemplateFile{
				{
					Path:        "workflows/main.wsl",
					IsTemplate:  true,
					Description: "Main workflow file",
					Content:     getBasicWorkflowTemplate(),
				},
				{
					Path:        "go.mod",
					IsTemplate:  true,
					Description: "Go module file",
					Content:     getGoModTemplate(),
				},
			},
			Metadata: ProjectMetadata{
				Author:  "Kuetix",
				Version: "1.0.0",
				Tags:    []string{"workflow", "basic", "starter"},
				License: "MIT",
			},
		},
		{
			Name:        "feature-project",
			Description: "Feature-based project with test workflows",
			Type:        "feature",
			Directories: []string{"workflows", "modules", "tests"},
			Files: []TemplateFile{
				{
					Path:        "workflows/feature.wsl",
					IsTemplate:  true,
					Description: "Feature workflow file",
					Content:     getFeatureWorkflowTemplate(),
				},
			},
			Metadata: ProjectMetadata{
				Author:  "Kuetix",
				Version: "1.0.0",
				Tags:    []string{"feature", "testing"},
				License: "MIT",
			},
		},
		{
			Name:        "solution-project",
			Description: "Solution-based project with orchestration",
			Type:        "solution",
			Directories: []string{"workflows", "modules"},
			Files: []TemplateFile{
				{
					Path:        "workflows/solution.wsl",
					IsTemplate:  true,
					Description: "Solution workflow file",
					Content:     getSolutionWorkflowTemplate(),
				},
			},
			Metadata: ProjectMetadata{
				Author:  "Kuetix",
				Version: "1.0.0",
				Tags:    []string{"solution", "orchestration"},
				License: "MIT",
			},
		},
	}

	r.Success = true
	r.Response = templates
	return
}

// CreateProject creates a new project from a template
func (t *templatesTransitions) CreateProject(templateName, projectName, outputPath string, variables map[string]interface{}) (r domain.FlowStepResult) {
	// Get templates
	templatesResult := t.ListTemplates()
	if !templatesResult.Success {
		r.Success = false
		r.Error = templatesResult.Error
		return
	}

	templates := templatesResult.Response.([]ProjectTemplate)
	var selectedTemplate *ProjectTemplate
	for i := range templates {
		if templates[i].Name == templateName {
			selectedTemplate = &templates[i]
			break
		}
	}

	if selectedTemplate == nil {
		r.Success = false
		r.Error = fmt.Errorf("template not found: %s", templateName)
		return
	}

	// Determine output path
	if outputPath == "" {
		outputPath = "./projects"
	}

	projectPath := filepath.Join(outputPath, projectName)

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to create project directory: %w", err)
		return
	}

	// Create directories
	for _, dir := range selectedTemplate.Directories {
		dirPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			r.Success = false
			r.Error = fmt.Errorf("failed to create directory %s: %w", dir, err)
			return
		}
	}

	// Create files
	for _, file := range selectedTemplate.Files {
		filePath := filepath.Join(projectPath, file.Path)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			r.Success = false
			r.Error = fmt.Errorf("failed to create parent directory for %s: %w", file.Path, err)
			return
		}

		// Process template variables
		content := file.Content
		if file.IsTemplate {
			content = processTemplate(content, variables, projectName)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			r.Success = false
			r.Error = fmt.Errorf("failed to write file %s: %w", file.Path, err)
			return
		}
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"message":     "Project created successfully",
		"projectPath": projectPath,
		"projectName": projectName,
	}
	return
}

// GetProjectTree returns the file tree of a project
func (t *templatesTransitions) GetProjectTree(projectPath string) (r domain.FlowStepResult) {
	tree, err := buildFileTree(projectPath)
	if err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to build file tree: %w", err)
		return
	}

	r.Success = true
	r.Response = tree
	return
}

// FileNode represents a node in the file tree
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

// buildFileTree builds a tree structure of files and directories
func buildFileTree(rootPath string) (FileNode, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return FileNode{}, err
	}

	node := FileNode{
		Name:  filepath.Base(rootPath),
		Path:  rootPath,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return node, nil
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return node, err
	}

	for _, entry := range entries {
		// Skip hidden files and certain directories
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" {
			continue
		}

		childPath := filepath.Join(rootPath, entry.Name())
		childNode, err := buildFileTree(childPath)
		if err != nil {
			continue // Skip files that can't be read
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// processTemplate replaces template variables
func processTemplate(content string, variables map[string]interface{}, projectName string) string {
	// Replace project name
	content = strings.ReplaceAll(content, "{{PROJECT_NAME}}", projectName)

	// Replace custom variables
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", strings.ToUpper(key))
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}

	return content
}

// Template content functions
func getBasicWorkflowTemplate() string {
	return `module {{PROJECT_NAME}}

import services/common

const {
    description: "{{PROJECT_NAME}} workflow",
    version: "1.0.0",
    enabled: true
}

workflow main {
  start: Initialize

  state Initialize {
    action services/common/response.Response(value: "{{PROJECT_NAME}} initialized", statusCode: 200) as Result
    on success -> Complete
  }

  state Complete {
    action services/common/response.Response(value: "Workflow completed", statusCode: 200) as FinalResult
    end ok
  }
}
`
}

func getFeatureWorkflowTemplate() string {
	return `module {{PROJECT_NAME}}

import services/common

const {
    description: "{{PROJECT_NAME}} feature",
    version: "1.0.0"
}

feature {{PROJECT_NAME}}_feature {
  start: Process

  state Process(input) {
    action services/common/response.Response(value: $input, statusCode: 200) as Result
    on success -> Complete
  }

  state Complete {
    end ok
  }
}
`
}

func getSolutionWorkflowTemplate() string {
	return `module {{PROJECT_NAME}}

import services/common

const {
    description: "{{PROJECT_NAME}} solution",
    version: "1.0.0"
}

solution {{PROJECT_NAME}}_solution {
  start: Execute

  state Execute {
    action services/common/response.Response(value: "Solution executing", statusCode: 200) as Result
    end ok
  }
}
`
}

func getGoModTemplate() string {
	return `module {{PROJECT_NAME}}

go 1.25.1

require github.com/kuetix/engine v0.1.4
`
}
