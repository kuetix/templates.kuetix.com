package transitions

import (
	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

type studioTransitions struct {
	workflow.BaseServiceTransition
}

func NewStudioTransitions() interfaces.ServiceTransitions {
	return &studioTransitions{}
}

// GetStudioConfig returns studio configuration
func (t *studioTransitions) GetStudioConfig() (r domain.FlowStepResult) {
	r.Success = true
	r.Response = map[string]interface{}{
		"studioVersion": "1.0.0",
		"apiEnabled":    true,
		"features": []string{
			"workflow-editor",
			"transition-manager",
			"api-explorer",
		},
	}
	return
}

// ValidateStudioAccess validates studio access
func (t *studioTransitions) ValidateStudioAccess(apiKey string) (r domain.FlowStepResult) {
	if apiKey == "" {
		r.Success = false
		r.Response = map[string]interface{}{
			"error": "API key is required",
		}
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"valid":  true,
		"access": "granted",
	}
	return
}

// GetStudioMetadata returns studio metadata
func (t *studioTransitions) GetStudioMetadata(projectId string) (r domain.FlowStepResult) {
	r.Success = true
	r.Response = map[string]interface{}{
		"projectId":   projectId,
		"name":        "Kuetix Studio",
		"description": "Studio API for workflow management",
		"endpoints": []string{
			"/api/workflows",
			"/api/transitions",
			"/api/modules",
		},
	}
	return
}
