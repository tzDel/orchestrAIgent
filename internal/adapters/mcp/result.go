package mcp

import mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

func newTextContent(text string) mcpsdk.Content {
	return &mcpsdk.TextContent{Text: text}
}

func newErrorResult(message string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{newTextContent(message)},
		IsError: true,
	}
}

func newSuccessResult(message string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{newTextContent(message)},
		IsError: false,
	}
}
