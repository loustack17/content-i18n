package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/loustack17/content-i18n/internal/config"
)

type Server struct {
	mcp        *server.MCPServer
	cfg        *config.Config
	configPath string
}

func NewServer(cfg *config.Config, configPath string) *Server {
	s := server.NewMCPServer(
		"content-i18n",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	srv := &Server{mcp: s, cfg: cfg, configPath: configPath}
	srv.registerTools()
	srv.registerResources()
	return srv
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcp)
}
