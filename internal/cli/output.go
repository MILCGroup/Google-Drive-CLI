package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// OutputWriter handles CLI output formatting
type OutputWriter struct {
	format   types.OutputFormat
	quiet    bool
	verbose  bool
	warnings []types.CLIWarning
}

// NewOutputWriter creates a new output writer
func NewOutputWriter(format types.OutputFormat, quiet, verbose bool) *OutputWriter {
	return &OutputWriter{
		format:   format,
		quiet:    quiet,
		verbose:  verbose,
		warnings: []types.CLIWarning{},
	}
}

// AddWarning adds a warning to the output
func (w *OutputWriter) AddWarning(code, message, severity string) {
	w.warnings = append(w.warnings, types.CLIWarning{
		Code:     code,
		Message:  message,
		Severity: severity,
	})
}

// WriteSuccess writes a successful result
func (w *OutputWriter) WriteSuccess(command string, data interface{}) error {
	output := types.CLIOutput{
		SchemaVersion: utils.SchemaVersion,
		TraceID:       uuid.New().String(),
		Command:       command,
		Data:          data,
		Warnings:      w.warnings,
		Errors:        []types.CLIError{},
	}

	if w.format == types.OutputFormatJSON {
		return w.writeJSON(output)
	}
	return w.writeTable(data)
}

// WriteError writes an error result
func (w *OutputWriter) WriteError(command string, cliErr types.CLIError) error {
	output := types.CLIOutput{
		SchemaVersion: utils.SchemaVersion,
		TraceID:       uuid.New().String(),
		Command:       command,
		Data:          nil,
		Warnings:      w.warnings,
		Errors:        []types.CLIError{cliErr},
	}

	return w.writeJSON(output)
}

func (w *OutputWriter) writeJSON(output types.CLIOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func (w *OutputWriter) writeTable(data interface{}) error {
	if renderable, ok := data.(types.TableRenderable); ok {
		return w.renderTable(renderable.AsTableRenderer())
	}
	if renderer, ok := data.(types.TableRenderer); ok {
		return w.renderTable(renderer)
	}
	switch v := data.(type) {
	case []*types.DriveFile:
		return w.writeFileTable(v)
	case *types.DriveFile:
		return w.writeFileTable([]*types.DriveFile{v})
	case []*types.Permission:
		return w.writePermissionTable(v)
	default:
		// Fallback to JSON for unknown types
		return w.writeJSON(types.CLIOutput{
			SchemaVersion: utils.SchemaVersion,
			TraceID:       uuid.New().String(),
			Command:       "unknown",
			Data:          data,
			Warnings:      w.warnings,
			Errors:        []types.CLIError{},
		})
	}
}

func (w *OutputWriter) renderTable(renderer types.TableRenderer) error {
	rows := renderer.Rows()
	if len(rows) == 0 {
		if !w.quiet {
			if _, err := fmt.Fprintln(os.Stdout, renderer.EmptyMessage()); err != nil {
				return err
			}
		}
		return nil
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Configure(func(config *tablewriter.Config) {
		config.Header.Alignment.Global = tw.AlignLeft
		config.Row.Alignment.Global = tw.AlignLeft
	})
	table.Header(renderer.Headers())
	if err := table.Bulk(rows); err != nil {
		return err
	}
	return table.Render()
}

func (w *OutputWriter) writeFileTable(files []*types.DriveFile) error {
	table := tablewriter.NewTable(os.Stdout)
	table.Configure(func(config *tablewriter.Config) {
		config.Header.Alignment.Global = tw.AlignLeft
		config.Row.Alignment.Global = tw.AlignLeft
	})

	data := make([][]interface{}, 0, len(files))
	for _, f := range files {
		size := "-"
		if f.Size > 0 {
			size = formatSize(f.Size)
		}
		data = append(data, []interface{}{
			truncate(f.ID, 15),
			truncate(f.Name, 40),
			truncate(f.MimeType, 30),
			size,
			f.ModifiedTime,
		})
	}

	table.Header([]string{"ID", "Name", "Type", "Size", "Modified"})
	if err := table.Bulk(data); err != nil {
		return err
	}
	return table.Render()
}

func (w *OutputWriter) writePermissionTable(perms []*types.Permission) error {
	table := tablewriter.NewTable(os.Stdout)
	table.Configure(func(config *tablewriter.Config) {
		config.Header.Alignment.Global = tw.AlignLeft
		config.Row.Alignment.Global = tw.AlignLeft
	})

	data := make([][]interface{}, 0, len(perms))
	for _, p := range perms {
		identity := p.EmailAddress
		if identity == "" {
			identity = p.Domain
		}
		if identity == "" {
			identity = "-"
		}
		data = append(data, []interface{}{p.ID, p.Type, p.Role, identity})
	}

	table.Header([]string{"ID", "Type", "Role", "Email/Domain"})
	if err := table.Bulk(data); err != nil {
		return err
	}
	return table.Render()
}

// handleCLIError converts an error into a structured CLI error output.
// It checks whether the error is an AppError (with a structured CLIError)
// and writes it directly; otherwise it wraps it as an unknown error.
// This replaces the repeated 4-line error handling pattern across all CLI commands.
func handleCLIError(w *OutputWriter, command string, err error) error {
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return w.WriteError(command, appErr.CLIError)
	}
	return w.WriteError(command, utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
}

// Log writes to stderr if not quiet
func (w *OutputWriter) Log(format string, args ...interface{}) {
	if !w.quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Verbose writes to stderr if verbose is enabled
func (w *OutputWriter) Verbose(format string, args ...interface{}) {
	if w.verbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
