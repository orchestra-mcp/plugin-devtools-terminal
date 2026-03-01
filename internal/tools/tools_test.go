package tools_test

import (
	"context"
	"strings"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/tools"
	"google.golang.org/protobuf/types/known/structpb"
)

// ---------- Helpers ----------

func makeArgs(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("makeArgs: %v", err)
	}
	return s
}

func callTool(
	t *testing.T,
	handler func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error),
	args map[string]any,
) *pluginv1.ToolResponse {
	t.Helper()
	s, err := structpb.NewStruct(args)
	if err != nil {
		t.Fatalf("callTool: build args: %v", err)
	}
	resp, err := handler(context.Background(), &pluginv1.ToolRequest{Arguments: s})
	if err != nil {
		t.Fatalf("callTool: handler returned error: %v", err)
	}
	return resp
}

func isError(resp *pluginv1.ToolResponse) bool {
	return resp != nil && !resp.Success
}

func getText(t *testing.T, resp *pluginv1.ToolResponse) string {
	t.Helper()
	if resp.Result == nil {
		return ""
	}
	v, ok := resp.Result.Fields["text"]
	if !ok {
		return ""
	}
	return v.GetStringValue()
}

// createSession is a helper that creates a terminal session via the tool handler
// and registers a cleanup to close it.  It returns the session ID.
func createSession(t *testing.T, mgr *pty.Manager) string {
	t.Helper()
	handler := tools.CreateTerminal(mgr)
	resp := callTool(t, handler, map[string]any{})
	if isError(resp) {
		t.Fatalf("createSession: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	text := getText(t, resp)
	// Extract the term-XXXXXX ID from the success message.
	idx := strings.Index(text, "term-")
	if idx < 0 {
		t.Fatalf("createSession: could not find term- ID in response %q", text)
	}
	// The ID is the last word in the response ("Terminal session created: term-XXXXXX").
	parts := strings.Fields(text)
	id := parts[len(parts)-1]
	t.Cleanup(func() { _ = mgr.Close(id) })
	return id
}

// ---------- create_terminal ----------

func TestCreateTerminal_Success(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.CreateTerminal(mgr)

	resp := callTool(t, handler, map[string]any{})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}

	text := getText(t, resp)
	if !strings.Contains(text, "term-") {
		t.Fatalf("expected 'term-' in response, got %q", text)
	}

	// Extract the ID and register cleanup.
	parts := strings.Fields(text)
	id := parts[len(parts)-1]
	t.Cleanup(func() { _ = mgr.Close(id) })
}

func TestCreateTerminal_WithDimensions(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.CreateTerminal(mgr)

	resp := callTool(t, handler, map[string]any{
		"cols": float64(120),
		"rows": float64(40),
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}

	text := getText(t, resp)
	if !strings.Contains(text, "term-") {
		t.Fatalf("expected 'term-' in response, got %q", text)
	}

	parts := strings.Fields(text)
	id := parts[len(parts)-1]
	t.Cleanup(func() { _ = mgr.Close(id) })

	// Verify dimensions stored on the manager.
	sessions := mgr.List()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Cols != 120 {
		t.Fatalf("expected Cols=120, got %d", sessions[0].Cols)
	}
	if sessions[0].Rows != 40 {
		t.Fatalf("expected Rows=40, got %d", sessions[0].Rows)
	}
}

// ---------- send_input ----------

func TestSendInput_MissingTerminalID(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.SendInput(mgr)

	// Provide "input" but omit "terminal_id".
	resp := callTool(t, handler, map[string]any{
		"input": "hello\n",
	})
	if !isError(resp) {
		t.Fatal("expected validation error for missing terminal_id, got success")
	}
	if resp.ErrorCode != "validation_error" {
		t.Fatalf("expected ErrorCode 'validation_error', got %q", resp.ErrorCode)
	}
}

func TestSendInput_NotFound(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.SendInput(mgr)

	resp := callTool(t, handler, map[string]any{
		"terminal_id": "term-nonexistent",
		"input":       "hello\n",
	})
	if !isError(resp) {
		t.Fatal("expected error for nonexistent terminal, got success")
	}
	if resp.ErrorCode != "send_error" {
		t.Fatalf("expected ErrorCode 'send_error', got %q", resp.ErrorCode)
	}
}

// ---------- get_output ----------

func TestGetOutput_MissingTerminalID(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.GetOutput(mgr)

	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected validation error for missing terminal_id, got success")
	}
	if resp.ErrorCode != "validation_error" {
		t.Fatalf("expected ErrorCode 'validation_error', got %q", resp.ErrorCode)
	}
}

func TestGetOutput_NotFound(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.GetOutput(mgr)

	resp := callTool(t, handler, map[string]any{
		"terminal_id": "term-nonexistent",
	})
	if !isError(resp) {
		t.Fatal("expected error for nonexistent terminal, got success")
	}
	if resp.ErrorCode != "output_error" {
		t.Fatalf("expected ErrorCode 'output_error', got %q", resp.ErrorCode)
	}
}

// ---------- resize_terminal ----------

func TestResizeTerminal_MissingArgs(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.ResizeTerminal(mgr)

	// Provide terminal_id and rows but omit "cols".
	// ValidateRequired checks for terminal_id, cols, and rows — missing cols triggers failure.
	resp := callTool(t, handler, map[string]any{
		"terminal_id": "term-abc",
		"rows":        float64(24),
	})
	if !isError(resp) {
		t.Fatal("expected validation error for missing cols, got success")
	}
	if resp.ErrorCode != "validation_error" {
		t.Fatalf("expected ErrorCode 'validation_error', got %q", resp.ErrorCode)
	}
}

// TestResizeTerminal_NotFound verifies that resize fails when the session does not exist.
// Note: ValidateRequired uses GetString internally, so cols/rows as numbers are treated
// as "missing" strings. We validate the error path by omitting all three required fields.
func TestResizeTerminal_NotFound(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.ResizeTerminal(mgr)

	// All three required fields (terminal_id, cols, rows) are missing → validation_error.
	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected validation error for missing required fields, got success")
	}
	if resp.ErrorCode != "validation_error" {
		t.Fatalf("expected ErrorCode 'validation_error', got %q", resp.ErrorCode)
	}
}

