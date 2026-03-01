package internal

import (
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/pty"
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// ToolsPlugin holds the PTY manager and registers all terminal tools.
type ToolsPlugin struct{}

// RegisterTools registers all 6 terminal tools with the plugin builder.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	mgr := pty.NewManager()

	builder.RegisterTool("create_terminal",
		"Create a new PTY terminal session",
		tools.CreateTerminalSchema(), tools.CreateTerminal(mgr))

	builder.RegisterTool("send_input",
		"Send input to a terminal session",
		tools.SendInputSchema(), tools.SendInput(mgr))

	builder.RegisterTool("get_output",
		"Get accumulated output from a terminal session",
		tools.GetOutputSchema(), tools.GetOutput(mgr))

	builder.RegisterTool("resize_terminal",
		"Resize a terminal session",
		tools.ResizeTerminalSchema(), tools.ResizeTerminal(mgr))

	builder.RegisterTool("list_terminals",
		"List all active terminal sessions",
		tools.ListTerminalsSchema(), tools.ListTerminals(mgr))

	builder.RegisterTool("close_terminal",
		"Close a terminal session",
		tools.CloseTerminalSchema(), tools.CloseTerminal(mgr))
}
