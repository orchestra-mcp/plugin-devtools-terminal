package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SendInputSchema returns the JSON Schema for the send_input tool.
func SendInputSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"terminal_id": map[string]any{
				"type":        "string",
				"description": "ID of the terminal session",
			},
			"input": map[string]any{
				"type":        "string",
				"description": "Input string to send to the terminal",
			},
		},
		"required": []any{"terminal_id", "input"},
	})
	return s
}

// SendInput returns a tool handler that sends input to a terminal session.
func SendInput(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "terminal_id", "input"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		terminalID := helpers.GetString(req.Arguments, "terminal_id")
		input := helpers.GetString(req.Arguments, "input")

		if err := mgr.SendInput(terminalID, input); err != nil {
			return helpers.ErrorResult("send_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Input sent to terminal %s", terminalID)), nil
	}
}
