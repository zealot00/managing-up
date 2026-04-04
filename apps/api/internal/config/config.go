package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	Database Database
	MCP      MCPConfig
}

type MCPConfig struct {
	Servers []MCPServerConfig
	Enabled bool
}

type MCPServerConfig struct {
	Name    string
	URL     string
	Command string
	Args    []string
	Env     []string
}

func Load() Config {
	loadEnvFiles()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		Port:     port,
		Database: loadDatabase(),
		MCP:      loadMCPConfig(),
	}
}

func loadEnvFiles() {
	candidates := []string{
		".env",
		filepath.Join("apps", "api", ".env"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("failed to load %s: %v", path, err)
			} else {
				log.Printf("loaded environment from %s", path)
			}
			return
		}
	}
}

func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}

func loadMCPConfig() MCPConfig {
	mcpEnabled := os.Getenv("MCP_ENABLED")
	if mcpEnabled == "" {
		mcpEnabled = "false"
	}

	var servers []MCPServerConfig
	for i := 1; ; i++ {
		prefix := fmt.Sprintf("MCP_SERVER_%d_", i)
		name := os.Getenv(prefix + "NAME")
		if name == "" {
			break
		}

		server := MCPServerConfig{
			Name: name,
			URL:  os.Getenv(prefix + "URL"),
		}

		if cmd := os.Getenv(prefix + "COMMAND"); cmd != "" {
			server.Command = cmd
			if args := os.Getenv(prefix + "ARGS"); args != "" {
				server.Args = strings.Split(args, ",")
			}
			if env := os.Getenv(prefix + "ENV"); env != "" {
				server.Env = strings.Split(env, ",")
			}
		}

		servers = append(servers, server)
	}

	enabled, _ := strconv.ParseBool(mcpEnabled)
	return MCPConfig{
		Servers: servers,
		Enabled: enabled,
	}
}
