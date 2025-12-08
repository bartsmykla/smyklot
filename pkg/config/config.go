// Package config provides configuration management for Smyklot using Viper
//
// It supports loading configuration from multiple sources with precedence:
// CLI flags > Environment variables > Config file > Defaults
package config

import (
	"encoding/json"
	"os"

	"github.com/spf13/viper"
)

const (
	// KeyQuietSuccess is the config key for quiet_success setting
	KeyQuietSuccess = "quiet_success"

	// KeyQuietReactions is the config key for quiet_reactions setting
	KeyQuietReactions = "quiet_reactions"

	// KeyQuietPending is the config key for quiet_pending setting
	KeyQuietPending = "quiet_pending"

	// KeyAllowedCommands is the config key for allowed_commands setting
	KeyAllowedCommands = "allowed_commands"

	// KeyCommandAliases is the config key for command_aliases setting
	KeyCommandAliases = "command_aliases"

	// KeyCommandPrefix is the config key for command_prefix setting
	KeyCommandPrefix = "command_prefix"

	// KeyDisableMentions is the config key for disable_mentions setting
	KeyDisableMentions = "disable_mentions"

	// KeyDisableBareCommands is the config key for disable_bare_commands setting
	KeyDisableBareCommands = "disable_bare_commands"

	// KeyDisableUnapprove is the config key for disable_unapprove setting
	KeyDisableUnapprove = "disable_unapprove"

	// KeyDisableReactions is the config key for disable_reactions setting
	KeyDisableReactions = "disable_reactions"

	// KeyDisableDeletedComments is the config key for disable_deleted_comments setting
	KeyDisableDeletedComments = "disable_deleted_comments"

	// KeyAllowSelfApproval is the config key for allow_self_approval setting
	KeyAllowSelfApproval = "allow_self_approval"

	// EnvPrefix is the prefix for environment variables
	EnvPrefix = "SMYKLOT"

	// EnvConfig is the environment variable for JSON configuration
	EnvConfig = "SMYKLOT_CONFIG"
)

// SetupViper configures Viper with default values and environment variable bindings
func SetupViper(v *viper.Viper) {
	// Set defaults
	v.SetDefault(KeyQuietSuccess, false)
	v.SetDefault(KeyQuietReactions, false)
	v.SetDefault(KeyQuietPending, false)
	v.SetDefault(KeyAllowedCommands, []string{})
	v.SetDefault(KeyCommandAliases, map[string]string{})
	v.SetDefault(KeyCommandPrefix, DefaultCommandPrefix)
	v.SetDefault(KeyDisableMentions, false)
	v.SetDefault(KeyDisableBareCommands, false)
	v.SetDefault(KeyDisableUnapprove, false)
	v.SetDefault(KeyDisableReactions, false)
	v.SetDefault(KeyDisableDeletedComments, false)
	v.SetDefault(KeyAllowSelfApproval, false)

	// Enable environment variable support
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
}

// LoadFromViper creates a Config from Viper settings
func LoadFromViper(v *viper.Viper) *Config {
	return &Config{
		QuietSuccess:           v.GetBool(KeyQuietSuccess),
		QuietReactions:         v.GetBool(KeyQuietReactions),
		QuietPending:           v.GetBool(KeyQuietPending),
		AllowedCommands:        v.GetStringSlice(KeyAllowedCommands),
		CommandAliases:         v.GetStringMapString(KeyCommandAliases),
		CommandPrefix:          v.GetString(KeyCommandPrefix),
		DisableMentions:        v.GetBool(KeyDisableMentions),
		DisableBareCommands:    v.GetBool(KeyDisableBareCommands),
		DisableUnapprove:       v.GetBool(KeyDisableUnapprove),
		DisableReactions:       v.GetBool(KeyDisableReactions),
		DisableDeletedComments: v.GetBool(KeyDisableDeletedComments),
		AllowSelfApproval:      v.GetBool(KeyAllowSelfApproval),
	}
}

// LoadJSONConfig reads and parses JSON configuration from SMYKLOT_CONFIG environment variable
func LoadJSONConfig(v *viper.Viper) error {
	configJSON := os.Getenv(EnvConfig)
	if configJSON == "" {
		return nil // No JSON config provided
	}

	// Parse JSON into a map
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &configMap); err != nil {
		return err
	}

	// Merge each config value into Viper
	for key, value := range configMap {
		// Viper expects snake_case keys
		v.Set(key, value)
	}

	return nil
}
