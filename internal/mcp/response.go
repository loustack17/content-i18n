package mcp

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func jsonResponse(v any) (*mcp.CallToolResult, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}

func textResponse(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func errorResponse(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}
