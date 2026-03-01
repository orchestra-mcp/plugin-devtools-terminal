package pty

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

// SessionInfo holds metadata about a terminal session for listing.
type SessionInfo struct {
	ID   string `json:"id"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// Session represents a single PTY terminal session.
type Session struct {
	ID     string
	cmd    *exec.Cmd
	ptmx   *os.File
	output bytes.Buffer
	mu     sync.Mutex
	cols   int
	rows   int
}

// Manager manages multiple PTY terminal sessions.
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a new PTY session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// generateID creates a session ID in the format "term-" + 6 random hex chars.
func generateID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return "term-" + hex.EncodeToString(b), nil
}

// Create starts a new PTY terminal session with the given shell and dimensions.
func (m *Manager) Create(shell string, cols, rows int) (string, error) {
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	id, err := generateID()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()

	winSize := &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	}

	ptmx, err := pty.StartWithSize(cmd, winSize)
	if err != nil {
		return "", fmt.Errorf("start pty: %w", err)
	}

	sess := &Session{
		ID:   id,
		cmd:  cmd,
		ptmx: ptmx,
		cols: cols,
		rows: rows,
	}

	// Background goroutine reads from ptmx into buffer.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				sess.mu.Lock()
				sess.output.Write(buf[:n])
				sess.mu.Unlock()
			}
			if err != nil {
				if err != io.EOF {
					// PTY closed or process exited; stop reading.
				}
				return
			}
		}
	}()

	m.mu.Lock()
	m.sessions[id] = sess
	m.mu.Unlock()

	return id, nil
}

// SendInput writes input to the terminal session's PTY.
func (m *Manager) SendInput(id, input string) error {
	m.mu.RLock()
	sess, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("terminal session %q not found", id)
	}

	_, err := sess.ptmx.WriteString(input)
	if err != nil {
		return fmt.Errorf("write to pty: %w", err)
	}
	return nil
}

// GetOutput returns accumulated output from the terminal session and clears the buffer.
func (m *Manager) GetOutput(id string) (string, error) {
	m.mu.RLock()
	sess, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("terminal session %q not found", id)
	}

	sess.mu.Lock()
	out := sess.output.String()
	sess.output.Reset()
	sess.mu.Unlock()

	return out, nil
}

// Resize changes the terminal dimensions for a session.
func (m *Manager) Resize(id string, cols, rows int) error {
	m.mu.RLock()
	sess, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("terminal session %q not found", id)
	}

	winSize := &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	}
	if err := pty.Setsize(sess.ptmx, winSize); err != nil {
		return fmt.Errorf("resize pty: %w", err)
	}

	sess.mu.Lock()
	sess.cols = cols
	sess.rows = rows
	sess.mu.Unlock()

	return nil
}

// List returns information about all active terminal sessions.
func (m *Manager) List() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]SessionInfo, 0, len(m.sessions))
	for _, sess := range m.sessions {
		sess.mu.Lock()
		infos = append(infos, SessionInfo{
			ID:   sess.ID,
			Cols: sess.cols,
			Rows: sess.rows,
		})
		sess.mu.Unlock()
	}
	return infos
}

// Close terminates a terminal session and cleans up resources.
func (m *Manager) Close(id string) error {
	m.mu.Lock()
	sess, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("terminal session %q not found", id)
	}
	delete(m.sessions, id)
	m.mu.Unlock()

	// Close PTY (this will also signal the read goroutine to stop).
	_ = sess.ptmx.Close()

	// Kill the process if still running.
	if sess.cmd.Process != nil {
		_ = sess.cmd.Process.Kill()
		_ = sess.cmd.Wait()
	}

	return nil
}
