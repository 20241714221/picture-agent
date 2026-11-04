package mcp

import (
	"context"
	"fmt"
	"os/exec"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client 封装 MCP Client 和对应的 Session
type Client struct {
	Client  *sdkmcp.Client
	Session *sdkmcp.ClientSession
}

// NewClient 创建 MCP Client，并使用 CommandTransport 启动 Server 子进程，
// 建立连接后返回包含 Session 的 Client。
//
// projectRoot 用于解析相对路径，确保 command 中的相对路径以项目根目录为基准。
func NewClient(
	ctx context.Context,
	serverName string,
	command string,
	args []string,
	projectRoot string,
) (*Client, error) {
	if command == "" {
		return nil, fmt.Errorf("MCP Server %s 的 command 不能为空", serverName)
	}

	impl := &sdkmcp.Implementation{
		Name:    "picture-agent-mcp-client",
		Version: "1.0.0",
	}

	mcpClient := sdkmcp.NewClient(impl, nil)

	cmd := exec.Command(command, args...)
	cmd.Dir = projectRoot

	transport := &sdkmcp.CommandTransport{
		Command: cmd,
	}

	session, err := mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("连接 MCP Server %s 失败: %w", serverName, err)
	}

	return &Client{
		Client:  mcpClient,
		Session: session,
	}, nil
}

// Close 关闭 MCP Session，释放子进程资源
func (c *Client) Close() error {
	if c.Session != nil {
		return c.Session.Close()
	}
	return nil
}
