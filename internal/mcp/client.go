package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

type Client struct {
	command *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	
	nextID  uint64
	pending map[uint64]chan JSONRPCMessage
	mu      sync.Mutex
	
	Tools []Tool
}

func NewClient(ctx context.Context, cmdName string, args ...string) (*Client, error) {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	
	c := &Client{
		command: cmd,
		stdin:   stdin,
		stdout:  stdout,
		pending: make(map[uint64]chan JSONRPCMessage),
	}
	
	go c.readLoop()
	
	if err := c.Initialize(); err != nil {
		c.Close()
		return nil, fmt.Errorf("initialize failed: %v", err)
	}
	
	if err := c.RefreshTools(); err != nil {
		c.Close()
		return nil, fmt.Errorf("refresh tools failed: %v", err)
	}
	
	return c, nil
}

func (c *Client) readLoop() {
	scanner := bufio.NewScanner(c.stdout)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg JSONRPCMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue // skip invalid json
		}
		
		if msg.ID != nil {
			var id uint64
			if err := json.Unmarshal(*msg.ID, &id); err == nil {
				c.mu.Lock()
				ch, ok := c.pending[id]
				if ok {
					delete(c.pending, id)
				}
				c.mu.Unlock()
				
				if ok {
					ch <- msg
				}
			}
		}
	}
}

func (c *Client) sendRequest(method string, params interface{}) (JSONRPCMessage, error) {
	id := atomic.AddUint64(&c.nextID, 1)
	
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		return JSONRPCMessage{}, err
	}
	
	ch := make(chan JSONRPCMessage, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()
	
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return JSONRPCMessage{}, err
	}
	
	// Wait for response
	msg := <-ch
	if msg.Error != nil {
		return msg, fmt.Errorf("RPC error %d: %s", msg.Error.Code, msg.Error.Message)
	}
	return msg, nil
}

func (c *Client) sendNotification(method string, params interface{}) error {
	notif := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	_, err = c.stdin.Write(append(data, '\n'))
	return err
}

func (c *Client) Initialize() error {
	req := InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{},
		ClientInfo: Implementation{
			Name:    "cure-code",
			Version: "1.0",
		},
	}
	
	_, err := c.sendRequest("initialize", req)
	if err != nil {
		return err
	}
	
	// Must send initialized notification after successful initialize
	return c.sendNotification("notifications/initialized", map[string]interface{}{})
}

func (c *Client) RefreshTools() error {
	msg, err := c.sendRequest("tools/list", map[string]interface{}{})
	if err != nil {
		return err
	}
	
	var result ListToolsResult
	if err := json.Unmarshal(msg.Result, &result); err != nil {
		return err
	}
	
	c.Tools = result.Tools
	return nil
}

func (c *Client) CallTool(name string, arguments map[string]interface{}) (string, bool, error) {
	req := CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}
	
	msg, err := c.sendRequest("tools/call", req)
	if err != nil {
		return "", true, err
	}
	
	var result CallToolResult
	if err := json.Unmarshal(msg.Result, &result); err != nil {
		return "", true, err
	}
	
	output := ""
	for _, content := range result.Content {
		if content.Type == "text" {
			output += content.Text + "\n"
		}
	}
	
	return output, result.IsError, nil
}

func (c *Client) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.command != nil && c.command.Process != nil {
		return c.command.Process.Kill()
	}
	return nil
}
