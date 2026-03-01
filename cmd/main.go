package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/plugin-devtools-terminal/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("devtools.terminal").
		Version("0.1.0").
		Description("PTY terminal session manager — create and control pseudo-terminal sessions").
		Author("Orchestra").
		Binary("devtools-terminal")

	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("devtools.terminal: %v", err)
	}
}
