package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
	"github.com/loustack17/content-i18n/internal/core"
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

func (s *Server) registerTools() {
	s.mcp.AddTool(mcp.NewTool("content_i18n_status",
		mcp.WithDescription("Return project status: source/target counts, missing translations"),
	), s.handleStatus)

	s.mcp.AddTool(mcp.NewTool("content_i18n_list_missing",
		mcp.WithDescription("List files with missing or stale translations"),
		mcp.WithString("language",
			mcp.Description("Filter by target language"),
		),
	), s.handleListMissing)

	s.mcp.AddTool(mcp.NewTool("content_i18n_read_source",
		mcp.WithDescription("Read source file content for translation context"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to source file"),
		),
	), s.handleReadSource)

	s.mcp.AddTool(mcp.NewTool("content_i18n_create_work_packet",
		mcp.WithDescription("Create a work packet for translating a source file"),
		mcp.WithString("source",
			mcp.Required(),
			mcp.Description("Absolute path to source file"),
		),
		mcp.WithString("language",
			mcp.Required(),
			mcp.Description("Target language code"),
		),
	), s.handleCreateWorkPacket)

	s.mcp.AddTool(mcp.NewTool("content_i18n_write_translation_target",
		mcp.WithDescription("Write translated content to a work packet target file"),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Work packet slug"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Translated markdown content"),
		),
	), s.handleWriteTranslationTarget)

	s.mcp.AddTool(mcp.NewTool("content_i18n_validate_translation",
		mcp.WithDescription("Validate a translated file for integrity issues"),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("Path to target file to validate"),
		),
		mcp.WithString("source",
			mcp.Description("Path to source file for comparison"),
		),
	), s.handleValidateTranslation)

	s.mcp.AddTool(mcp.NewTool("content_i18n_validate_site",
		mcp.WithDescription("Validate Hugo site URL policy and canonical routes"),
	), s.handleValidateSite)
}

func (s *Server) handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	report, err := core.Status(s.cfg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	out, _ := json.MarshalIndent(map[string]any{
		"project_type":              report.ProjectType,
		"source_language":           report.SourceLanguage,
		"source_path":               report.SourcePath,
		"target_languages":          report.TargetLanguages,
		"source_file_count":         report.SourceFileCount,
		"target_file_count":         report.TargetFileCount,
		"missing_translation_count": report.MissingCount,
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func (s *Server) handleListMissing(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	files, err := core.List(s.cfg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lang := req.GetString("language", "")
	missing := content.MissingTranslations(files)

	var results []map[string]string
	for _, f := range missing {
		if lang != "" && f.Language != lang {
			continue
		}
		results = append(results, map[string]string{
			"source": f.SourcePath,
			"target": f.TargetPath,
			"lang":   f.Language,
			"status": string(f.Status),
		})
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func (s *Server) handleReadSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleCreateWorkPacket(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, err := req.RequireString("source")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lang, err := req.RequireString("language")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	slug := core.SlugFromPath(source, s.cfg.Paths.Source)
	packet, err := core.GenerateWorkPacket(s.cfg, source, lang)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	out, _ := json.MarshalIndent(map[string]string{
		"slug":     slug,
		"dir":      packet.Dir,
		"source":   packet.SourcePath,
		"target":   packet.TargetPath,
		"prompt":   packet.PromptPath,
		"glossary": packet.GlossaryPath,
		"style":    packet.StylePath,
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func (s *Server) handleWriteTranslationTarget(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slug, err := req.RequireString("slug")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	targetPath := filepath.Join("work", slug, "target.md")
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("wrote %d bytes to %s", len(content), targetPath)), nil
}

func (s *Server) handleValidateTranslation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	source := req.GetString("source", "")

	opts := &core.ValidateOptions{
		SourcePath: source,
		Config:     s.cfg,
	}
	result, err := core.ValidateContent(file, opts)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if result.Passed {
		return mcp.NewToolResultText("PASS"), nil
	}

	var lines []string
	for _, v := range result.Violations {
		lines = append(lines, fmt.Sprintf("[%s] %s: %s (fix: %s)", v.Field, v.Section, v.Message, v.SuggestedFix))
	}
	return mcp.NewToolResultText("FAIL\n" + strings.Join(lines, "\n")), nil
}

func (s *Server) handleValidateSite(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.cfg.Adapter.Name != core.AdapterHugo {
		return mcp.NewToolResultError(fmt.Sprintf("validate-site only supports hugo adapter (got: %s)", s.cfg.Adapter.Name)), nil
	}

	warnings := core.ValidateSiteConfig(s.cfg)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	hugoRoot := s.cfg.ConfigDir
	result, err := core.ValidateSite(s.cfg, hugoRoot)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if result.Passed {
		return mcp.NewToolResultText("PASS"), nil
	}

	out, _ := json.MarshalIndent(map[string]any{
		"hugo_output": result.HugoOutput,
		"violations":  result.Violations,
	}, "", "  ")
	return mcp.NewToolResultText("FAIL\n" + string(out)), nil
}

func (s *Server) registerResources() {
	s.mcp.AddResource(
		mcp.NewResource("content-i18n://config", "Project configuration", mcp.WithMIMEType("application/yaml")),
		s.handleConfigResource,
	)

	s.mcp.AddResource(
		mcp.NewResource("content-i18n://glossary", "Translation glossary", mcp.WithMIMEType("application/yaml")),
		s.handleGlossaryResource,
	)

	s.mcp.AddResource(
		mcp.NewResource("content-i18n://style-pack", "Style pack configuration", mcp.WithMIMEType("application/yaml")),
		s.handleStylePackResource,
	)

	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"content-i18n://post/{language}/{path}",
			"Post content by language and path",
		),
		s.handlePostResource,
	)
}

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
