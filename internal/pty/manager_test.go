package pty

import (
	"strings"
	"testing"
	"time"
)

// TestManager_Create verifies that creating a session returns an ID starting with "term-"
// and that the session appears in the list.
func TestManager_Create(t *testing.T) {
	mgr := NewManager()

	id, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id) })

	if !strings.HasPrefix(id, "term-") {
		t.Fatalf("expected ID to start with 'term-', got %q", id)
	}

	sessions := mgr.List()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != id {
		t.Fatalf("expected session ID %q, got %q", id, sessions[0].ID)
	}
}

// TestManager_Create_DefaultShell verifies that creating with an empty shell
// falls back to $SHELL or /bin/sh and succeeds.
func TestManager_Create_DefaultShell(t *testing.T) {
	mgr := NewManager()

	id, err := mgr.Create("", 80, 24)
	if err != nil {
		t.Fatalf("Create with empty shell: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id) })

	if !strings.HasPrefix(id, "term-") {
		t.Fatalf("expected ID to start with 'term-', got %q", id)
	}
}

// TestManager_Create_CustomCols verifies that custom cols/rows are stored correctly.
func TestManager_Create_CustomCols(t *testing.T) {
	mgr := NewManager()

	id, err := mgr.Create("/bin/sh", 120, 40)
	if err != nil {
		t.Fatalf("Create with custom dims: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id) })

	sessions := mgr.List()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	s := sessions[0]
	if s.Cols != 120 {
		t.Fatalf("expected Cols=120, got %d", s.Cols)
	}
	if s.Rows != 40 {
		t.Fatalf("expected Rows=40, got %d", s.Rows)
	}
}

// TestManager_SendInput_NotFound verifies that SendInput returns an error for unknown IDs.
func TestManager_SendInput_NotFound(t *testing.T) {
	mgr := NewManager()

	err := mgr.SendInput("term-nonexistent", "hello\n")
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
}

// TestManager_GetOutput_NotFound verifies that GetOutput returns an error for unknown IDs.
func TestManager_GetOutput_NotFound(t *testing.T) {
	mgr := NewManager()

	_, err := mgr.GetOutput("term-nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
}

// TestManager_Resize_NotFound verifies that Resize returns an error for unknown IDs.
func TestManager_Resize_NotFound(t *testing.T) {
	mgr := NewManager()

	err := mgr.Resize("term-nonexistent", 120, 40)
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
}

// TestManager_Close_NotFound verifies that Close returns an error for unknown IDs.
func TestManager_Close_NotFound(t *testing.T) {
	mgr := NewManager()

	err := mgr.Close("term-nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
}

// TestManager_Close_Success verifies that a session can be closed and a second
// close returns an error.
func TestManager_Close_Success(t *testing.T) {
	mgr := NewManager()

	id, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	if err := mgr.Close(id); err != nil {
		t.Fatalf("first Close: unexpected error: %v", err)
	}

	// Second close should fail — session no longer exists.
	if err := mgr.Close(id); err == nil {
		t.Fatal("expected error on second Close, got nil")
	}
}

// TestManager_List_Empty verifies that a fresh manager returns an empty slice.
func TestManager_List_Empty(t *testing.T) {
	mgr := NewManager()

	sessions := mgr.List()
	if len(sessions) != 0 {
		t.Fatalf("expected empty list, got %d sessions", len(sessions))
	}
}

// TestManager_List_Multiple verifies that multiple sessions all appear in the list.
func TestManager_List_Multiple(t *testing.T) {
	mgr := NewManager()

	id1, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id1) })

	id2, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id2) })

	sessions := mgr.List()
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
}

// TestManager_SendInput_GetOutput creates a session, sends a command, and verifies
// that output is captured (non-empty) after a short delay.
func TestManager_SendInput_GetOutput(t *testing.T) {
	mgr := NewManager()

	id, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id) })

	if err := mgr.SendInput(id, "echo hello\n"); err != nil {
		t.Fatalf("SendInput: unexpected error: %v", err)
	}

	// Give the shell a moment to process and write output.
	time.Sleep(100 * time.Millisecond)

	output, err := mgr.GetOutput(id)
	if err != nil {
		t.Fatalf("GetOutput: unexpected error: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output after echo command, got empty string")
	}
}

// TestManager_Resize_Success verifies that resizing an existing session succeeds.
func TestManager_Resize_Success(t *testing.T) {
	mgr := NewManager()

	id, err := mgr.Create("/bin/sh", 80, 24)
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close(id) })

	if err := mgr.Resize(id, 200, 50); err != nil {
		t.Fatalf("Resize: unexpected error: %v", err)
	}

	// Verify the stored dimensions were updated.
	sessions := mgr.List()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	s := sessions[0]
	if s.Cols != 200 {
		t.Fatalf("expected Cols=200 after resize, got %d", s.Cols)
	}
	if s.Rows != 50 {
		t.Fatalf("expected Rows=50 after resize, got %d", s.Rows)
	}
}
