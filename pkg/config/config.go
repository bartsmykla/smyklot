// Package config provides configuration management for Smyklot using Viper
//
// It supports loading configuration from multiple sources with precedence:
// CLI flags > Environment variables > Config file > Defaults
package config

import (
	"github.com/spf13/viper"
)

const (
	// KeyQuietSuccess is the config key for quiet_success setting
	KeyQuietSuccess = "quiet_success"

	// KeyQuietReactions is the config key for quiet_reactions setting
	KeyQuietReactions = "quiet_reactions"

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

	// EnvPrefix is the prefix for environment variables
	EnvPrefix = "SMYKLOT"
)

// SetupViper configures Viper with default values and environment variable bindings
func SetupViper(v *viper.Viper) {
	// Set defaults
	v.SetDefault(KeyQuietSuccess, false)
	v.SetDefault(KeyQuietReactions, false)
	v.SetDefault(KeyAllowedCommands, []string{})
	v.SetDefault(KeyCommandAliases, map[string]string{})
	v.SetDefault(KeyCommandPrefix, DefaultCommandPrefix)
	v.SetDefault(KeyDisableMentions, false)
	v.SetDefault(KeyDisableBareCommands, false)
	v.SetDefault(KeyDisableUnapprove, false)
	v.SetDefault(KeyDisableReactions, false)

	// Enable environment variable support
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
}

// LoadFromViper creates a Config from Viper settings
func LoadFromViper(v *viper.Viper) *Config {
	return &Config{
		QuietSuccess:        v.GetBool(KeyQuietSuccess),
		QuietReactions:      v.GetBool(KeyQuietReactions),
		AllowedCommands:     v.GetStringSlice(KeyAllowedCommands),
		CommandAliases:      v.GetStringMapString(KeyCommandAliases),
		CommandPrefix:       v.GetString(KeyCommandPrefix),
		DisableMentions:     v.GetBool(KeyDisableMentions),
		DisableBareCommands: v.GetBool(KeyDisableBareCommands),
		DisableUnapprove:    v.GetBool(KeyDisableUnapprove),
		DisableReactions:    v.GetBool(KeyDisableReactions),
	}
}
