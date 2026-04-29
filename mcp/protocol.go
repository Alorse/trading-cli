package mcp

import (
	"context"
	"encoding/json"
)

// JSON-RPC 2.0 types

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP protocol types

type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ClientCapabilities struct {
	Roots *RootsCapability `json:"roots,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe"`
	ListChanged bool `json:"listChanged"`
}

type LoggingCapability struct{}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tools types

type ToolsListResult struct {
	Tools []ToolDef `json:"tools"`
}

type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema InputSchema     `json:"inputSchema"`
	Annotations *ToolAnnotation `json:"annotations,omitempty"`
}

type ToolAnnotation struct {
	ReadOnlyHint bool `json:"readOnlyHint,omitempty"`
}

type InputSchema struct {
	Type       string         `json:"type"`
	Properties map[string]Prop `json:"properties"`
	Required   []string       `json:"required,omitempty"`
}

type Prop struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type ToolCallResult struct {
	Content           []ContentBlock `json:"content"`
	StructuredContent interface{}    `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type        string             `json:"type"`
	Text        string             `json:"text,omitempty"`
	Annotations *ContentAnnotation `json:"annotations,omitempty"`
}

type ContentAnnotation struct {
	Audience []string `json:"audience,omitempty"`
	Priority float64  `json:"priority,omitempty"`
}

// ToolHandler processes a tool call and returns content blocks, structured content, and an error.
type ToolHandler func(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error)

// Prompts types

type PromptsListResult struct {
	Prompts []PromptDef `json:"prompts"`
}

type PromptDef struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type PromptsGetParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type PromptsGetResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// Resources types

type ResourcesListResult struct {
	Resources []ResourceDef `json:"resources"`
}

type ResourceDef struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourcesReadParams struct {
	URI string `json:"uri"`
}

type ResourcesReadResult struct {
	Contents []ResourceContent `json:"contents"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

// Notification types

type ProgressParams struct {
	ProgressToken string  `json:"progressToken"`
	Progress      float64 `json:"progress"`
	Total         float64 `json:"total"`
	Message       string  `json:"message,omitempty"`
}

type LogParams struct {
	Level  string `json:"level"`
	Logger string `json:"logger"`
	Data   string `json:"data"`
}

