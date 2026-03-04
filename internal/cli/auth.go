package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/config"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type AuthCmd struct {
	Login          AuthLoginCmd          `cmd:"" help:"Authenticate with Google Drive"`
	Logout         AuthLogoutCmd         `cmd:"" help:"Remove stored credentials"`
	ServiceAccount AuthServiceAccountCmd `cmd:"service-account" help:"Authenticate with a service account"`
	Status         AuthStatusCmd         `cmd:"" help:"Show authentication status"`
	Device         AuthDeviceCmd         `cmd:"" help:"Authenticate with device code flow"`
	Profiles       AuthProfilesCmd       `cmd:"" help:"List credential profiles"`
	Diagnose       AuthDiagnoseCmd       `cmd:"" help:"Diagnose authentication configuration"`
}

type AuthLoginCmd struct {
	Scopes       []string `help:"OAuth scopes to request" name:"scopes"`
	NoBrowser    bool     `help:"Do not open a browser; use manual code entry" name:"no-browser"`
	Wide         bool     `help:"Request full Drive access scope" name:"wide"`
	Preset       string   `help:"Scope preset: workspace-basic, workspace-full, admin, workspace-with-admin, workspace-activity, workspace-labels, workspace-sync, workspace-complete, gmail, gmail-readonly, calendar, calendar-readonly, people, tasks, forms, appscript, groups, suite-complete" name:"preset"`
	ClientID     *string  `help:"OAuth client ID" name:"client-id"`
	ClientSecret *string  `help:"OAuth client secret" name:"client-secret"`
}

type AuthLogoutCmd struct{}

type AuthServiceAccountCmd struct {
	KeyFile         string   `help:"Path to service account JSON key file" name:"key-file" required:""`
	ImpersonateUser string   `help:"User email to impersonate (required for Admin SDK scopes)" name:"impersonate-user"`
	Scopes          []string `help:"OAuth scopes to request" name:"scopes"`
	Wide            bool     `help:"Request full Drive access scope" name:"wide"`
	Preset          string   `help:"Scope preset: workspace-basic, workspace-full, admin, workspace-with-admin, workspace-activity, workspace-labels, workspace-sync, workspace-complete, gmail, gmail-readonly, calendar, calendar-readonly, people, tasks, forms, appscript, groups, suite-complete" name:"preset"`
}

type AuthStatusCmd struct{}

type AuthDeviceCmd struct {
	Wide   bool   `help:"Request full Drive access scope" name:"wide"`
	Preset string `help:"Scope preset: workspace-basic, workspace-full, admin, workspace-with-admin, workspace-activity, workspace-labels, workspace-sync, workspace-complete, gmail, gmail-readonly, calendar, calendar-readonly, people, tasks, forms, appscript, groups, suite-complete" name:"preset"`
}

type AuthProfilesCmd struct{}

type AuthDiagnoseCmd struct {
	RefreshCheck bool `help:"Attempt a token refresh and report errors" name:"refresh-check"`
}

func (cmd *AuthLoginCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	configDir := getConfigDir()
	resolvedID, resolvedSecret, source, cliErr := resolveOAuthClient(cmd.ClientID, cmd.ClientSecret, configDir, false)
	if cliErr != nil {
		return out.WriteError("auth.login", cliErr.Build())
	}

	if source == oauthClientSourceBundled {
		out.Log("Using default public OAuth client credentials.")
	}

	mgr := auth.NewManager(configDir)

	// Display storage warning if any
	if warning := mgr.GetStorageWarning(); warning != "" {
		out.Log("%s", warning)
	}

	scopes, err := resolveAuthScopes(out, cmd.Preset, cmd.Wide, cmd.Scopes)
	if err != nil {
		return err
	}
	mgr.SetOAuthConfig(resolvedID, resolvedSecret, scopes)

	ctx := context.Background()
	var creds *types.Credentials
	creds, err = mgr.Authenticate(ctx, flags.Profile, openBrowser, auth.OAuthAuthOptions{
		NoBrowser: cmd.NoBrowser,
	})

	if err != nil {
		return out.WriteError("auth.login", buildAuthFlowError(err, source, resolvedID, resolvedSecret).Build())
	}

	out.Log("Successfully authenticated!")
	return out.WriteSuccess("auth.login", map[string]interface{}{
		"profile":        flags.Profile,
		"scopes":         creds.Scopes,
		"expiry":         creds.ExpiryDate.Format("2006-01-02T15:04:05Z07:00"),
		"storageBackend": mgr.GetStorageBackend(),
	})
}

