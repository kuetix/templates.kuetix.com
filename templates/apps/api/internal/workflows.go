package internal

// FunctionMetadata mirrors the metadata block emitted by the CLI for each
// referenced action. It is resolved at upload time from the runtime registry.
type FunctionMetadata struct {
	GoModule    string   `json:"go_module"`
	ModulePath  string   `json:"module_path"`
	FilePath    string   `json:"file_path"`
	Namespace   string   `json:"namespace"`
	Class       string   `json:"class"`
	Name        string   `json:"name"`
	NumIn       int      `json:"num_in"`
	NumOut      int      `json:"num_out"`
	ArgTypes    []string `json:"arg_types"`
	ReturnTypes []string `json:"return_types"`
	ArgNames    []string `json:"arg_names"`
	ReturnNames []string `json:"return_names"`
}

// WorkflowArg is one argument captured from a workflow action call.
type WorkflowArg struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
	Raw   string `json:"raw"`
}

// WorkflowActionInfo describes one action invocation inside a workflow state.
type WorkflowActionInfo struct {
	Workflow string            `json:"workflow"`
	State    string            `json:"state"`
	Module   string            `json:"module,omitempty"`
	Name     string            `json:"name"`
	As       string            `json:"as,omitempty"`
	Args     []WorkflowArg     `json:"args,omitempty"`
	Params   []string          `json:"params,omitempty"`
	Terminal string            `json:"terminal,omitempty"`
	Metadata *FunctionMetadata `json:"metadata,omitempty"`
}

// WorkflowDependency describes one (go_module, namespace/class) tuple
// referenced by any action in the workflow.
type WorkflowDependency struct {
	GoModule        string                 `json:"go_module"`
	ModulePath      string                 `json:"module_path"`
	Namespace       string                 `json:"namespace"`
	Class           string                 `json:"class"`
	ModuleInfo      map[string]interface{} `json:"module_info,omitempty"`
	ModulesJSONPath string                 `json:"modules_json_path,omitempty"`
}

// WorkflowPayload is the canonical body for create/replace requests.
// Field names are authoritative — they come from the CLI's struct tags.
type WorkflowPayload struct {
	Name         string               `json:"name"`
	Project      string               `json:"project,omitempty"`
	FilePath     string               `json:"file_path,omitempty"`
	Content      string               `json:"content"`
	Imports      map[string]string    `json:"imports,omitempty"`
	Actions      []WorkflowActionInfo `json:"actions"`
	Dependencies []WorkflowDependency `json:"dependencies"`
	Public       bool                 `json:"public,omitempty"`
	Published    bool                 `json:"published,omitempty"`
}

// WorkflowRecord is the durable, server-side representation of a workflow.
// It embeds the payload and adds owner/version/timestamp bookkeeping.
type WorkflowRecord struct {
	Uri      string `json:"uri"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Project  string `json:"project,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	Content  string `json:"content"`

	Imports      map[string]string    `json:"imports,omitempty"`
	Actions      []WorkflowActionInfo `json:"actions"`
	Dependencies []WorkflowDependency `json:"dependencies"`

	Visibility string `json:"visibility"`
	Published  bool   `json:"published"`
	Public     bool   `json:"public"`

	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// WorkflowSummary is the per-item shape returned by list endpoints — the full
// payload is intentionally omitted to keep list responses cheap.
type WorkflowSummary struct {
	Name              string `json:"name"`
	Version           int    `json:"version"`
	UpdatedAt         string `json:"updated_at"`
	ActionsCount      int    `json:"actions_count"`
	DependenciesCount int    `json:"dependencies_count"`
}
