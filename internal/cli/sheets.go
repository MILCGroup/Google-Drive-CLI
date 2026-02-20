package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/files"
	sheetsmgr "github.com/dl-alexandre/gdrv/internal/sheets"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
	sheetsapi "google.golang.org/api/sheets/v4"
)

type SheetsCmd struct {
	List        SheetsListCmd        `cmd:"" help:"List spreadsheets"`
	Get         SheetsGetCmd         `cmd:"" help:"Get spreadsheet metadata"`
	Create      SheetsCreateCmd      `cmd:"" help:"Create a spreadsheet"`
	BatchUpdate SheetsBatchUpdateCmd `cmd:"batch-update" help:"Batch update spreadsheet"`
	Values      SheetsValuesCmd      `cmd:"" help:"Spreadsheet values operations"`
}

type SheetsValuesCmd struct {
	Get    SheetsValuesGetCmd    `cmd:"" help:"Get values from a range"`
	Update SheetsValuesUpdateCmd `cmd:"" help:"Update values in a range"`
	Append SheetsValuesAppendCmd `cmd:"" help:"Append values to a range"`
	Clear  SheetsValuesClearCmd  `cmd:"" help:"Clear values from a range"`
}

type SheetsListCmd struct {
	Parent    string `help:"Parent folder ID" name:"parent"`
	Query     string `help:"Additional search query" name:"query"`
	Limit     int    `help:"Maximum files to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	OrderBy   string `help:"Sort order" name:"order-by"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type SheetsGetCmd struct {
	SpreadsheetID string `arg:"" name:"spreadsheet-id" help:"Spreadsheet ID"`
}

type SheetsCreateCmd struct {
	Name   string `arg:"" name:"name" help:"Spreadsheet name"`
	Parent string `help:"Parent folder ID" name:"parent"`
}

type SheetsBatchUpdateCmd struct {
	SpreadsheetID string `arg:"" name:"spreadsheet-id" help:"Spreadsheet ID"`
	Requests      string `help:"Batch update requests JSON" name:"requests"`
	RequestsFile  string `help:"Path to JSON file with batch update requests" name:"requests-file"`
}

type SheetsValuesGetCmd struct {
	SpreadsheetID string `arg:"" name:"spreadsheet-id" help:"Spreadsheet ID"`
	Range         string `arg:"" name:"range" help:"Range (e.g., Sheet1!A1:B10)"`
}

type SheetsValuesUpdateCmd struct {
	SpreadsheetID    string `arg:"" name:"spreadsheet-id" help:"Spreadsheet ID"`
	Range            string `arg:"" name:"range" help:"Range (e.g., Sheet1!A1:B10)"`
	Values           string `help:"Values JSON (2D array)" name:"values"`
	ValuesFile       string `help:"Path to JSON file with values (2D array)" name:"values-file"`
	ValueInputOption string `help:"Value input option (RAW or USER_ENTERED)" default:"USER_ENTERED" name:"value-input-option"`
}

type SheetsValuesAppendCmd struct {
	SpreadsheetID    string `arg:"" name:"spreadsheet-id" help:"Spreadsheet ID"`
	Range            string `arg:"" name:"range" help:"Range (e.g., Sheet1!A1:B10)"`
	Values           string `help:"Values JSON (2D array)" name:"values"`
	ValuesFile       string `help:"Path to JSON file with values (2D array)" name:"values-file"`
	ValueInputOption string `help:"Value input option (RAW or USER_ENTERED)" default:"USER_ENTERED" name:"value-input-option"`
}

type SheetsValuesClearCmd struct {
	SpreadsheetID string `arg:"" name:"spreadsheet-id" help:"Spreadsheet ID"`
	Range         string `arg:"" name:"range" help:"Range (e.g., Sheet1!A1:B10)"`
}

func (cmd *SheetsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	_, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			if appErr, ok := err.(*utils.AppError); ok {
				return out.WriteError("sheets.list", appErr.CLIError)
			}
			return out.WriteError("sheets.list", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	query := fmt.Sprintf("mimeType = '%s'", utils.MimeTypeSpreadsheet)
	if parentID != "" {
		query = fmt.Sprintf("'%s' in parents and %s", parentID, query)
	}
	if cmd.Query != "" {
		query = fmt.Sprintf("%s and (%s)", query, cmd.Query)
	}

	opts := files.ListOptions{
		ParentID:       parentID,
		Query:          query,
		PageSize:       cmd.Limit,
		PageToken:      cmd.PageToken,
		OrderBy:        cmd.OrderBy,
		IncludeTrashed: false,
		Fields:         cmd.Fields,
	}

	mgr := files.NewManager(client)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		allFiles, err := mgr.ListAll(ctx, reqCtx, opts)
		if err != nil {
			return handleCLIError(out, "sheets.list", err)
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("sheets.list", allFiles)
		}
		return out.WriteSuccess("sheets.list", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.List(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "sheets.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("sheets.list", result.Files)
	}
	return out.WriteSuccess("sheets.list", result)
}

func (cmd *SheetsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	spreadsheetID, err := ResolveFileID(ctx, client, flags, cmd.SpreadsheetID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return out.WriteError("sheets.get", appErr.CLIError)
		}
		return out.WriteError("sheets.get", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := sheetsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetSpreadsheet(ctx, reqCtx, spreadsheetID)
	if err != nil {
		return handleCLIError(out, "sheets.get", err)
	}

	return out.WriteSuccess("sheets.get", result)
}

func (cmd *SheetsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	_, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			if appErr, ok := err.(*utils.AppError); ok {
				return out.WriteError("sheets.create", appErr.CLIError)
			}
			return out.WriteError("sheets.create", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	reqCtx.RequestType = types.RequestTypeMutation
	metadata := &drive.File{
		Name:     cmd.Name,
		MimeType: utils.MimeTypeSpreadsheet,
	}
	if parentID != "" {
		metadata.Parents = []string{parentID}
		reqCtx.InvolvedParentIDs = append(reqCtx.InvolvedParentIDs, parentID)
	}

	call := client.Service().Files.Create(metadata)
	call = api.NewRequestShaper(client).ShapeFilesCreate(call, reqCtx)
	call = call.Fields("id,name,mimeType,size,createdTime,modifiedTime,parents,resourceKey,trashed,capabilities")

	result, err := api.ExecuteWithRetry(ctx, client, reqCtx, func() (*drive.File, error) {
		return call.Do()
	})
	if err != nil {
		return handleCLIError(out, "sheets.create", err)
	}

	if result.ResourceKey != "" {
		client.ResourceKeys().UpdateFromAPIResponse(result.Id, result.ResourceKey)
	}

	return out.WriteSuccess("sheets.create", convertDriveFile(result))
}

func (cmd *SheetsBatchUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.batch-update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	requests, err := readSheetsBatchRequestsFrom(cmd.Requests, cmd.RequestsFile)
	if err != nil {
		return out.WriteError("sheets.batch-update", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	spreadsheetID, err := ResolveFileID(ctx, client, flags, cmd.SpreadsheetID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return out.WriteError("sheets.batch-update", appErr.CLIError)
		}
		return out.WriteError("sheets.batch-update", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := sheetsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.BatchUpdate(ctx, reqCtx, spreadsheetID, requests)
	if err != nil {
		return handleCLIError(out, "sheets.batch-update", err)
	}

	return out.WriteSuccess("sheets.batch-update", result)
}

func (cmd *SheetsValuesGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.values.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	spreadsheetID, err := ResolveFileID(ctx, client, flags, cmd.SpreadsheetID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return out.WriteError("sheets.values.get", appErr.CLIError)
		}
		return out.WriteError("sheets.values.get", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := sheetsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetValues(ctx, reqCtx, spreadsheetID, cmd.Range)
	if err != nil {
		return handleCLIError(out, "sheets.values.get", err)
	}

	return out.WriteSuccess("sheets.values.get", result)
}

func (cmd *SheetsValuesUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.values.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	values, err := readSheetValuesFrom(cmd.Values, cmd.ValuesFile)
	if err != nil {
		return out.WriteError("sheets.values.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	spreadsheetID, err := ResolveFileID(ctx, client, flags, cmd.SpreadsheetID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return out.WriteError("sheets.values.update", appErr.CLIError)
		}
		return out.WriteError("sheets.values.update", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := sheetsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UpdateValues(ctx, reqCtx, spreadsheetID, cmd.Range, values, cmd.ValueInputOption)
	if err != nil {
		return handleCLIError(out, "sheets.values.update", err)
	}

	return out.WriteSuccess("sheets.values.update", result)
}

func (cmd *SheetsValuesAppendCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.values.append", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	values, err := readSheetValuesFrom(cmd.Values, cmd.ValuesFile)
	if err != nil {
		return out.WriteError("sheets.values.append", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	spreadsheetID, err := ResolveFileID(ctx, client, flags, cmd.SpreadsheetID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return out.WriteError("sheets.values.append", appErr.CLIError)
		}
		return out.WriteError("sheets.values.append", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := sheetsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.AppendValues(ctx, reqCtx, spreadsheetID, cmd.Range, values, cmd.ValueInputOption)
	if err != nil {
		return handleCLIError(out, "sheets.values.append", err)
	}

	return out.WriteSuccess("sheets.values.append", result)
}

func (cmd *SheetsValuesClearCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if strings.TrimSpace(cmd.Range) == "" {
		return out.WriteError("sheets.values.clear", utils.NewCLIError(utils.ErrCodeInvalidArgument, "range is required").Build())
	}

	svc, client, reqCtx, err := getSheetsService(ctx, flags)
	if err != nil {
		return out.WriteError("sheets.values.clear", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	spreadsheetID, err := ResolveFileID(ctx, client, flags, cmd.SpreadsheetID)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			return out.WriteError("sheets.values.clear", appErr.CLIError)
		}
		return out.WriteError("sheets.values.clear", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := sheetsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.ClearValues(ctx, reqCtx, spreadsheetID, cmd.Range)
	if err != nil {
		return handleCLIError(out, "sheets.values.clear", err)
	}

	return out.WriteSuccess("sheets.values.clear", result)
}

func getSheetsService(ctx context.Context, flags types.GlobalFlags) (*sheetsapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceSheets); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetSheetsService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return svc, client, reqCtx, nil
}

func readSheetValuesFrom(valuesJSON, valuesFile string) ([][]interface{}, error) {
	if valuesJSON == "" && valuesFile == "" {
		return nil, fmt.Errorf("values required via --values or --values-file")
	}

	var raw []byte
	if valuesFile != "" {
		data, err := os.ReadFile(valuesFile)
		if err != nil {
			return nil, err
		}
		raw = data
	} else {
		raw = []byte(valuesJSON)
	}

	var values [][]interface{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func readSheetsBatchRequestsFrom(requestsJSON, requestsFile string) ([]*sheetsapi.Request, error) {
	if requestsJSON == "" && requestsFile == "" {
		return nil, fmt.Errorf("requests required via --requests or --requests-file")
	}

	var raw []byte
	if requestsFile != "" {
		data, err := os.ReadFile(requestsFile)
		if err != nil {
			return nil, err
		}
		raw = data
	} else {
		raw = []byte(requestsJSON)
	}

	var requests []*sheetsapi.Request
	if err := json.Unmarshal(raw, &requests); err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, fmt.Errorf("at least one request is required")
	}
	return requests, nil
}