func (cmd *AuthDeviceCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	configDir := getConfigDir()
	resolvedID, resolvedSecret, source, cliErr := resolveOAuthClient(nil, nil, configDir, false)
	if cliErr != nil {
		return out.WriteError("auth.device", cliErr.Build())
	}
	if source == oauthClientSourceBundled {
		out.Log("Using default public OAuth client credentials.")
	}

	mgr := auth.NewManager(configDir)

	// Display storage warning if any
	if warning := mgr.GetStorageWarning(); warning != "" {
		out.Log("%s", warning)
	}

	scopes, err := resolveAuthScopes(out, cmd.Preset, cmd.Wide, nil)
	if err != nil {
		return err
	}

	mgr.SetOAuthConfig(resolvedID, resolvedSecret, scopes)

	ctx := context.Background()
	out.Log("Using device code authentication flow...")
	creds, err := mgr.AuthenticateWithDeviceCode(ctx, flags.Profile)

	if err != nil {
		return out.WriteError("auth.device", buildAuthFlowError(err, source, resolvedID, resolvedSecret).Build())
	}

	out.Log("Successfully authenticated!")
	return out.WriteSuccess("auth.device", map[string]interface{}{
		"profile":        flags.Profile,
		"scopes":         creds.Scopes,
		"expiry":         creds.ExpiryDate.Format("2006-01-02T15:04:05Z07:00"),
		"storageBackend": mgr.GetStorageBackend(),
	})
}

func (cmd *AuthLogoutCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	configDir := getConfigDir()
	mgr := auth.NewManager(configDir)

	if err := mgr.DeleteCredentials(flags.Profile); err != nil {
		return out.WriteError("auth.logout", utils.NewCLIError(utils.ErrCodeAuthRequired,
			fmt.Sprintf("No credentials found for profile '%s'", flags.Profile)).Build())
	}

	out.Log("Credentials removed for profile: %s", flags.Profile)
	return out.WriteSuccess("auth.logout", map[string]interface{}{
		"profile": flags.Profile,
		"status":  "logged_out",
	})
}

func (cmd *AuthStatusCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	configDir := getConfigDir()
	mgr := auth.NewManager(configDir)

	creds, err := mgr.LoadCredentials(flags.Profile)
	if err != nil {
		if flags.OutputFormat == types.OutputFormatJSON {
			return out.WriteSuccess("auth.status", map[string]interface{}{
				"profile":        flags.Profile,
				"authenticated":  false,
				"storageBackend": mgr.GetStorageBackend(),
			})
		}
		// Pretty printed error for terminal
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
		fmt.Println(style.Render("✗ Not authenticated"))
		fmt.Printf("Profile: %s\n", flags.Profile)
		fmt.Printf("Storage: %s\n", mgr.GetStorageBackend())
		return nil
	}

	expired := time.Now().After(creds.ExpiryDate)
	authenticated := !expired || (creds.Type != types.AuthTypeOAuth)

	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("auth.status", map[string]interface{}{
			"profile":        flags.Profile,
			"authenticated":  authenticated,
			"scopes":         creds.Scopes,
			"expiry":         creds.ExpiryDate.Format("2006-01-02T15:04:05Z07:00"),
			"type":           creds.Type,
			"needsRefresh":   mgr.NeedsRefresh(creds),
			"expired":        expired,
			"serviceAccount": creds.ServiceAccountEmail,
			"impersonated":   creds.ImpersonatedUser,
			"storageBackend": mgr.GetStorageBackend(),
		})
	}

	// Styled terminal output
	renderStyledAuthStatus(flags.Profile, creds, authenticated, expired, mgr.NeedsRefresh(creds), mgr.GetStorageBackend())
	return nil
}

