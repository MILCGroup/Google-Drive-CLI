package safety

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/utils"
)

// Confirm prompts the user for confirmation with a yes/no question.
// Returns true if the user confirms, false otherwise.
//
// If opts.Force or opts.Yes is set, returns true without prompting.
// If opts.DryRun is set, returns true without prompting (dry-run always proceeds).
// If opts.Interactive is false and no auto-confirm flag is set, returns false.
//
// Requirements:
//   - Requirement 13.2: Support --force flag to skip confirmations
//   - Requirement 13.3: Implement confirmation requirements for bulk operations
//   - Requirement 13.5: Support --yes flag for non-interactive runs
func Confirm(message string, opts SafetyOptions) (bool, error) {
	// Auto-confirm if flags are set
	if opts.AutoConfirm() {
		if !opts.Quiet && !opts.DryRun {
			fmt.Printf("%s [auto-confirmed]\n", message)
		}
		return true, nil
	}

	// Non-interactive mode without auto-confirm flag
	if !opts.Interactive {
		return false, utils.NewAppError(utils.NewCLIError(
			utils.ErrCodeInvalidArgument,
			"Confirmation required but running in non-interactive mode. Use --yes or --force to proceed.",
		).Build())
	}

	// Interactive prompt
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// ConfirmBulkOperation prompts for confirmation of a bulk operation affecting multiple items.
// Returns true if the user confirms, false otherwise.
//
// Parameters:
//   - itemCount: Number of items that will be affected
//   - operation: Description of the operation (e.g., "delete", "update permissions")
//   - opts: Safety options controlling confirmation behavior
//
// Requirements:
//   - Requirement 13.3: Implement confirmation requirements for bulk operations
func ConfirmBulkOperation(itemCount int, operation string, opts SafetyOptions) (bool, error) {
	if itemCount == 0 {
		return false, nil
	}

	if itemCount == 1 {
		return Confirm(fmt.Sprintf("About to %s 1 item. Continue?", operation), opts)
	}

	return Confirm(fmt.Sprintf("About to %s %d items. Continue?", operation, itemCount), opts)
}

// ConfirmDestructive prompts for confirmation of a destructive operation with item details.
// Lists the items that will be affected and asks for confirmation.
// Returns true if the user confirms, false otherwise.
//
// Parameters:
//   - items: List of item names or identifiers to be affected
//   - operation: Description of the operation (e.g., "permanently delete", "trash")
//   - opts: Safety options controlling confirmation behavior
//
// If the list is very long (>10 items), shows first 10 and indicates remaining count.
//
// Requirements:
//   - Requirement 13.3: Implement confirmation requirements for bulk operations
func ConfirmDestructive(items []string, operation string, opts SafetyOptions) (bool, error) {
	if len(items) == 0 {
		return false, nil
	}

	// Auto-confirm if flags are set
	if opts.AutoConfirm() {
		if !opts.Quiet && !opts.DryRun {
			fmt.Printf("About to %s %d item(s) [auto-confirmed]\n", operation, len(items))
		}
		return true, nil
	}

	// Non-interactive mode without auto-confirm flag
	if !opts.Interactive {
		return false, utils.NewAppError(utils.NewCLIError(
			utils.ErrCodeInvalidArgument,
			"Confirmation required but running in non-interactive mode. Use --yes or --force to proceed.",
		).Build())
	}

	// Display items to be affected
	fmt.Printf("\n⚠️  WARNING: About to %s the following items:\n\n", operation)

	displayCount := len(items)
	if displayCount > 10 {
		displayCount = 10
	}

	for i := 0; i < displayCount; i++ {
		fmt.Printf("  - %s\n", items[i])
	}

	if len(items) > displayCount {
		fmt.Printf("  ... and %d more items\n", len(items)-displayCount)
	}

	fmt.Printf("\nTotal: %d item(s)\n\n", len(items))

	// Interactive prompt
	fmt.Printf("This operation cannot be undone. Continue? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// ConfirmWithDefault prompts for confirmation with a default value.
// If the user presses Enter without typing anything, returns the default value.
//
// Parameters:
//   - message: The confirmation message
//   - defaultValue: The default value if user presses Enter
//   - opts: Safety options controlling confirmation behavior
func ConfirmWithDefault(message string, defaultValue bool, opts SafetyOptions) (bool, error) {
	// Auto-confirm if flags are set
	if opts.AutoConfirm() {
		if !opts.Quiet && !opts.DryRun {
			fmt.Printf("%s [auto-confirmed]\n", message)
		}
		return true, nil
	}

	// Non-interactive mode without auto-confirm flag
	if !opts.Interactive {
		return false, utils.NewAppError(utils.NewCLIError(
			utils.ErrCodeInvalidArgument,
			"Confirmation required but running in non-interactive mode. Use --yes or --force to proceed.",
		).Build())
	}

	// Interactive prompt with default indicator
	defaultIndicator := "[y/N]"
	if defaultValue {
		defaultIndicator = "[Y/n]"
	}

	fmt.Printf("%s %s: ", message, defaultIndicator)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))

	// Empty response uses default
	if response == "" {
		return defaultValue, nil
	}

	return response == "y" || response == "yes", nil
}
