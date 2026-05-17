package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type resourceSpec struct {
	def     any
	handler func(*Server) func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error)
}

func allResourceSpecs(s *Server) []resourceSpec {
	return []resourceSpec{
		{
			def: mcp.NewResource("content-i18n://config", "Project configuration", mcp.WithMIMEType("application/yaml")),
			handler: func(*Server) func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return s.handleConfigResource
			},
		},
		{
			def: mcp.NewResource("content-i18n://glossary", "Translation glossary", mcp.WithMIMEType("application/yaml")),
			handler: func(*Server) func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return s.handleGlossaryResource
			},
		},
		{
			def: mcp.NewResource("content-i18n://style-pack", "Style pack configuration", mcp.WithMIMEType("application/yaml")),
			handler: func(*Server) func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return s.handleStylePackResource
			},
		},
		{
			def: mcp.NewResourceTemplate(
				"content-i18n://post/{language}/{path}",
				"Post content by language and path",
			),
			handler: func(*Server) func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return s.handlePostResource
			},
		},
	}
}
