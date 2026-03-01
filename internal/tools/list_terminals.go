package tools

import (
	"context"
	"encoding/json"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ListTerminalsSchema returns the JSON Schema for the list_terminals tool.
func ListTerminalsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

// ListTerminals returns a tool handler that lists all active terminal sessions.
func ListTerminals(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		sessions := mgr.List()

		if len(sessions) == 0 {
			return helpers.TextResult("No active terminal sessions"), nil
		}

		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return helpers.ErrorResult("list_error", fmt.Sprintf("marshal sessions: %v", err)), nil
		}

		return helpers.TextResult(string(data)), nil
	}
}
