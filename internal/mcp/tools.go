package mcp

func (s *Server) registerTools() {
	for _, spec := range allToolSpecs(s) {
		s.mcp.AddTool(spec.def, spec.handler)
	}
}
