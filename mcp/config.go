package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ServerConfig 描述一个 MCP Server 的启动配置
type ServerConfig struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Enabled bool     `json:"enabled"`
}

// Config 是 MCP 配置文件的结构
type Config struct {
	Servers []ServerConfig `json:"servers"`
}

// LoadConfig 从指定路径读取 MCP 配置文件。
// 如果文件不存在，返回空配置（不报错），保证原有 Agent 仍可启动。
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Servers: []ServerConfig{}}, nil
		}
		return nil, fmt.Errorf("读取 MCP 配置失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 MCP 配置失败: %w", err)
	}

	// 规范化：确保 Servers 非 nil
	if cfg.Servers == nil {
		cfg.Servers = []ServerConfig{}
	}

	return &cfg, nil
}

// DefaultConfigPath 返回默认配置文件路径（基于项目根目录）
func DefaultConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, "config", "mcp.json")
}
