package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const (
	serverName         = "trading-cli"
	protocolVersion    = "2025-11-25"
	toolTimeout        = 60 * time.Second
	maxConcurrentTools = 4
)

var serverVersion = "dev"

type Server struct {
	tools    []ToolDef
	handlers map[string]ToolHandler

	notifyWriter io.Writer
	notifyMu     sync.Mutex
	toolSem      chan struct{}
}

func NewServer() *Server {
	s := &Server{
		handlers: make(map[string]ToolHandler),
		toolSem:  make(chan struct{}, maxConcurrentTools),
	}
	registerTools(s)
	return s
}

func (s *Server) HandleRequest(req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return nil
	case "notifications/cancelled":
		return nil
	case "ping":
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "logging/setLevel":
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "prompts/list":
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: PromptsListResult{Prompts: nil}}
	case "prompts/get":
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: PromptsGetResult{Messages: nil}}
	case "resources/list":
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: ResourcesListResult{Resources: nil}}
	case "resources/read":
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: ResourcesReadResult{Contents: nil}}
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)},
		}
	}
}

func (s *Server) handleInitialize(req *Request) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: InitializeResult{
			ProtocolVersion: protocolVersion,
			Capabilities: Capabilities{
				Tools:     &ToolsCapability{ListChanged: false},
				Prompts:   &PromptsCapability{ListChanged: false},
				Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
				Logging:   &LoggingCapability{},
			},
			ServerInfo: ServerInfo{
				Name:    serverName,
				Version: serverVersion,
			},
		},
	}
}

func (s *Server) handleToolsList(req *Request) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ToolsListResult{Tools: s.tools},
	}
}

func (s *Server) handleToolsCall(req *Request) *Response {
	var params ToolCallParams
	raw, err := json.Marshal(req.Params)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: ErrInvalidParams, Message: fmt.Sprintf("invalid params: %v", err)},
		}
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32602, Message: fmt.Sprintf("invalid params: %v", err)},
		}
	}

	handler, ok := s.handlers[params.Name]
	if !ok {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", params.Name)},
		}
	}

	s.SendLog("info", fmt.Sprintf("Calling tool: %s", params.Name))

	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
	defer cancel()

	content, structured, err := handler(ctx, params.Arguments)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ToolCallResult{
				Content: []ContentBlock{{Type: "text", Text: err.Error()}},
				IsError: true,
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolCallResult{
			Content:           content,
			StructuredContent: structured,
		},
	}
}

func (s *Server) SendNotification(method string, params interface{}) error {
	s.notifyMu.Lock()
	defer s.notifyMu.Unlock()
	if s.notifyWriter == nil {
		return nil
	}
	notif := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	return writeJSON(s.notifyWriter, notif)
}

func (s *Server) SendLog(level, message string) {
	_ = s.SendNotification("notifications/message", LogParams{
		Level: level, Logger: "trading-cli", Data: message,
	})
}

func (s *Server) ServeStdio(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	s.notifyWriter = out

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			resp := Response{
				JSONRPC: "2.0",
				Error:   &Error{Code: -32700, Message: fmt.Sprintf("parse error: %v", err)},
			}
			if writeErr := writeJSON(out, resp); writeErr != nil {
				return writeErr
			}
			continue
		}

		resp := s.HandleRequest(&req)
		if resp == nil {
			continue
		}
		if err := writeJSON(out, resp); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	return nil
}

func writeJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	if _, err := fmt.Fprintf(w, "%s\n", data); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

func Run() error {
	s := NewServer()
	log.SetOutput(io.Discard)
	return s.ServeStdio(os.Stdin, os.Stdout)
}
