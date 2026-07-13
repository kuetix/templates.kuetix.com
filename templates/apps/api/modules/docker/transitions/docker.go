package transitions

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

type dockerTransitions struct {
	workflow.BaseServiceTransition
}

func NewDockerTransitions() interfaces.ServiceTransitions {
	return &dockerTransitions{}
}

// BuildImage builds a Docker image from a project
func (t *dockerTransitions) BuildImage(projectPath, imageName, imageTag string, buildArgs map[string]string) (r domain.FlowStepResult) {
	if imageTag == "" {
		imageTag = "latest"
	}

	imageFullTag := fmt.Sprintf("%s:%s", imageName, imageTag)

	// Check if Dockerfile exists
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		// Create default Dockerfile if it doesn't exist
		if err := createDefaultDockerfile(dockerfilePath); err != nil {
			r.Success = false
			r.Error = fmt.Errorf("failed to create Dockerfile: %w", err)
			return
		}
	}

	// Prepare docker build command
	args := []string{"build", "-t", imageFullTag}

	// Add build args
	for key, value := range buildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, projectPath)

	// Execute docker build
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("docker build failed: %s\nOutput: %s", err.Error(), stderr.String())
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"message":  "Docker image built successfully",
		"imageTag": imageFullTag,
		"output":   stdout.String(),
	}
	return
}

// RunWorkflow runs a workflow in a Docker container
func (t *dockerTransitions) RunWorkflow(imageName, imageTag, workflowPath string, environment map[string]string) (r domain.FlowStepResult) {
	if imageTag == "" {
		imageTag = "latest"
	}

	imageFullTag := fmt.Sprintf("%s:%s", imageName, imageTag)

	// Prepare docker run command
	args := []string{"run", "--rm"}

	// Add environment variables
	for key, value := range environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add image and workflow path
	args = append(args, imageFullTag, "./app", "--workflow", workflowPath)

	// Execute docker run
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("docker run failed: %s\nOutput: %s", err.Error(), stderr.String())
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"message":  "Workflow executed successfully",
		"workflow": workflowPath,
		"output":   stdout.String(),
	}
	return
}

// ListImages lists Docker images
func (t *dockerTransitions) ListImages() (r domain.FlowStepResult) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.ID}}\t{{.Size}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to list docker images: %w", err)
		return
	}

	images := []map[string]string{}
	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			images = append(images, map[string]string{
				"image": parts[0],
				"id":    parts[1],
				"size":  parts[2],
			})
		}
	}

	r.Success = true
	r.Response = images
	return
}

// createDefaultDockerfile creates a default Dockerfile
func createDefaultDockerfile(path string) error {
	dockerfile := `FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download || true
RUN CGO_ENABLED=0 go build -o /app/bin/app ./cmd/cli || echo "No cmd/cli, creating runtime only"

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/app . 2>/dev/null || true
COPY --from=builder /app/workflows ./workflows 2>/dev/null || true
COPY --from=builder /app/modules ./modules 2>/dev/null || true
COPY --from=builder /app/runtime ./runtime 2>/dev/null || true

CMD ["./app"]
`
	return os.WriteFile(path, []byte(dockerfile), 0644)
}
