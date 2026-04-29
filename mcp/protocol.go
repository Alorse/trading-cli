package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// JSON-RPC 2.0 core types
// ---------------------------------------------------------------------------

// Request is a generic JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response (success or error).
type Response struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      any     `json:"id,omitempty"`
	Result  any     `json:"result,omitempty"`
	Error   *Error  `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
)

// ---------------------------------------------------------------------------
// Initialize
// ---------------------------------------------------------------------------

// InitializeParams is sent by the client during the initialize handshake.
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// InitializeResult is returned by the server after a successful initialize.
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

// ClientInfo describes the client application.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo describes the server application.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapabilities declares what the client supports.
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability indicates the client can provide roots.
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability indicates the client supports sampling (LLM requests).
type SamplingCapability struct{}

// Capabilities declares what the server supports.
type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

// ToolsCapability indicates the server exposes tools.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability indicates the server exposes prompts.
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability indicates the server exposes resources.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability indicates the server supports logging.
type LoggingCapability struct{}

// ---------------------------------------------------------------------------
// Tools
// ---------------------------------------------------------------------------

// ToolsListResult is returned when listing available tools.
type ToolsListResult struct {
	Tools []ToolDef `json:"tools"`
}

// ToolDef describes a tool exposed by the server.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema InputSchema     `json:"inputSchema"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// InputSchema is a JSON Schema object describing tool parameters.
type InputSchema struct {
	Type       string            `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
}

// Property describes a single property in a tool's input schema.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// ToolAnnotations provides hints about tool behavior.
type ToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`
	DestructiveHint bool   `json:"destructiveHint,omitempty"`
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`
	OpenWorldHint   bool   `json:"openWorldHint,omitempty"`
}

// ToolCallParams is sent by the client to invoke a tool.
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolCallResult is returned after tool execution.
type ToolCallResult struct {
	Content           []ContentBlock `json:"content"`
	StructuredContent interface{}    `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError,omitempty"`
}

// ---------------------------------------------------------------------------
// Content
// ---------------------------------------------------------------------------

// ContentBlock represents a piece of content (text, image, or resource).
type ContentBlock struct {
	Type        string             `json:"type"`
	Text        string             `json:"text,omitempty"`
	Data        string             `json:"data,omitempty"`
	MimeType    string             `json:"mimeType,omitempty"`
	Resource    *ResourceContent   `json:"resource,omitempty"`
	Annotations *ContentAnnotation `json:"annotations,omitempty"`
}

// ContentAnnotation provides metadata about a content block.
type ContentAnnotation struct {
	Audience []string `json:"audience,omitempty"`
	Priority float64  `json:"priority,omitempty"`
}

// ---------------------------------------------------------------------------
// Prompts
// ---------------------------------------------------------------------------

// PromptsListResult is returned when listing available prompts.
type PromptsListResult struct {
	Prompts []PromptDef `json:"prompts"`
}

// PromptDef describes a prompt template exposed by the server.
type PromptDef struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument describes a single argument a prompt accepts.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptsGetParams is sent by the client to retrieve a prompt.
type PromptsGetParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

// PromptsGetResult is returned after fetching a prompt.
type PromptsGetResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage is a single message within a prompt.
type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// ---------------------------------------------------------------------------
// Resources
// ---------------------------------------------------------------------------

// ResourcesListResult is returned when listing available resources.
type ResourcesListResult struct {
	Resources []ResourceDef `json:"resources"`
}

// ResourceDef describes a resource exposed by the server.
type ResourceDef struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourcesReadParams is sent by the client to read a resource.
type ResourcesReadParams struct {
	URI string `json:"uri"`
}

// ResourcesReadResult is returned after reading a resource.
type ResourcesReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent holds the actual content of a resource.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

// ProgressParams is sent to report progress on a long-running operation.
type ProgressParams struct {
	Progress      float64 `json:"progress"`
	Total         float64 `json:"total,omitempty"`
	ProgressToken any     `json:"progressToken"`
}

// LogParams is sent to deliver a log message to the client.
type LogParams struct {
	Level  string `json:"level"`
	Logger string `json:"logger,omitempty"`
	Data   any    `json:"data"`
}

// Log levels.
const (
	LogLevelDebug     = "debug"
	LogLevelInfo      = "info"
	LogLevelNotice    = "notice"
	LogLevelWarning   = "warning"
	LogLevelError     = "error"
	LogLevelCritical  = "critical"
	LogLevelAlert     = "alert"
	LogLevelEmergency = "emergency"
)

// ---------------------------------------------------------------------------
// Handler / callback types
// ---------------------------------------------------------------------------

// ToolHandler is the function signature for handling tool calls.
// Returns content blocks, an optional structured result, and an error.
type ToolHandler func(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error)

// ElicitFunc is an optional callback for eliciting information from the user.
// May be nil if the server does not need this capability.
type ElicitFunc func(ctx context.Context, message string, schema map[string]any) (map[string]any, error)

// SamplingFunc is an optional callback for making LLM sampling requests via the client.
// May be nil if the server does not need this capability.
type SamplingFunc func(ctx context.Context, params any) (any, error)

// ProgressFunc is an optional callback for reporting progress to the client.
// May be nil if the server does not need this capability.
type ProgressFunc func(ctx context.Context, token any, progress float64, total float64)
