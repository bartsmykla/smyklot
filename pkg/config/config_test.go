package config_test

import (
	"bytes"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/bartsmykla/smyklot/pkg/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config [Unit]", func() {
	Describe("Default", func() {
		It("should return default config values", func() {
			cfg := config.Default()

			Expect(cfg.QuietSuccess).To(BeFalse())
			Expect(cfg.AllowedCommands).To(BeEmpty())
			Expect(cfg.CommandAliases).To(BeEmpty())
			Expect(cfg.CommandPrefix).To(Equal("/"))
			Expect(cfg.DisableMentions).To(BeFalse())
		})
	})

	Describe("SetupViper", func() {
		var v *viper.Viper

		BeforeEach(func() {
			v = viper.New()
			config.SetupViper(v)
		})

		It("should set default values", func() {
			Expect(v.GetBool(config.KeyQuietSuccess)).To(BeFalse())
			Expect(v.GetStringSlice(config.KeyAllowedCommands)).To(BeEmpty())
			Expect(v.GetStringMapString(config.KeyCommandAliases)).To(BeEmpty())
			Expect(v.GetString(config.KeyCommandPrefix)).To(Equal("/"))
			Expect(v.GetBool(config.KeyDisableMentions)).To(BeFalse())
		})

		It("should configure environment variable prefix", func() {
			// This is tested indirectly through LoadFromViper with env vars
			Expect(v).NotTo(BeNil())
		})
	})

	Describe("LoadFromViper", func() {
		var v *viper.Viper

		BeforeEach(func() {
			v = viper.New()
			config.SetupViper(v)
		})

		It("should load default values", func() {
			cfg := config.LoadFromViper(v)

			Expect(cfg.QuietSuccess).To(BeFalse())
			Expect(cfg.AllowedCommands).To(BeEmpty())
			Expect(cfg.CommandAliases).To(BeEmpty())
			Expect(cfg.CommandPrefix).To(Equal("/"))
			Expect(cfg.DisableMentions).To(BeFalse())
		})

		It("should load from explicit settings", func() {
			v.Set(config.KeyQuietSuccess, true)
			v.Set(config.KeyAllowedCommands, []string{"approve", "merge"})
			v.Set(config.KeyCommandAliases, map[string]string{"app": "approve"})
			v.Set(config.KeyCommandPrefix, "!")
			v.Set(config.KeyDisableMentions, true)

			cfg := config.LoadFromViper(v)

			Expect(cfg.QuietSuccess).To(BeTrue())
			Expect(cfg.AllowedCommands).To(ConsistOf("approve", "merge"))
			Expect(cfg.CommandAliases).To(HaveKeyWithValue("app", "approve"))
			Expect(cfg.CommandPrefix).To(Equal("!"))
			Expect(cfg.DisableMentions).To(BeTrue())
		})

		It("should load from environment variables", func() {
			Expect(os.Setenv("SMYKLOT_QUIET_SUCCESS", "true")).To(Succeed())
			Expect(os.Setenv("SMYKLOT_COMMAND_PREFIX", "!")).To(Succeed())
			Expect(os.Setenv("SMYKLOT_DISABLE_MENTIONS", "true")).To(Succeed())

			defer func() {
				_ = os.Unsetenv("SMYKLOT_QUIET_SUCCESS")
				_ = os.Unsetenv("SMYKLOT_COMMAND_PREFIX")
				_ = os.Unsetenv("SMYKLOT_DISABLE_MENTIONS")
			}()

			// Create a new viper instance to pick up env vars
			v = viper.New()
			config.SetupViper(v)

			cfg := config.LoadFromViper(v)

			Expect(cfg.QuietSuccess).To(BeTrue())
			Expect(cfg.CommandPrefix).To(Equal("!"))
			Expect(cfg.DisableMentions).To(BeTrue())
		})

		It("should load from JSON config", func() {
			jsonConfig := `{
				"quiet_success": true,
				"allowed_commands": ["approve", "merge"],
				"command_aliases": {"app": "approve", "a": "approve"},
				"command_prefix": "!",
				"disable_mentions": true
			}`

			v.SetConfigType("json")

			err := v.ReadConfig(bytes.NewReader([]byte(jsonConfig)))
			Expect(err).NotTo(HaveOccurred())

			cfg := config.LoadFromViper(v)

			Expect(cfg.QuietSuccess).To(BeTrue())
			Expect(cfg.AllowedCommands).To(ConsistOf("approve", "merge"))
			Expect(cfg.CommandAliases).To(HaveKeyWithValue("app", "approve"))
			Expect(cfg.CommandAliases).To(HaveKeyWithValue("a", "approve"))
			Expect(cfg.CommandPrefix).To(Equal("!"))
			Expect(cfg.DisableMentions).To(BeTrue())
		})

		It("should respect precedence: explicit > env > config > default", func() {
			// Set the config file
			jsonConfig := `{
				"quiet_success": true,
				"command_prefix": "!",
				"disable_mentions": false
			}`

			v.SetConfigType("json")

			err := v.ReadConfig(bytes.NewReader([]byte(jsonConfig)))
			Expect(err).NotTo(HaveOccurred())

			// Set env var (should override config)
			Expect(os.Setenv("SMYKLOT_QUIET_SUCCESS", "false")).To(Succeed())
			Expect(os.Setenv("SMYKLOT_DISABLE_MENTIONS", "true")).To(Succeed())

			defer func() {
				_ = os.Unsetenv("SMYKLOT_QUIET_SUCCESS")
				_ = os.Unsetenv("SMYKLOT_DISABLE_MENTIONS")
			}()

			// Create a new viper to pick up env
			v = viper.New()
			config.SetupViper(v)

			v.SetConfigType("json")
			err = v.ReadConfig(bytes.NewReader([]byte(jsonConfig)))
			Expect(err).NotTo(HaveOccurred())

			// Set an explicit value (should override everything)
			v.Set(config.KeyQuietSuccess, true)

			cfg := config.LoadFromViper(v)

			// Explicit overrides env and config
			Expect(cfg.QuietSuccess).To(BeTrue())

			// Env overrides config
			Expect(cfg.DisableMentions).To(BeTrue())

			// Config provides value
			Expect(cfg.CommandPrefix).To(Equal("!"))
		})
	})

	Describe("IsCommandAllowed", func() {
		It("should allow all commands when AllowedCommands is empty", func() {
			cfg := config.Default()

			Expect(cfg.IsCommandAllowed("approve")).To(BeTrue())
			Expect(cfg.IsCommandAllowed("merge")).To(BeTrue())
			Expect(cfg.IsCommandAllowed("anything")).To(BeTrue())
		})

		It("should only allow specified commands", func() {
			cfg := &config.Config{
				AllowedCommands: []string{"approve", "merge"},
			}

			Expect(cfg.IsCommandAllowed("approve")).To(BeTrue())
			Expect(cfg.IsCommandAllowed("merge")).To(BeTrue())
			Expect(cfg.IsCommandAllowed("close")).To(BeFalse())
		})
	})

	Describe("ResolveAlias", func() {
		It("should resolve alias to command name", func() {
			cfg := &config.Config{
				CommandAliases: map[string]string{
					"app": "approve",
					"a":   "approve",
				},
			}

			Expect(cfg.ResolveAlias("app")).To(Equal("approve"))
			Expect(cfg.ResolveAlias("a")).To(Equal("approve"))
		})

		It("should return the original command if no alias exists", func() {
			cfg := &config.Config{
				CommandAliases: map[string]string{
					"app": "approve",
				},
			}

			Expect(cfg.ResolveAlias("approve")).To(Equal("approve"))
			Expect(cfg.ResolveAlias("merge")).To(Equal("merge"))
		})
	})
})