func renderStyledAuthStatus(profile string, creds *types.Credentials, authenticated, expired, needsRefresh bool, storageBackend string) {
	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f8fafc")).
		Background(lipgloss.Color("#3b82f6")).
		Padding(0, 1).
		Width(60)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8")).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e2e8f0"))

	statusStyle := lipgloss.NewStyle().
		Bold(true)

	if authenticated {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#22c55e"))
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#ef4444"))
	}

	// Status indicator
	statusIcon := "✓"
	statusText := "Active"
	if expired {
		statusIcon = "✗"
		statusText = "Expired"
	} else if needsRefresh {
		statusIcon = "◐"
		statusText = "Needs Refresh"
	}

	// Build type display
	typeDisplay := string(creds.Type)
	if creds.ServiceAccountEmail != "" {
		typeDisplay = fmt.Sprintf("%s (%s)", creds.Type, creds.ServiceAccountEmail)
	} else if creds.ImpersonatedUser != "" {
		typeDisplay = fmt.Sprintf("%s → %s", creds.Type, creds.ImpersonatedUser)
	}

	// Format expiry
	expiryDisplay := creds.ExpiryDate.Format("Jan 2, 2006 15:04")
	if creds.Type == types.AuthTypeServiceAccount || creds.Type == types.AuthTypeImpersonated {
		expiryDisplay = "Never (Service Account)"
	}

	// Build scopes summary
	scopeSummary := summarizeScopes(creds.Scopes)

	// Render
	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf(" 🔐  Authentication Status: %s", profile)))
	fmt.Println()

	// Status row
	fmt.Printf("%s %s\n",
		labelStyle.Render("Status:"),
		statusStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText)))

	// Type row
	fmt.Printf("%s %s\n",
		labelStyle.Render("Type:"),
		valueStyle.Render(typeDisplay))

	// Scopes row
	fmt.Printf("%s %s\n",
		labelStyle.Render("Scopes:"),
		valueStyle.Render(scopeSummary))

	// Expiry row
	expiryColor := lipgloss.Color("#e2e8f0")
	if expired {
		expiryColor = lipgloss.Color("#ef4444")
	} else if needsRefresh {
		expiryColor = lipgloss.Color("#f59e0b")
	}
	fmt.Printf("%s %s\n",
		labelStyle.Render("Expires:"),
		lipgloss.NewStyle().Foreground(expiryColor).Render(expiryDisplay))

	// Storage row
	fmt.Printf("%s %s\n",
		labelStyle.Render("Storage:"),
		valueStyle.Render(storageBackend))

	fmt.Println()
}

func summarizeScopes(scopes []string) string {
	if len(scopes) == 0 {
		return "None"
	}

	// Count by category
	hasDrive := false
	hasAdmin := false
	hasGmail := false
	hasCalendar := false
	hasSheets := false
	hasDocs := false

	for _, scope := range scopes {
		if strings.Contains(scope, "drive") {
			hasDrive = true
		}
		if strings.Contains(scope, "admin") {
			hasAdmin = true
		}
		if strings.Contains(scope, "gmail") {
			hasGmail = true
		}
		if strings.Contains(scope, "calendar") {
			hasCalendar = true
		}
		if strings.Contains(scope, "spreadsheets") {
			hasSheets = true
		}
		if strings.Contains(scope, "documents") {
			hasDocs = true
		}
	}

	parts := []string{}
	if hasDrive {
		parts = append(parts, "Drive")
	}
	if hasAdmin {
		parts = append(parts, "Admin")
	}
	if hasGmail {
		parts = append(parts, "Gmail")
	}
	if hasCalendar {
		parts = append(parts, "Calendar")
	}
	if hasSheets {
		parts = append(parts, "Sheets")
	}
	if hasDocs {
		parts = append(parts, "Docs")
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%d scope(s)", len(scopes))
	}

	return fmt.Sprintf("%s (%d total)", strings.Join(parts, ", "), len(scopes))
}

