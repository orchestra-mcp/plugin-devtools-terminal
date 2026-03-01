package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// GetOutputSchema returns the JSON Schema for the get_output tool.
func GetOutputSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"terminal_id": map[string]any{
				"type":        "string",
				"description": "ID of the terminal session",
			},
		},
		"required": []any{"terminal_id"},
	})
	return s
}

// GetOutput returns a tool handler that retrieves accumulated output from a terminal session.
func GetOutput(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "terminal_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		terminalID := helpers.GetString(req.Arguments, "terminal_id")

		output, err := mgr.GetOutput(terminalID)
		if err != nil {
			return helpers.ErrorResult("output_error", err.Error()), nil
		}

		return helpers.TextResult(output), nil
	}
}
