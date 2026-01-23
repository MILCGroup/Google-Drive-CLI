// Package safety provides safety controls for destructive operations including
// dry-run mode, confirmation requirements, and idempotency support.
package safety

// SafetyOptions configures safety controls for destructive operations.
//
// DryRun mode allows previewing changes without executing them.
// Force and Yes flags control confirmation behavior for automation.
// Quiet and Interactive flags control output verbosity.
//
// Requirements:
//   - Requirement 13.1: Support --dry-run mode for destructive operations
//   - Requirement 13.2: Support --force flag to skip confirmations
//   - Requirement 13.5: Support --yes flag for non-interactive runs
type SafetyOptions struct {
	// DryRun enables preview mode - operations are planned but not executed.
	// All operations will log what they would do without making actual changes.
	DryRun bool

	// Force skips all confirmation prompts.
	// Use with caution - bypasses all safety checks.
	Force bool

	// Yes automatically confirms all prompts without user interaction.
	// Useful for scripting and automation.
	Yes bool

	// Quiet suppresses confirmation output and progress messages.
	// Only errors and final results are displayed.
	Quiet bool

	// Interactive enables interactive mode with detailed prompts.
	// When false, operations fail if confirmation is required but not provided.
	Interactive bool
}

// Default returns a SafetyOptions with defaults for CLI use.
// By default, commands are non-interactive (no prompts) for agent-friendliness.
// Use --dry-run to preview operations, or Interactive() for interactive mode.
func Default() SafetyOptions {
	return SafetyOptions{
		DryRun:      false,
		Force:       false,
		Yes:         true, // Auto-confirm by default for agent-friendly CLI
		Quiet:       false,
		Interactive: false,
	}
}

// NonInteractive returns a SafetyOptions configured for strict non-interactive use.
// Operations will fail if confirmation is required and Yes is not set.
// Use this when you want to ensure no auto-confirmation happens.
func NonInteractive() SafetyOptions {
	return SafetyOptions{
		DryRun:      false,
		Force:       false,
		Yes:         false,
		Quiet:       false,
		Interactive: false,
	}
}

// Interactive returns a SafetyOptions configured for interactive use with prompts.
// Operations will prompt for confirmation before executing.
func Interactive() SafetyOptions {
	return SafetyOptions{
		DryRun:      false,
		Force:       false,
		Yes:         false,
		Quiet:       false,
		Interactive: true,
	}
}

// DryRunMode returns a SafetyOptions configured for dry-run mode.
// Operations are previewed but not executed.
func DryRunMode() SafetyOptions {
	return SafetyOptions{
		DryRun:      true,
		Force:       false,
		Yes:         false,
		Quiet:       false,
		Interactive: true,
	}
}

// ShouldConfirm returns true if the operation should request user confirmation.
// Returns false if Force or Yes flags are set, or if in DryRun mode.
func (o SafetyOptions) ShouldConfirm() bool {
	return !o.Force && !o.Yes && !o.DryRun && o.Interactive
}

// AutoConfirm returns true if operations should be automatically confirmed.
// Returns true if Force, Yes, or DryRun flags are set.
func (o SafetyOptions) AutoConfirm() bool {
	return o.Force || o.Yes || o.DryRun
}

// ShouldExecute returns true if operations should be executed.
// Returns false only in DryRun mode.
func (o SafetyOptions) ShouldExecute() bool {
	return !o.DryRun
}
