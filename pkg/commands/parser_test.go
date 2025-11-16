package commands_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bartsmykla/smyklot/pkg/commands"
	"github.com/bartsmykla/smyklot/pkg/config"
)

var _ = Describe("Command Parser [Unit]", func() {
	Describe("ParseCommand", func() {
		Context("when parsing slash commands", func() {
			It("should parse /approve command", func() {
				cmd, err := commands.ParseCommand("/approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("/approve"))
			})

			It("should parse /merge command", func() {
				cmd, err := commands.ParseCommand("/merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("/merge"))
			})

			It("should handle slash commands with leading whitespace", func() {
				cmd, err := commands.ParseCommand("  /approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle slash commands with trailing whitespace", func() {
				cmd, err := commands.ParseCommand("/merge  ", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should be case-insensitive for slash commands", func() {
				cmd, err := commands.ParseCommand("/APPROVE", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing mention commands", func() {
			It("should parse @smyklot approve command", func() {
				cmd, err := commands.ParseCommand("@smyklot approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("@smyklot approve"))
			})

			It("should parse @smyklot merge command", func() {
				cmd, err := commands.ParseCommand("@smyklot merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("@smyklot merge"))
			})

			It("should handle mention commands with extra text before", func() {
				cmd, err := commands.ParseCommand("Hey @smyklot approve this please", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle mention commands with extra text after", func() {
				cmd, err := commands.ParseCommand("@smyklot merge when ready", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should be case-insensitive for mention commands", func() {
				cmd, err := commands.ParseCommand("@smyklot APPROVE", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle @Smyklot with capital S", func() {
				cmd, err := commands.ParseCommand("@Smyklot approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing invalid commands", func() {
			It("should return unknown for non-command text", func() {
				cmd, err := commands.ParseCommand("just a regular comment", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return unknown for invalid slash command", func() {
				cmd, err := commands.ParseCommand("/invalid", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return unknown for mention without a command", func() {
				cmd, err := commands.ParseCommand("@smyklot hello", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should handle empty string", func() {
				cmd, err := commands.ParseCommand("", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should handle whitespace only", func() {
				cmd, err := commands.ParseCommand("   ", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})
		})

		Context("when parsing commands in multiline text", func() {
			It("should find the slash command in the first line", func() {
				text := `/approve
Some additional context
about why this should be approved`
				cmd, err := commands.ParseCommand(text, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should find mention command in multiline text", func() {
				text := `This looks good!
@smyklot approve
Thanks for the PR!`
				cmd, err := commands.ParseCommand(text, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing multiple commands in the same comment", func() {
			It("should prioritize the first command found", func() {
				cmd, err := commands.ParseCommand("/approve /merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should prioritize slash command over mention command", func() {
				cmd, err := commands.ParseCommand("/merge @smyklot approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing bare commands", func() {
			It("should parse 'approve' as bare command", func() {
				cmd, err := commands.ParseCommand("approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("approve"))
			})

			It("should parse 'accept' as bare command", func() {
				cmd, err := commands.ParseCommand("accept", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse 'lgtm' as bare command", func() {
				cmd, err := commands.ParseCommand("lgtm", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse 'merge' as bare command", func() {
				cmd, err := commands.ParseCommand("merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle bare commands with leading whitespace", func() {
				cmd, err := commands.ParseCommand("  approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle bare commands with trailing whitespace", func() {
				cmd, err := commands.ParseCommand("merge  ", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should be case-insensitive for bare commands", func() {
				cmd, err := commands.ParseCommand("APPROVE", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())

				cmd, err = commands.ParseCommand("LGTM", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should NOT parse bare commands with extra text before", func() {
				cmd, err := commands.ParseCommand("please approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should NOT parse bare commands with extra text after", func() {
				cmd, err := commands.ParseCommand("approve this", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should prioritize slash commands over bare commands", func() {
				cfg := config.Default()
				cfg.CommandPrefix = "/"

				cmd, err := commands.ParseCommand("/merge", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should prioritize mention commands over bare commands", func() {
				cmd, err := commands.ParseCommand("@smyklot approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should work with bare commands when slash and mention commands fail", func() {
				cmd, err := commands.ParseCommand("lgtm", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("with custom configuration", func() {
			It("should use a custom command prefix", func() {
				cfg := config.Default()
				cfg.CommandPrefix = "!"

				cmd, err := commands.ParseCommand("!approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should not parse the default prefix when a custom prefix is set", func() {
				cfg := config.Default()
				cfg.CommandPrefix = "!"

				cmd, err := commands.ParseCommand("/approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should resolve command aliases", func() {
				cfg := config.Default()
				cfg.CommandAliases = map[string]string{
					"app": "approve",
					"a":   "approve",
					"m":   "merge",
				}

				cmd, err := commands.ParseCommand("/app", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())

				cmd, err = commands.ParseCommand("@smyklot m", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should disable mention commands when configured", func() {
				cfg := config.Default()
				cfg.DisableMentions = true

				cmd, err := commands.ParseCommand("@smyklot approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should still parse slash commands when mentions are disabled", func() {
				cfg := config.Default()
				cfg.DisableMentions = true

				cmd, err := commands.ParseCommand("/approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should only allow specified commands", func() {
				cfg := config.Default()
				cfg.AllowedCommands = []string{"approve"}

				cmd, err := commands.ParseCommand("/approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())

				cmd, err = commands.ParseCommand("/merge", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should combine multiple config options", func() {
				cfg := config.Default()
				cfg.CommandPrefix = "!"
				cfg.CommandAliases = map[string]string{
					"app": "approve",
				}
				cfg.AllowedCommands = []string{"approve"}

				cmd, err := commands.ParseCommand("!app", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should resolve aliases for bare commands", func() {
				cfg := config.Default()
				cfg.CommandAliases = map[string]string{
					"ok": "approve",
				}

				cmd, err := commands.ParseCommand("ok", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should respect allowed commands for bare commands", func() {
				cfg := config.Default()
				cfg.AllowedCommands = []string{"approve"}

				cmd, err := commands.ParseCommand("approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())

				cmd, err = commands.ParseCommand("merge", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should disable bare commands when configured", func() {
				cfg := config.Default()
				cfg.DisableBareCommands = true

				cmd, err := commands.ParseCommand("approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should still parse slash commands when bare commands are disabled", func() {
				cfg := config.Default()
				cfg.DisableBareCommands = true

				cmd, err := commands.ParseCommand("/approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should still parse mention commands when bare commands are disabled", func() {
				cfg := config.Default()
				cfg.DisableBareCommands = true

				cmd, err := commands.ParseCommand("@smyklot approve", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})
	})
})
