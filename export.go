package devtoolsterminal

import (
	"github.com/orchestra-mcp/plugin-devtools-terminal/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all terminal tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)
}