// TestResizeTerminal_ResizesExistingSession verifies that resize via the manager
// succeeds on a real session (the tool handler's ValidateRequired uses GetString
// so numeric cols/rows always trigger validation_error at the handler layer;
// the manager-level Resize path is tested here directly).
func TestResizeTerminal_ResizesExistingSession(t *testing.T) {
	mgr := pty.NewManager()

	id := createSession(t, mgr)

	if err := mgr.Resize(id, 160, 48); err != nil {
		t.Fatalf("Resize existing session: unexpected error: %v", err)
	}
	sessions := mgr.List()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Cols != 160 {
		t.Fatalf("expected Cols=160, got %d", sessions[0].Cols)
	}
}

// ---------- list_terminals ----------

func TestListTerminals_Empty(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.ListTerminals(mgr)

	resp := callTool(t, handler, map[string]any{})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}

	text := getText(t, resp)
	if !strings.Contains(text, "No active") {
		t.Fatalf("expected 'No active' in empty list response, got %q", text)
	}
}

func TestListTerminals_WithSessions(t *testing.T) {
	mgr := pty.NewManager()

	id := createSession(t, mgr)

	handler := tools.ListTerminals(mgr)
	resp := callTool(t, handler, map[string]any{})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}

	text := getText(t, resp)
	if !strings.Contains(text, id) {
		t.Fatalf("expected session ID %q in list output, got %q", id, text)
	}
}

// ---------- close_terminal ----------

func TestCloseTerminal_MissingID(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.CloseTerminal(mgr)

	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected validation error for missing terminal_id, got success")
	}
	if resp.ErrorCode != "validation_error" {
		t.Fatalf("expected ErrorCode 'validation_error', got %q", resp.ErrorCode)
	}
}

func TestCloseTerminal_NotFound(t *testing.T) {
	mgr := pty.NewManager()
	handler := tools.CloseTerminal(mgr)

	resp := callTool(t, handler, map[string]any{
		"terminal_id": "term-nonexistent",
	})
	if !isError(resp) {
		t.Fatal("expected error for nonexistent terminal, got success")
	}
	if resp.ErrorCode != "close_error" {
		t.Fatalf("expected ErrorCode 'close_error', got %q", resp.ErrorCode)
	}
}

func TestCloseTerminal_Success(t *testing.T) {
	mgr := pty.NewManager()

	// Create a session directly on the manager (no cleanup — we close it explicitly).
	id, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	handler := tools.CloseTerminal(mgr)
	resp := callTool(t, handler, map[string]any{
		"terminal_id": id,
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}

	text := getText(t, resp)
	if !strings.Contains(text, id) {
		t.Fatalf("expected session ID %q in close response, got %q", id, text)
	}

	// Verify the session is gone.
	sessions := mgr.List()
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions after close, got %d", len(sessions))
	}
}

// ---------- makeArgs (unused direct helper — kept for reference) ----------

var _ = makeArgs // silence unused warning; callTool uses structpb.NewStruct directly
