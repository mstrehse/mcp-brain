package actions

import (
	"context"
	"runtime"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/ask/cli"
)

type AskAction struct {
	AskRepository contracts.AskRepository
}

func NewAskAction() *AskAction {
	var askRepo contracts.AskRepository
	switch runtime.GOOS {
	case "linux":
		askRepo = &cli.LinuxRepository{}
	case "darwin":
		askRepo = &cli.OsxRepository{}
	}

	return &AskAction{
		AskRepository: askRepo,
	}
}

func (a *AskAction) AskUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if a.AskRepository == nil {
		return mcp.NewToolResultError("Unsupported OS"), nil
	}

	question, err := request.RequireString("question")
	if err != nil {
		return mcp.NewToolResultError("Missing 'question' parameter: " + err.Error()), nil
	}

	response, err := a.AskRepository.Ask(question)
	if err != nil {
		return mcp.NewToolResultError("Failed to show dialog: " + err.Error()), nil
	}

	return mcp.NewToolResultText(response), nil
}
