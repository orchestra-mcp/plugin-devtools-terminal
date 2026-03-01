package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// CreateTerminalSchema returns the JSON Schema for the create_terminal tool.
func CreateTerminalSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"shell": map[string]any{
				"type":        "string",
				"description": "Shell to launch (e.g. /bin/bash, /bin/zsh). Defaults to $SHELL or /bin/sh",
			},
			"cols": map[string]any{
				"type":        "number",
				"description": "Terminal width in columns (default: 80)",
			},
			"rows": map[string]any{
				"type":        "number",
				"description": "Terminal height in rows (default: 24)",
			},
		},
	})
	return s
}

// CreateTerminal returns a tool handler that creates a new PTY terminal session.
func CreateTerminal(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		shell := helpers.GetString(req.Arguments, "shell")
		cols := helpers.GetInt(req.Arguments, "cols")
		rows := helpers.GetInt(req.Arguments, "rows")

		id, err := mgr.Create(shell, cols, rows)
		if err != nil {
			return helpers.ErrorResult("create_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Terminal session created: %s", id)), nil
	}
}
