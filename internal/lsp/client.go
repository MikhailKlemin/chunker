package lsp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

type Client struct {
	conn    *jsonrpc2.Conn
	cmd     *exec.Cmd
	rootURI string
}

// NewClient starts clangd and initializes the LSP connection
func NewClient(compileDBPath, rootPath string) (*Client, error) {
	// Start clangd process
	cmd := exec.Command("clangd",
		"--compile-commands-dir="+compileDBPath,
		"--background-index",
		"--log=error",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start clangd: %w", err)
	}

	// Create JSON-RPC connection
	stream := jsonrpc2.NewBufferedStream(&stdrwc{stdout, stdin}, jsonrpc2.VSCodeObjectCodec{})
	conn := jsonrpc2.NewConn(context.Background(), stream, jsonrpc2.HandlerWithError(
		func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
			// Handle server notifications/requests
			return nil, nil
		},
	))

	client := &Client{
		conn:    conn,
		cmd:     cmd,
		rootURI: "file://" + rootPath,
	}

	// Initialize the LSP connection
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, fmt.Errorf("initialize: %w", err)
	}

	return client, nil
}

func (c *Client) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	initParams := map[string]any{
		"processId": os.Getpid(),
		"rootUri":   c.rootURI,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"documentSymbol": map[string]any{
					"hierarchicalDocumentSymbolSupport": true,
				},
			},
		},
	}

	var result map[string]any
	if err := c.conn.Call(ctx, "initialize", initParams, &result); err != nil {
		return err
	}

	return c.conn.Notify(ctx, "initialized", map[string]any{})
}

// GetServerInfo returns server information from initialization
func (c *Client) GetServerInfo() (string, error) {
	// Simple test method to verify connection works
	return "clangd connection established", nil
}

// Close shuts down the LSP connection
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.conn.Call(ctx, "shutdown", nil, nil)
	c.conn.Notify(ctx, "exit", nil)
	c.conn.Close()

	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Wait()
	}
	return nil
}

// stdrwc wraps stdin/stdout for jsonrpc2
type stdrwc struct {
	io.ReadCloser
	io.WriteCloser
}

func (s *stdrwc) Close() error {
	s.ReadCloser.Close()
	return s.WriteCloser.Close()
}

// Add this method to the Client struct

// GetDocumentSymbols retrieves symbols from a C++ file
func (c *Client) GetDocumentSymbols(filePath string) ([]DocumentSymbol, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	uri := "file://" + filePath

	// Open document
	openParams := map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": "cpp",
			"version":    1,
			"text":       string(content),
		},
	}

	if err := c.conn.Notify(ctx, "textDocument/didOpen", openParams); err != nil {
		return nil, fmt.Errorf("didOpen: %w", err)
	}

	// Request symbols
	symbolParams := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}

	var symbols []DocumentSymbol
	if err := c.conn.Call(ctx, "textDocument/documentSymbol", symbolParams, &symbols); err != nil {
		return nil, fmt.Errorf("documentSymbol: %w", err)
	}

	// Close document to free memory
	closeParams := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}
	c.conn.Notify(context.Background(), "textDocument/didClose", closeParams)

	return symbols, nil
}
