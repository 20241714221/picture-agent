package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Manager 管理多个 MCP Server 的连接
type Manager struct {
	projectRoot string
	clients     map[string]*Client // key: server name
	mu          sync.Mutex
}

// NewManager 创建 MCP Manager
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		clients:     make(map[string]*Client),
	}
}

// StartServers 根据配置启动所有启用的 MCP Server
func (m *Manager) StartServers(ctx context.Context, cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, serverCfg := range cfg.Servers {
		if !serverCfg.Enabled {
			continue
		}

		if _, exists := m.clients[serverCfg.Name]; exists {
			return fmt.Errorf("MCP Server 名称重复: %s", serverCfg.Name)
		}

		client, err := NewClient(
			ctx,
			serverCfg.Name,
			serverCfg.Command,
			serverCfg.Args,
			m.projectRoot,
		)
		if err != nil {
			// 启动失败时关闭已连接的 Server，避免资源泄漏
			m.closeAllLocked()
			return err
		}

		m.clients[serverCfg.Name] = client
	}

	return nil
}

// GetClient 根据 Server 名称获取对应的 MCP Client
func (m *Manager) GetClient(name string) (*Client, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.clients[name]
	return c, ok
}

// GetAllClients 返回所有已连接的 MCP Client
func (m *Manager) GetAllClients() map[string]*Client {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 返回副本，避免外部修改
	result := make(map[string]*Client, len(m.clients))
	for k, v := range m.clients {
		result[k] = v
	}
	return result
}

// CloseAll 关闭所有 MCP Session
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeAllLocked()
}

func (m *Manager) closeAllLocked() {
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			// 记录错误但继续关闭其他连接
			fmt.Printf("关闭 MCP Server %s 失败: %v\n", name, err)
		}
		delete(m.clients, name)
	}
}
