package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// CloseTerminalSchema returns the JSON Schema for the close_terminal tool.
func CloseTerminalSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"terminal_id": map[string]any{
				"type":        "string",
				"description": "ID of the terminal session to close",
			},
		},
		"required": []any{"terminal_id"},
	})
	return s
}

// CloseTerminal returns a tool handler that closes a terminal session.
func CloseTerminal(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "terminal_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		terminalID := helpers.GetString(req.Arguments, "terminal_id")

		if err := mgr.Close(terminalID); err != nil {
			return helpers.ErrorResult("close_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Terminal session %s closed", terminalID)), nil
	}
}
