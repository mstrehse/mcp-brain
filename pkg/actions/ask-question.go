package actions

import (
	"context"
	"runtime"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mstrehse/mcp-brain/pkg/contracts"
	"github.com/mstrehse/mcp-brain/pkg/repositories/ask/cli"
)

type AskQuestionAction struct {
	AskRepository contracts.AskRepository
}

func NewAskQuestionAction() *AskQuestionAction {
	var askRepo contracts.AskRepository
	switch runtime.GOOS {
	case "linux":
		askRepo = &cli.LinuxRepository{}
	case "darwin":
		askRepo = &cli.OsxRepository{}
	}

	return &AskQuestionAction{
		AskRepository: askRepo,
	}
}

func (a *AskQuestionAction) AskQuestion(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
