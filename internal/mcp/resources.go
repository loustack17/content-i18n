package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
	for _, spec := range allResourceSpecs(s) {
		switch d := spec.def.(type) {
		case mcp.Resource:
			s.mcp.AddResource(d, spec.handler(s))
		case mcp.ResourceTemplate:
			s.mcp.AddResourceTemplate(d, spec.handler(s))
		}
	}
}
