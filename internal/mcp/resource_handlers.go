package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleConfigResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "content-i18n://config",
			MIMEType: "application/yaml",
			Text:     string(data),
		},
	}, nil
}

func (s *Server) handleGlossaryResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return s.handleFileResource("content-i18n://glossary", "application/yaml", s.cfg.Style.Glossary, "glossary")
}

func (s *Server) handleStylePackResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return s.handleFileResource("content-i18n://style-pack", "application/yaml", s.cfg.Style.Pack, "style pack")
}

func (s *Server) handleFileResource(uri, mime, cfgPath, label string) ([]mcp.ResourceContents, error) {
	if cfgPath == "" {
		return nil, fmt.Errorf("no %s configured", label)
	}
	if !filepath.IsAbs(cfgPath) {
		cfgPath = filepath.Join(s.cfg.ConfigDir, cfgPath)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mime,
			Text:     string(data),
		},
	}, nil
}

func (s *Server) handlePostResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI

	var lang, path string
	if strings.HasPrefix(uri, "content-i18n://post/") {
		rest := strings.TrimPrefix(uri, "content-i18n://post/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) == 2 {
			lang = parts[0]
			path = parts[1]
		}
	}

	if lang == "" || path == "" {
		return nil, fmt.Errorf("invalid post URI: %s", uri)
	}

	targetDir := s.cfg.Paths.Targets[lang]
	if lang == s.cfg.Project.SourceLanguage {
		targetDir = s.cfg.Paths.Source
	}
	if targetDir == "" {
		return nil, fmt.Errorf("unknown language: %s", lang)
	}

	if !filepath.IsAbs(targetDir) {
		targetDir = filepath.Join(s.cfg.ConfigDir, targetDir)
	}

	fullPath := filepath.Join(targetDir, path)
	cleanPath := filepath.Clean(fullPath)
	cleanTargetDir := filepath.Clean(targetDir)

	rel, err := filepath.Rel(cleanTargetDir, cleanPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path traversal detected: %s", path)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}
