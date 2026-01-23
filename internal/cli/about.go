package cli

import (
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Display Drive account information and API capabilities",
	Long:  "Retrieve and display information about the authenticated Drive account and supported API capabilities",
	RunE:  runAbout,
}

var aboutFields string

func init() {
	aboutCmd.Flags().StringVar(&aboutFields, "fields", "*", "Fields to retrieve")
	rootCmd.AddCommand(aboutCmd)
}

func runAbout(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	capabilities := map[string]interface{}{
		"version": "1.0.0",
		"api": map[string]interface{}{
			"supported_operations": []string{
				"files.list", "files.get", "files.upload", "files.download", "files.delete",
				"files.copy", "files.move", "files.trash", "files.restore", "files.revisions",
				"folders.create", "folders.list", "folders.delete", "folders.move",
				"permissions.list", "permissions.create", "permissions.update", "permissions.delete", "permissions.public",
				"drives.list", "drives.get",
				"auth.login", "auth.device", "auth.status", "auth.profiles", "auth.logout",
			},
			"supported_exports": []string{
				"pdf", "docx", "xlsx", "pptx", "txt", "html", "rtf", "csv",
			},
			"features": []string{
				"batch_operations", "path_resolution", "caching", "dry_run", "safety_checks",
				"shared_drives", "permissions", "revisions", "trash_management",
			},
		},
		"authentication": map[string]interface{}{
			"oauth2_flows": []string{"web", "device_code"},
			"scopes": []string{
				"https://www.googleapis.com/auth/drive",
				"https://www.googleapis.com/auth/drive.file",
				"https://www.googleapis.com/auth/drive.metadata",
			},
		},
		"output_formats": []string{"json", "table"},
		"configuration": map[string]interface{}{
			"config_file":     "~/.gdrive/config.json",
			"credentials_dir": "~/.gdrive/credentials",
			"cache_dir":       "~/.gdrive/cache",
		},
	}

	return out.WriteSuccess("about", capabilities)
}
