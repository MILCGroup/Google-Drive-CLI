package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/milcgroup/gdrv/internal/config"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
)

type ConfigCmd struct {
	Show  ConfigShowCmd  `cmd:"" help:"Show current config"`
	Set   ConfigSetCmd   `cmd:"" help:"Set config value"`
	Reset ConfigResetCmd `cmd:"" help:"Reset to defaults"`
}

type ConfigShowCmd struct {
}

type ConfigSetCmd struct {
	Key   string `arg:"" name:"key" help:"Configuration key"`
	Value string `arg:"" name:"value" help:"Configuration value"`
}

type ConfigResetCmd struct {
}

func (cmd *ConfigShowCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	cfg, err := config.Load()
	if err != nil {
		return out.WriteError("config.show", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	return out.WriteSuccess("config.show", cfg)
}

func (cmd *ConfigSetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	cfg, err := config.Load()
	if err != nil {
		return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	// Set the value based on key
	switch strings.ToLower(cmd.Key) {
	case "defaultprofile":
		cfg.DefaultProfile = cmd.Value
	case "defaultoutputformat":
		if cmd.Value != string(types.OutputFormatJSON) && cmd.Value != string(types.OutputFormatTable) {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				"Invalid output format. Must be 'json' or 'table'").Build())
		}
		cfg.DefaultOutputFormat = types.OutputFormat(cmd.Value)
	case "defaultfields":
		preset := config.FieldMaskPreset(cmd.Value)
		if preset != config.FieldMaskMinimal && preset != config.FieldMaskStandard && preset != config.FieldMaskFull {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				"Invalid field mask preset. Must be 'minimal', 'standard', or 'full'").Build())
		}
		cfg.DefaultFields = preset
	case "cachettl":
		ttl, err := strconv.Atoi(cmd.Value)
		if err != nil || ttl < 0 {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				"Cache TTL must be a non-negative integer").Build())
		}
		cfg.CacheTTL = ttl
	case "includeexportlinks":
		cfg.IncludeExportLinks = parseBool(cmd.Value)
	case "maxretries":
		retries, err := strconv.Atoi(cmd.Value)
		if err != nil || retries < 0 || retries > 10 {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				"Max retries must be between 0 and 10").Build())
		}
		cfg.MaxRetries = retries
	case "retrybasedelay":
		delay, err := strconv.Atoi(cmd.Value)
		if err != nil || delay < 100 || delay > 60000 {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				"Retry base delay must be between 100 and 60000 ms").Build())
		}
		cfg.RetryBaseDelay = delay
	case "requesttimeout":
		timeout, err := strconv.Atoi(cmd.Value)
		if err != nil || timeout < 1 || timeout > 3600 {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				"Request timeout must be between 1 and 3600 seconds").Build())
		}
		cfg.RequestTimeout = timeout
	case "loglevel":
		validLevels := []string{"quiet", "normal", "verbose", "debug"}
		valid := false
		for _, level := range validLevels {
			if cmd.Value == level {
				valid = true
				break
			}
		}
		if !valid {
			return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				fmt.Sprintf("Invalid log level. Must be one of: %s", strings.Join(validLevels, ", "))).Build())
		}
		cfg.LogLevel = cmd.Value
	case "coloroutput":
		cfg.ColorOutput = parseBool(cmd.Value)
	case "oauthclientid":
		cfg.OAuthClientID = cmd.Value
	case "oauthclientsecret":
		cfg.OAuthClientSecret = cmd.Value
	default:
		return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			fmt.Sprintf("Unknown configuration key: %s", cmd.Key)).Build())
	}

	// Save the configuration
	if err := cfg.Save(); err != nil {
		return out.WriteError("config.set", utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to save configuration: %v", err)).Build())
	}

	out.Log("Configuration updated: %s = %s", cmd.Key, cmd.Value)
	return out.WriteSuccess("config.set", map[string]interface{}{
		"key":   cmd.Key,
		"value": cmd.Value,
	})
}

func (cmd *ConfigResetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	cfg := config.DefaultConfig()
	if err := cfg.Save(); err != nil {
		return out.WriteError("config.reset", utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to reset configuration: %v", err)).Build())
	}

	out.Log("Configuration reset to defaults")
	return out.WriteSuccess("config.reset", cfg)
}

// parseBool parses a boolean value from a string
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on"
}
