package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"google.golang.org/protobuf/types/known/structpb"
)

// TerminalStreamSchema returns the JSON Schema for the terminal_stream streaming tool.
func TerminalStreamSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"terminal_id": map[string]any{
				"type":        "string",
				"description": "ID of the terminal session to stream output from",
			},
		},
		"required": []any{"terminal_id"},
	})
	return s
}

// TerminalStream returns a streaming tool handler that subscribes to a terminal
// session's PTY output and pushes each chunk through the streaming channel.
// The stream runs until the context is cancelled or the terminal session closes.
func TerminalStream(mgr *pty.Manager) func(ctx context.Context, req *pluginv1.StreamStart, chunks chan<- []byte) error {
	return func(ctx context.Context, req *pluginv1.StreamStart, chunks chan<- []byte) error {
		terminalID := ""
		if req.Arguments != nil {
			if v, ok := req.Arguments.GetFields()["terminal_id"]; ok {
				terminalID = v.GetStringValue()
			}
		}
		if terminalID == "" {
			return fmt.Errorf("missing required parameter: terminal_id")
		}

		ch, unsub, err := mgr.Subscribe(terminalID)
		if err != nil {
			return err
		}
		defer unsub()

		for {
			select {
			case data, ok := <-ch:
				if !ok {
					// Terminal session closed — subscriber channel was closed.
					return nil
				}
				select {
				case chunks <- data:
				case <-ctx.Done():
					return ctx.Err()
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
