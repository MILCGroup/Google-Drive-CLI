package cli

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/ai"
	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/option"
)

type AICmd struct {
	Models   AIModelsCmd   `cmd:"" help:"Manage AI models"`
	Generate AIGenerateCmd `cmd:"" help:"Generate content with AI"`
}

type AIModelsCmd struct {
	List AIModelsListCmd `cmd:"" help:"List available AI models"`
	Get  AIModelsGetCmd  `cmd:"" help:"Get details about a specific model"`
}

type AIGenerateCmd struct {
	Text   AIGenerateTextCmd   `cmd:"" help:"Generate text content"`
	Stream AIGenerateStreamCmd `cmd:"" help:"Stream text generation"`
}

type AIModelsListCmd struct {
	Limit     int32  `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
}

type AIModelsGetCmd struct {
	ModelName string `arg:"" name:"model-name" help:"Model name (e.g., models/gemini-pro)"`
}

type AIGenerateTextCmd struct {
	Model       string  `arg:"" name:"model" help:"Model name (e.g., models/gemini-pro)"`
	Prompt      string  `arg:"" name:"prompt" help:"Text prompt"`
	Temperature float32 `help:"Temperature for generation (0.0-1.0)" default:"0.7" name:"temperature"`
}

type AIGenerateStreamCmd struct {
	Model       string  `arg:"" name:"model" help:"Model name (e.g., models/gemini-pro)"`
	Prompt      string  `arg:"" name:"prompt" help:"Text prompt"`
	Temperature float32 `help:"Temperature for generation (0.0-1.0)" default:"0.7" name:"temperature"`
}

func (cmd *AIModelsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getAIManager(ctx, flags)
	if err != nil {
		return out.WriteError("ai.models.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListModels(ctx, reqCtx, cmd.Limit, cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "ai.models.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("ai.models.list", result.Models)
	}
	return out.WriteSuccess("ai.models.list", result)
}

func (cmd *AIModelsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getAIManager(ctx, flags)
	if err != nil {
		return out.WriteError("ai.models.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetModel(ctx, reqCtx, cmd.ModelName)
	if err != nil {
		return handleCLIError(out, "ai.models.get", err)
	}

	return out.WriteSuccess("ai.models.get", result)
}

func (cmd *AIGenerateTextCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getAIManager(ctx, flags)
	if err != nil {
		return out.WriteError("ai.generate.text", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.GenerateContent(ctx, reqCtx, cmd.Model, cmd.Prompt, cmd.Temperature)
	if err != nil {
		return handleCLIError(out, "ai.generate.text", err)
	}

	return out.WriteSuccess("ai.generate.text", result)
}

func (cmd *AIGenerateStreamCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getAIManager(ctx, flags)
	if err != nil {
		return out.WriteError("ai.generate.stream", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeMutation

	contentChan, errChan := mgr.StreamGenerateContent(ctx, reqCtx, cmd.Model, cmd.Prompt, cmd.Temperature)

	if !flags.Quiet {
		fmt.Println("Generated content:")
		fmt.Println("---")
	}

	for {
		select {
		case content, ok := <-contentChan:
			if !ok {
				if !flags.Quiet {
					fmt.Println("---")
				}
				return out.WriteSuccess("ai.generate.stream", map[string]string{"status": "completed"})
			}
			if !flags.Quiet {
				fmt.Print(content.Text)
			}
		case err := <-errChan:
			if err != nil {
				return handleCLIError(out, "ai.generate.stream", err)
			}
		case <-ctx.Done():
			return handleCLIError(out, "ai.generate.stream", ctx.Err())
		}
	}
}

func getAIManager(ctx context.Context, flags types.GlobalFlags) (*ai.Manager, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, err
	}

	tokenSource := authMgr.GetTokenSource(ctx, creds)
	opt := option.WithTokenSource(tokenSource)

	mgr, err := ai.NewManager(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, nil
}
