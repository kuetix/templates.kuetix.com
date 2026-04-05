# Kuetix Engine Skeleton Templates

This directory contains all templates used by the `kue init` command to generate new applications and packages.

## Directory Structure

```
templates/
├── apps/               # Application type templates
│   ├── api/           # API server template
│   ├── cli/           # CLI application template
│   ├── consumer/      # AMQP consumer template
│   └── service/       # Background service template
├── config/            # Configuration file templates
│   ├── Dockerfile.tmpl
│   ├── Makefile.tmpl
│   ├── README-app.md.tmpl
│   ├── README-package.md.tmpl
│   ├── docker-compose.yml.tmpl
│   ├── gitignore.tmpl
│   ├── go.mod.tmpl
│   ├── kuetix.json.tmpl
│   └── runner.sh.tmpl
├── modules/           # Module templates
│   └── transitions/   # Transition module templates
│       └── example.go.tmpl
└── workflows/         # Workflow templates
    ├── common/        # Basic workflow examples
    ├── features/      # Feature workflow examples
    └── solutions/     # Solution workflow examples
```

## Template Variables

All templates use Go's `text/template` syntax and have access to the following variables:

- `{{.ProjectName}}` - The name of the project being generated
- `{{.AppType}}` - The type of application (cli, api, consumer, service, package, all)
- `{{.KuetixEngineVersion}}` - The version of kuetix/engine to use (e.g., "v0.1.4")
- `{{.MinGoVersion}}` - The minimum Go version required (e.g., "1.21")

## Generated Files

When initializing a new project, the following files are generated:

### For All App Types
- **Configuration files**: go.mod, Makefile, .gitignore, README.md, Dockerfile, docker-compose.yml
- **Runner script**: runner.sh for easy development workflow
- **Example transition**: modules/services/example/transitions/example.go

### For Package Type
- **Workflows**: Example workflow, feature, and solution files
- **Package metadata**: kuetix.json for package management

### For Specific App Types
- **CLI**: Command-line application in cmd/cli/main.go
- **API**: HTTP API server in cmd/api/main.go
- **Consumer**: AMQP consumer in cmd/consumer/main.go
- **Service**: Background service in cmd/service/main.go

## Transition Templates

The `modules/transitions/` directory contains templates for generating custom transition modules. Transitions are the core building blocks for implementing custom workflow logic.

### Example Transition Structure

Generated transitions follow this pattern:
```go
package transitions

import (
	"github.com/kuetix/engine/pkg/domain"
	"github.com/kuetix/engine/pkg/domain/interfaces"
	"github.com/kuetix/engine/pkg/workflow"
)

type exampleTransitions struct {
	workflow.BaseServiceTransition
}

func NewExampleTransitions() interfaces.ServiceTransitions {
	return &exampleTransitions{}
}

func (t *exampleTransitions) Execute(message string) (r domain.FlowStepResult) {
	// Your transition logic here
	r.Success = true
	r.Response = map[string]interface{}{
		"message": message,
		"status":  "executed",
	}
	return
}
```

### Using Generated Transitions

After generation, you can:
1. Modify the example transition methods or add new ones
2. Reference them in your workflows: `action example/Execute`
3. Run `generate cache` or `kue run` to register the transitions

## Example Usage

When a user runs:
```bash
kue init --name myapp --app-type cli
```

The template system will:
1. Load the appropriate templates from this directory
2. Replace `{{.ProjectName}}` with "myapp"
3. Replace `{{.AppType}}` with "cli"
4. Generate the files in the output directory

## Customizing Templates

To customize templates:

1. Edit the `.tmpl` files directly
2. Use Go template syntax for any dynamic content
3. Test your changes by running `kue init` with different options
4. Rebuild the kue binary: `make kue` or `go build ./cmd/kue/`

## Adding New Templates

To add a new template:

1. Create your template file in the appropriate subdirectory
2. Use the `.tmpl` extension
3. Add template variable placeholders where needed
4. Update `cmd/kue/generators.go` to use your new template
5. Add the template rendering call in the appropriate generator function

Example:
```go
func generateMyNewFile(projectPath, name string) error {
    data := TemplateData{
        ProjectName:         name,
        KuetixEngineVersion: KuetixEngineVersion,
        MinGoVersion:        MinGoVersion,
    }
    
    return writeTemplateToFile(
        "templates/path/to/myfile.tmpl",
        filepath.Join(projectPath, "path/to/output"),
        data,
        0644,
    )
}
```

## Template Embedding

Templates are embedded directly into the `kue` binary using Go's `embed` package. This means:
- No external template files needed at runtime
- Templates are always available with the binary
- Changes to templates require rebuilding the binary

The embedding is done in `cmd/kue/template_loader.go`:
```go
//go:embed templates
var templatesFS embed.FS
```

## Best Practices

1. **Keep templates simple** - Avoid complex logic in templates
2. **Use meaningful variable names** - Make it clear what each placeholder represents
3. **Document template variables** - Add comments for any non-obvious placeholders
4. **Test thoroughly** - Test each template with different project names and types
5. **Follow Go conventions** - Generated Go code should follow standard formatting
6. **Version compatibility** - Ensure templates work with the specified engine version

## Troubleshooting

If templates fail to render:
- Check template syntax with Go's text/template package
- Verify all placeholders have corresponding data fields
- Ensure file paths in generators.go match template locations
- Rebuild the binary after template changes

For more information, see the main project documentation.