func getConfigDir() string {
	dir, err := config.GetConfigDir()
	if err == nil {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gdrv")
}

func (cmd *AuthProfilesCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	configDir := getConfigDir()
	mgr := auth.NewManager(configDir)

	profiles, err := mgr.ListProfiles()
	if err != nil {
		return out.WriteError("auth.profiles", utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to list profiles: %v", err)).Build())
	}

	// Get detailed info for each profile
	var profileDetails []map[string]interface{}
	styledProfiles := []styledProfile{}

	for _, profile := range profiles {
		detail := map[string]interface{}{
			"profile": profile,
		}

		creds, err := mgr.LoadCredentials(profile)
		if err == nil {
			detail["authenticated"] = true
			detail["type"] = creds.Type
			detail["expiry"] = creds.ExpiryDate.Format("2006-01-02T15:04:05Z07:00")
			detail["needsRefresh"] = mgr.NeedsRefresh(creds)
			detail["scopes"] = creds.Scopes

			expired := time.Now().After(creds.ExpiryDate)
			authenticated := !expired || creds.Type != types.AuthTypeOAuth
			styledProfiles = append(styledProfiles, styledProfile{
				name:         profile,
				profileType:  string(creds.Type),
				scopes:       summarizeScopes(creds.Scopes),
				status:       authenticated,
				expired:      expired,
				needsRefresh: mgr.NeedsRefresh(creds),
				expiry:       creds.ExpiryDate,
				serviceAcct:  creds.ServiceAccountEmail,
				impersonated: creds.ImpersonatedUser,
			})
		} else {
			detail["authenticated"] = false
			detail["error"] = err.Error()
			styledProfiles = append(styledProfiles, styledProfile{
				name:        profile,
				profileType: "unknown",
				scopes:      "-",
				status:      false,
				expired:     true,
			})
		}

		profileDetails = append(profileDetails, detail)
	}

	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("auth.profiles", map[string]interface{}{
			"profiles":       profileDetails,
			"count":          len(profiles),
			"storageBackend": mgr.GetStorageBackend(),
		})
	}

	// Styled terminal output
	renderStyledProfilesTable(styledProfiles, mgr.GetStorageBackend())
	return nil
}

type styledProfile struct {
	name         string
	profileType  string
	scopes       string
	status       bool
	expired      bool
	needsRefresh bool
	expiry       time.Time
	serviceAcct  string
	impersonated string
}

func renderStyledProfilesTable(profiles []styledProfile, storageBackend string) {
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f8fafc")).
		Background(lipgloss.Color("#3b82f6")).
		Padding(0, 1).
		Width(80)

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf(" 🔐  Authentication Profiles (%d)", len(profiles))))
	fmt.Println()

	if len(profiles) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94a3b8")).
			Italic(true)
		fmt.Println(emptyStyle.Render("No profiles found. Run 'gdrv auth login' to create one."))
		fmt.Println()
		return
	}

	// Table header
	tblHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#64748b"))

	fmt.Printf("%-12s %-12s %-25s %-12s %-15s\n",
		tblHeaderStyle.Render("Profile"),
		tblHeaderStyle.Render("Type"),
		tblHeaderStyle.Render("Scopes"),
		tblHeaderStyle.Render("Status"),
		tblHeaderStyle.Render("Expires"))
	fmt.Println(strings.Repeat("─", 80))

	// Table rows
	for _, p := range profiles {
		statusIcon := "✓"
		statusText := "Active"
		statusColor := lipgloss.Color("#22c55e")

		if p.expired {
			statusIcon = "✗"
			statusText = "Expired"
			statusColor = lipgloss.Color("#ef4444")
		} else if !p.status {
			statusIcon = "✗"
			statusText = "Invalid"
			statusColor = lipgloss.Color("#ef4444")
		} else if p.needsRefresh {
			statusIcon = "◐"
			statusText = "Refresh"
			statusColor = lipgloss.Color("#f59e0b")
		}

		// Type display
		typeDisplay := p.profileType
		if p.serviceAcct != "" {
			typeDisplay = fmt.Sprintf("svc:%s", truncate(p.serviceAcct, 12))
		} else if p.impersonated != "" {
			typeDisplay = fmt.Sprintf("imp:%s", truncate(p.impersonated, 8))
		}

		// Expiry display
		expiryDisplay := p.expiry.Format("Jan 02")
		if p.profileType == "service_account" || p.profileType == "impersonated" {
			expiryDisplay = "Never"
		} else if p.expired {
			expiryDisplay = "Expired"
		}

		statusStyle := lipgloss.NewStyle().Foreground(statusColor)
		scopesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))

		fmt.Printf("%-12s %-12s %-25s %-12s %-15s\n",
			lipgloss.NewStyle().Bold(true).Render(truncate(p.name, 11)),
			truncate(typeDisplay, 11),
			scopesStyle.Render(truncate(p.scopes, 24)),
			statusStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText)),
			expiryDisplay)
	}

	fmt.Println()
	fmt.Printf("Storage: %s\n", storageBackend)
	fmt.Println()
}

