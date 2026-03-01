package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ResizeTerminalSchema returns the JSON Schema for the resize_terminal tool.
func ResizeTerminalSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"terminal_id": map[string]any{
				"type":        "string",
				"description": "ID of the terminal session",
			},
			"cols": map[string]any{
				"type":        "number",
				"description": "New terminal width in columns",
			},
			"rows": map[string]any{
				"type":        "number",
				"description": "New terminal height in rows",
			},
		},
		"required": []any{"terminal_id", "cols", "rows"},
	})
	return s
}

// ResizeTerminal returns a tool handler that resizes a terminal session.
func ResizeTerminal(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "terminal_id", "cols", "rows"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		terminalID := helpers.GetString(req.Arguments, "terminal_id")
		cols := helpers.GetInt(req.Arguments, "cols")
		rows := helpers.GetInt(req.Arguments, "rows")

		if err := mgr.Resize(terminalID, cols, rows); err != nil {
			return helpers.ErrorResult("resize_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Terminal %s resized to %dx%d", terminalID, cols, rows)), nil
	}
}