func (cmd *AuthServiceAccountCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	if cmd.KeyFile == "" {
		return fmt.Errorf("service account key file required via --key-file")
	}

	scopes, err := resolveAuthScopes(out, cmd.Preset, cmd.Wide, cmd.Scopes)
	if err != nil {
		return err
	}
	if err := validateAdminScopesRequireImpersonation(scopes, cmd.ImpersonateUser); err != nil {
		return out.WriteError("auth.service-account", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	configDir := getConfigDir()
	mgr := auth.NewManager(configDir)

	creds, err := mgr.LoadServiceAccount(context.Background(), cmd.KeyFile, scopes, cmd.ImpersonateUser)
	if err != nil {
		return out.WriteError("auth.service-account", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	if err := mgr.SaveCredentials(flags.Profile, creds); err != nil {
		return out.WriteError("auth.service-account", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	out.Log("Service account loaded")
	return out.WriteSuccess("auth.service-account", map[string]interface{}{
		"profile":        flags.Profile,
		"scopes":         creds.Scopes,
		"type":           creds.Type,
		"serviceAccount": creds.ServiceAccountEmail,
		"impersonated":   creds.ImpersonatedUser,
		"storageBackend": mgr.GetStorageBackend(),
	})
}

func resolveAuthScopes(out *OutputWriter, preset string, wide bool, commandScopes []string) ([]string, error) {
	if preset != "" {
		scopes, err := scopesForPreset(preset)
		if err != nil {
			return nil, err
		}
		out.Log("Using scope preset: %s", preset)
		return scopes, nil
	}
	if wide {
		out.Log("Using full Drive scope (%s)", utils.ScopeFull)
		return []string{utils.ScopeFull}, nil
	}
	if len(commandScopes) == 0 {
		out.Log("Using default scope preset: workspace-full")
		return utils.ScopesWorkspaceFull, nil
	}
	return commandScopes, nil
}

func scopesForPreset(preset string) ([]string, error) {
	switch preset {
	case "workspace-basic":
		return utils.ScopesWorkspaceBasic, nil
	case "workspace-full":
		return utils.ScopesWorkspaceFull, nil
	case "admin":
		return utils.ScopesAdmin, nil
	case "workspace-with-admin":
		return utils.ScopesWorkspaceWithAdmin, nil
	case "workspace-activity":
		return utils.ScopesWorkspaceActivity, nil
	case "workspace-labels":
		return utils.ScopesWorkspaceLabels, nil
	case "workspace-sync":
		return utils.ScopesWorkspaceSync, nil
	case "workspace-complete":
		return utils.ScopesWorkspaceComplete, nil
	case "gmail":
		return utils.ScopesGmail, nil
	case "gmail-readonly":
		return utils.ScopesGmailReadonly, nil
	case "calendar":
		return utils.ScopesCalendar, nil
	case "calendar-readonly":
		return utils.ScopesCalendarReadonly, nil
	case "people":
		return utils.ScopesPeople, nil
	case "tasks":
		return utils.ScopesTasks, nil
	case "forms":
		return utils.ScopesForms, nil
	case "appscript":
		return utils.ScopesAppScript, nil
	case "groups":
		return utils.ScopesGroups, nil
	case "suite-complete":
		return utils.ScopesSuiteComplete, nil
	default:
		return nil, fmt.Errorf("unknown preset: %s", preset)
	}
}

func validateAdminScopesRequireImpersonation(scopes []string, impersonateUser string) error {
	adminScopes := []string{
		utils.ScopeAdminDirectoryUser,
		utils.ScopeAdminDirectoryUserReadonly,
		utils.ScopeAdminDirectoryGroup,
		utils.ScopeAdminDirectoryGroupReadonly,
	}

	hasAdminScope := false
	for _, scope := range scopes {
		for _, adminScope := range adminScopes {
			if scope == adminScope {
				hasAdminScope = true
				break
			}
		}
		if hasAdminScope {
			break
		}
	}

	if hasAdminScope && impersonateUser == "" {
		return fmt.Errorf("admin SDK scopes require --impersonate-user")
	}

	return nil
}

func (cmd *AuthDiagnoseCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	configDir := getConfigDir()
	mgr := auth.NewManager(configDir)

	resolvedID, resolvedSecret, source, cliErr := resolveOAuthClient(nil, nil, configDir, !cmd.RefreshCheck)
	if cliErr != nil {
		return out.WriteError("auth.diagnose", cliErr.Build())
	}
	if source == oauthClientSourceBundled {
		out.Log("Using default public OAuth client credentials.")
	}
	if resolvedID != "" {
		mgr.SetOAuthConfig(resolvedID, resolvedSecret, []string{})
	}

	creds, err := mgr.LoadCredentials(flags.Profile)
	if err != nil {
		return out.WriteError("auth.diagnose", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	location, _ := mgr.CredentialLocation(flags.Profile)
	metadata, _ := mgr.LoadAuthMetadata(flags.Profile)

	clientHash := ""
	clientFingerprint := ""
	if metadata != nil {
		clientHash = metadata.ClientIDHash
		clientFingerprint = metadata.ClientIDLast4
	}

	diagnostics := map[string]interface{}{
		"profile":                     flags.Profile,
		"storageBackend":              mgr.GetStorageBackend(),
		"oauthClientSource":           source,
		"oauthClientSecretConfigured": strings.TrimSpace(resolvedSecret) != "",
		"tokenLocation":               location,
		"clientIdHash":                clientHash,
		"clientIdLast4":               clientFingerprint,
		"scopes":                      creds.Scopes,
		"expiry":                      creds.ExpiryDate.Format(time.RFC3339),
		"refreshToken":                creds.RefreshToken != "",
		"type":                        creds.Type,
		"serviceAccount":              creds.ServiceAccountEmail,
		"impersonatedUser":            creds.ImpersonatedUser,
	}

	if hint := oauthClientSecretHint(source, resolvedSecret); hint != "" {
		diagnostics["oauthClientSecretHint"] = hint
	}

	if cmd.RefreshCheck && creds.Type == types.AuthTypeOAuth {
		if mgr.GetOAuthConfig() == nil {
			return out.WriteError("auth.diagnose", utils.NewCLIError(utils.ErrCodeAuthClientMissing,
				"OAuth client credentials required for refresh check. Set GDRV_CLIENT_ID (and GDRV_CLIENT_SECRET if required) or pass --client-id/--client-secret.").Build())
		}
		_, refreshErr := mgr.RefreshCredentials(context.Background(), creds)
		if refreshErr != nil {
			var appErr *utils.AppError
			if errors.As(refreshErr, &appErr) {
				diagnostics["refreshCheck"] = map[string]interface{}{
					"success": false,
					"error":   appErr.CLIError,
				}
			} else {
				diagnostics["refreshCheck"] = map[string]interface{}{
					"success": false,
					"error":   refreshErr.Error(),
				}
			}
		} else {
			diagnostics["refreshCheck"] = map[string]interface{}{
				"success": true,
			}
		}
	}

	return out.WriteSuccess("auth.diagnose", diagnostics)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

type oauthClientSource string

const (
	oauthClientSourceFlags   oauthClientSource = "flags"
	oauthClientSourceEnv     oauthClientSource = "env"
	oauthClientSourceConfig  oauthClientSource = "config"
	oauthClientSourceBundled oauthClientSource = "bundled"
)

func buildAuthFlowError(err error, source oauthClientSource, resolvedClientID, resolvedClientSecret string) *utils.CLIErrorBuilder {
	message := err.Error()
	lower := strings.ToLower(message)

	if strings.Contains(lower, "client_secret is missing") || strings.Contains(lower, "invalid_client") {
		if hint := oauthClientSecretHint(source, resolvedClientSecret); hint != "" {
			message = fmt.Sprintf("%s\n%s", message, hint)
		}
	}

	builder := utils.NewCLIError(utils.ErrCodeAuthRequired, message)
	if source != "" {
		builder = builder.WithContext("oauthClientSource", string(source))
	}
	builder = builder.WithContext("clientIdConfigured", strings.TrimSpace(resolvedClientID) != "")
	builder = builder.WithContext("clientSecretConfigured", strings.TrimSpace(resolvedClientSecret) != "")

	return builder
}

func oauthClientSecretHint(source oauthClientSource, resolvedClientSecret string) string {
	if strings.TrimSpace(resolvedClientSecret) != "" {
		return ""
	}

	switch source {
	case oauthClientSourceFlags:
		return "The OAuth client set via flags does not include a client secret. If your client type is confidential, pass --client-secret."
	case oauthClientSourceEnv:
		return "The OAuth client set via environment variables does not include a client secret. If your client type is confidential, set GDRV_CLIENT_SECRET."
	case oauthClientSourceConfig:
		return "The OAuth client in config does not include a client secret. If your client type is confidential, set oauthClientSecret in config."
	case oauthClientSourceBundled:
		return "This build is using bundled OAuth credentials without a client secret. If Google requires a secret for this client ID, configure custom credentials with GDRV_CLIENT_ID and GDRV_CLIENT_SECRET or upgrade to a newer release."
	default:
		return ""
	}
}

func resolveOAuthClient(clientID *string, clientSecret *string, configDir string, allowMissing bool) (string, string, oauthClientSource, *utils.CLIErrorBuilder) {
	requireCustom := isTruthyEnv("GDRV_REQUIRE_CUSTOM_OAUTH")
	requireSecret := false

	flagIDSet := clientID != nil
	flagSecretSet := clientSecret != nil
	if flagIDSet || flagSecretSet {
		resolvedID := ""
		if clientID != nil {
			resolvedID = *clientID
		}
		resolvedSecret := ""
		if clientSecret != nil {
			resolvedSecret = *clientSecret
		}

		if resolvedID == "" || (requireSecret && resolvedSecret == "") {
			return "", "", "", buildOAuthClientError(utils.ErrCodeAuthClientPartial, configDir,
				"Partial OAuth client override not allowed. Set client ID (and secret if required by your client type) via flags, or clear them to use the default client if available.")
		}
		return resolvedID, resolvedSecret, oauthClientSourceFlags, nil
	}

	envID := strings.TrimSpace(os.Getenv("GDRV_CLIENT_ID"))
	envSecret := strings.TrimSpace(os.Getenv("GDRV_CLIENT_SECRET"))
	if envID != "" || envSecret != "" {
		if envID == "" || (requireSecret && envSecret == "") {
			return "", "", "", buildOAuthClientError(utils.ErrCodeAuthClientPartial, configDir,
				"Partial OAuth client override not allowed. Set client ID (and secret if required by your client type) via environment variables, or clear them to use the default client if available.")
		}
		return envID, envSecret, oauthClientSourceEnv, nil
	}

	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		return "", "", "", utils.NewCLIError(utils.ErrCodeInvalidArgument, fmt.Sprintf("Failed to load config: %v", cfgErr))
	}
	if cfg.OAuthClientID != "" || cfg.OAuthClientSecret != "" {
		if cfg.OAuthClientID == "" || (requireSecret && cfg.OAuthClientSecret == "") {
			return "", "", "", buildOAuthClientError(utils.ErrCodeAuthClientPartial, configDir,
				"Partial OAuth client override not allowed. Set client ID (and secret if required by your client type) in config, or remove them to use the default client if available.")
		}
		return cfg.OAuthClientID, cfg.OAuthClientSecret, oauthClientSourceConfig, nil
	}

	if requireCustom && !allowMissing {
		return "", "", "", buildOAuthClientError(utils.ErrCodeAuthClientMissing, configDir,
			"Custom OAuth client required. Set GDRV_CLIENT_ID (and GDRV_CLIENT_SECRET if required) or configure the client in the config file. Default credentials are disabled by GDRV_REQUIRE_CUSTOM_OAUTH.")
	}

	if bundledID, bundledSecret, ok := auth.GetBundledOAuthClient(); ok {
		if requireCustom {
			return "", "", "", buildOAuthClientError(utils.ErrCodeAuthClientMissing, configDir,
				"Custom OAuth client required. Set GDRV_CLIENT_ID (and GDRV_CLIENT_SECRET if required) or configure the client in the config file. Default credentials are disabled by GDRV_REQUIRE_CUSTOM_OAUTH.")
		}
		return bundledID, bundledSecret, oauthClientSourceBundled, nil
	}

	if allowMissing {
		return "", "", "", nil
	}

	return "", "", "", buildOAuthClientError(utils.ErrCodeAuthClientMissing, configDir,
		"OAuth client ID missing. Default credentials are not available in this build. Provide a custom client ID via environment variables or config.")
}

func buildOAuthClientError(code, configDir, message string) *utils.CLIErrorBuilder {
	configPath, err := config.GetConfigPath()
	if err != nil {
		configPath = filepath.Join(configDir, config.ConfigFileName)
	}
	tokenHint := filepath.Join(configDir, "credentials")

	fullMessage := fmt.Sprintf(
		"%s\nConfig path: %s\nToken storage: system keyring (preferred) or %s\nUse --no-browser for manual login when running headless.",
		message,
		configPath,
		tokenHint,
	)

	return utils.NewCLIError(code, fullMessage).
		WithContext("configPath", configPath).
		WithContext("tokenLocation", tokenHint)
}

func isTruthyEnv(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
