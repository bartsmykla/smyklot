package commands_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/smyklot/pkg/commands"
	"github.com/smykla-labs/smyklot/pkg/config"
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

		Context("when parsing multiple commands in the same comment (legacy)", func() {
			It("should set Type to first command for backward compatibility", func() {
				cmd, err := commands.ParseCommand("/approve /merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should find all commands regardless of format", func() {
				cmd, err := commands.ParseCommand("/merge @smyklot approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands).To(ContainElements(commands.CommandApprove, commands.CommandMerge))
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

			It("should parse bare commands with extra text before", func() {
				cmd, err := commands.ParseCommand("please approve", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse bare commands with extra text after", func() {
				cmd, err := commands.ParseCommand("approve this", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
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
				// Still finds "approve" as bare command
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
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

		Context("when parsing multiple commands", func() {
			It("should parse multiple bare commands separated by space", func() {
				cmd, err := commands.ParseCommand("approve merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse multiple bare commands separated by newline", func() {
				cmd, err := commands.ParseCommand("approve\nmerge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse multiple slash commands", func() {
				cmd, err := commands.ParseCommand("/approve /merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse multiple mention commands", func() {
				cmd, err := commands.ParseCommand("@smyklot approve @smyklot merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse mixed command formats", func() {
				cmd, err := commands.ParseCommand("/approve @smyklot merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle duplicate commands", func() {
				cmd, err := commands.ParseCommand("approve approve merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse lgtm and merge together", func() {
				cmd, err := commands.ParseCommand("lgtm merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should ignore unknown words in multi-command text", func() {
				cmd, err := commands.ParseCommand("please approve and merge this", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when avoiding accidental bare command matches", func() {
			It("should NOT match bare commands in regular sentences", func() {
				cmd, err := commands.ParseCommand("I think we should approve this change", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should NOT match in longer explanatory text", func() {
				cmd, err := commands.ParseCommand("Let's approve this PR after the tests pass", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should NOT match in questions", func() {
				cmd, err := commands.ParseCommand("Can someone approve this?", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should match when line is mostly commands", func() {
				cmd, err := commands.ParseCommand("approve and merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should match on command-only lines in multiline text", func() {
				text := `This looks good!
approve
merge
Thanks!`
				cmd, err := commands.ParseCommand(text, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing unapprove commands", func() {
			It("should parse /unapprove command", func() {
				cmd, err := commands.ParseCommand("/unapprove", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandUnapprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse @smyklot unapprove command", func() {
				cmd, err := commands.ParseCommand("@smyklot unapprove", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandUnapprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse bare unapprove command", func() {
				cmd, err := commands.ParseCommand("unapprove", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandUnapprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse disapprove as alias for unapprove", func() {
				cmd, err := commands.ParseCommand("/disapprove", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandUnapprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should NOT parse unapprove when disabled via config", func() {
				cfg := &config.Config{
					DisableUnapprove: true,
				}
				cmd, err := commands.ParseCommand("/unapprove", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})
		})

		Context("when detecting contradicting commands", func() {
			It("should return error for approve and unapprove", func() {
				cmd, err := commands.ParseCommand("/approve /unapprove", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("contradicting commands"))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return error for merge and unapprove", func() {
				cmd, err := commands.ParseCommand("/merge /unapprove", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("contradicting commands"))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return error for approve, merge, and unapprove", func() {
				cmd, err := commands.ParseCommand("approve merge unapprove", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("contradicting commands"))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return error for lgtm and disapprove", func() {
				cmd, err := commands.ParseCommand("lgtm disapprove", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("contradicting commands"))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should allow approve and merge together", func() {
				cmd, err := commands.ParseCommand("/approve /merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should allow only unapprove", func() {
				cmd, err := commands.ParseCommand("/unapprove", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandUnapprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when protecting against single-letter alias matches", func() {
			cfg := &config.Config{
				CommandPrefix: "/",
				CommandAliases: map[string]string{
					"a": "approve",
					"m": "merge",
				},
			}

			It("should NOT match single-letter bare commands in sentences", func() {
				cmd, err := commands.ParseCommand("This is a test", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should NOT match single-letter in multi-word context", func() {
				cmd, err := commands.ParseCommand("I have a question", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should match single-letter when alone on line", func() {
				cmd, err := commands.ParseCommand("a", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should match single-letter in command-only context", func() {
				text := `a
m`
				cmd, err := commands.ParseCommand(text, cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should still match multi-letter commands normally", func() {
				cmd, err := commands.ParseCommand("This is approve and merge", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(BeEmpty())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should work with slash commands regardless", func() {
				cmd, err := commands.ParseCommand("This is /a test", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing help command", func() {
			It("should parse /help command", func() {
				cmd, err := commands.ParseCommand("/help", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandHelp))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse @smyklot help command", func() {
				cmd, err := commands.ParseCommand("@smyklot help", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandHelp))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse bare help command", func() {
				cmd, err := commands.ParseCommand("help", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(1))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandHelp))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should not contradict with other commands", func() {
				cmd, err := commands.ParseCommand("approve help", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands).To(ContainElement(commands.CommandApprove))
				Expect(cmd.Commands).To(ContainElement(commands.CommandHelp))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing technical content that should not be treated as commands", func() {
			It("should not parse 'Fixed' in PR review summary", func() {
				comment := `Addressed Copilot review feedback. Summary of changes:

**Fixed (Real Issues):**
1. ✅ exec_loader.go:89 - Fixed verifyExecutable() to properly check errors
2. ✅ grpc_loader.go:142,166 - Context propagation fixed: use parent context

All changes tested:
- task build ✅
- task lint ✅
- task test ✅`

				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
				Expect(cmd.Commands).To(BeEmpty())
			})

			It("should not parse technical summary with 'Fixed' keyword", func() {
				comment := "Fixed verifyExecutable() to properly check errors instead of ignoring them"

				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should not parse PR review feedback list", func() {
				comment := `Addressed review feedback:
- Fixed context propagation in loader
- Added error handling for edge cases
- Updated documentation`

				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should not parse commit message style text", func() {
				comment := "feat(loader): add support for plugin validation"

				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should not parse code review summary", func() {
				comment := "Reviewed the changes. All tests pass. Ready to merge once CI completes."

				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should not parse GitHub Actions workflow summary", func() {
				comment := `Workflow completed successfully:
- Build: ✅
- Test: ✅
- Lint: ✅`

				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})
		})

		Context("when parsing 'after CI' modifier with merge commands", func() {
			// Basic slash command variations
			It("should parse '/merge after CI' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/merge after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse '/squash after CI' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/squash after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse '/rebase after CI' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/rebase after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandRebase))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			// Case insensitivity
			It("should be case-insensitive for 'after ci'", func() {
				cmd, err := commands.ParseCommand("/merge after ci", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should be case-insensitive for 'AFTER CI'", func() {
				cmd, err := commands.ParseCommand("/merge AFTER CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should be case-insensitive for 'After Ci'", func() {
				cmd, err := commands.ParseCommand("/merge After Ci", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "when green" variations
			It("should parse '/merge when green'", func() {
				cmd, err := commands.ParseCommand("/merge when green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse '/squash when green'", func() {
				cmd, err := commands.ParseCommand("/squash when green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "when checks pass" variations
			It("should parse '/merge when checks pass'", func() {
				cmd, err := commands.ParseCommand("/merge when checks pass", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse '/merge when check passes'", func() {
				cmd, err := commands.ParseCommand("/merge when check passes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "once CI passes" variations
			It("should parse '/merge once CI passes'", func() {
				cmd, err := commands.ParseCommand("/merge once CI passes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse '/merge once checks pass'", func() {
				cmd, err := commands.ParseCommand("/merge once checks pass", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "after checks" variations
			It("should parse '/merge after checks'", func() {
				cmd, err := commands.ParseCommand("/merge after checks", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse '/merge after check'", func() {
				cmd, err := commands.ParseCommand("/merge after check", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "when CI passes" variation
			It("should parse '/merge when CI passes'", func() {
				cmd, err := commands.ParseCommand("/merge when CI passes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "checks are green" variation
			It("should parse '/merge when checks are green'", func() {
				cmd, err := commands.ParseCommand("/merge when checks are green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// Mention command variations
			It("should parse '@smyklot merge after CI'", func() {
				cmd, err := commands.ParseCommand("@smyklot merge after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse '@smyklot squash when green'", func() {
				cmd, err := commands.ParseCommand("@smyklot squash when green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// Bare command variations with CI modifiers
			// CI-related terms are now recognized as command context, allowing bare commands
			It("should parse 'merge after CI' as bare command", func() {
				cmd, err := commands.ParseCommand("merge after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse 'squash after CI' as bare command", func() {
				cmd, err := commands.ParseCommand("squash after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse 'rebase after CI' as bare command", func() {
				cmd, err := commands.ParseCommand("rebase after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandRebase))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should parse 'squash when green' as bare command", func() {
				cmd, err := commands.ParseCommand("squash when green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			// Combined commands
			It("should parse '/approve /merge after CI' with approve first", func() {
				cmd, err := commands.ParseCommand("/approve /merge after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'lgtm merge after CI' as bare command with multiple commands", func() {
				cmd, err := commands.ParseCommand("lgtm merge after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Commands).To(HaveLen(2))
				Expect(cmd.Commands[0]).To(Equal(commands.CommandApprove))
				Expect(cmd.Commands[1]).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.IsValid).To(BeTrue())
			})

			// Multiline
			It("should parse modifier in multiline comment", func() {
				cmd, err := commands.ParseCommand("/merge\nafter CI please", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse modifier when on separate line", func() {
				comment := `/merge
when checks pass`
				cmd, err := commands.ParseCommand(comment, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})
		})

		Context("when parsing extended CI terminology", func() {
			// CD terminology
			It("should parse '/merge after CD' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/merge after CD", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'squash after CD' as bare command", func() {
				cmd, err := commands.ParseCommand("squash after CD", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// GHA terminology
			It("should parse '/merge after GHA' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/merge after GHA", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'rebase after GHA' as bare command", func() {
				cmd, err := commands.ParseCommand("rebase after GHA", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandRebase))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// Workflows terminology
			It("should parse '/merge after workflows' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/merge after workflows", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'squash after workflow' as bare command", func() {
				cmd, err := commands.ParseCommand("squash after workflow", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// GitHub Actions terminology
			It("should parse '/merge after github actions' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/merge after github actions", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'merge after github actions' as bare command", func() {
				cmd, err := commands.ParseCommand("merge after github actions", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "finishes" and "completes" variations
			It("should parse '/merge when CI finishes' with WaitForCI=true", func() {
				cmd, err := commands.ParseCommand("/merge when CI finishes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'squash when CI completes' as bare command", func() {
				cmd, err := commands.ParseCommand("squash when CI completes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should parse 'rebase once workflows finish' as bare command", func() {
				cmd, err := commands.ParseCommand("rebase once workflows finish", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandRebase))
				Expect(cmd.WaitForCI).To(BeTrue())
			})
		})

		Context("when parsing 'required checks only' modifier", func() {
			// Slash commands with "required" modifier
			It("should parse '/merge after required CI' with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("/merge after required CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			It("should parse '/squash after required CI' with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("/squash after required CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			It("should parse '/rebase after required CI' with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("/rebase after required CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandRebase))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			// Bare commands with "required" modifier
			It("should parse 'squash after required CI' as bare command with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("squash after required CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			It("should parse 'merge after required checks' as bare command with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("merge after required checks", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			It("should parse 'rebase when required CI passes' as bare command with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("rebase when required CI passes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandRebase))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			// Mention commands with "required" modifier
			It("should parse '@smyklot merge after required CI' with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("@smyklot merge after required CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			// "required" with extended terminology
			It("should parse 'squash after required GHA' as bare command with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("squash after required GHA", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandSquash))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			It("should parse 'merge when required workflows finish' as bare command with RequiredChecksOnly=true", func() {
				cmd, err := commands.ParseCommand("merge when required workflows finish", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			// Case insensitivity for "required"
			It("should be case-insensitive for 'REQUIRED'", func() {
				cmd, err := commands.ParseCommand("/merge after REQUIRED CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			It("should be case-insensitive for 'Required'", func() {
				cmd, err := commands.ParseCommand("squash after Required checks", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeTrue())
			})

			// Regular (non-required) CI should NOT set RequiredChecksOnly
			It("should NOT set RequiredChecksOnly for regular '/merge after CI'", func() {
				cmd, err := commands.ParseCommand("/merge after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeFalse())
			})

			It("should NOT set RequiredChecksOnly for regular 'squash when green'", func() {
				cmd, err := commands.ParseCommand("squash when green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
				Expect(cmd.RequiredChecksOnly).To(BeFalse())
			})
		})

		Context("when 'after CI' modifier should NOT apply", func() {
			// Regular merge without modifier
			It("should NOT set WaitForCI for plain '/merge'", func() {
				cmd, err := commands.ParseCommand("/merge", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			// Approve-only commands should not get WaitForCI
			It("should NOT set WaitForCI for '/approve' even with modifier text", func() {
				cmd, err := commands.ParseCommand("/approve after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			It("should NOT set WaitForCI for 'lgtm' alone", func() {
				cmd, err := commands.ParseCommand("lgtm", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			// Non-command text with CI-related words
			It("should NOT parse as command: 'waiting for CI to pass'", func() {
				cmd, err := commands.ParseCommand("waiting for CI to pass", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should NOT parse natural text mentioning CI", func() {
				cmd, err := commands.ParseCommand("Let's merge this after CI is done", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should NOT match partial phrases like 'after breakfast'", func() {
				cmd, err := commands.ParseCommand("/merge after breakfast", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			It("should NOT match 'after lunch'", func() {
				cmd, err := commands.ParseCommand("/merge after lunch", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			It("should NOT match 'when ready'", func() {
				cmd, err := commands.ParseCommand("/merge when ready", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			// Cleanup command should not get WaitForCI
			It("should NOT set WaitForCI for '/cleanup'", func() {
				cmd, err := commands.ParseCommand("/cleanup after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandCleanup))
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			// Help command should not get WaitForCI
			It("should NOT set WaitForCI for '/help'", func() {
				cmd, err := commands.ParseCommand("/help after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandHelp))
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			// Unapprove should not get WaitForCI
			It("should NOT set WaitForCI for '/unapprove'", func() {
				cmd, err := commands.ParseCommand("/unapprove after CI", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnapprove))
				Expect(cmd.WaitForCI).To(BeFalse())
			})
		})

		Context("edge cases for 'after CI' regex matching", func() {
			// Word boundary tests
			It("should NOT match 'CIrcle' as CI", func() {
				cmd, err := commands.ParseCommand("/merge after CIrcle", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			It("should NOT match 'checking' as check", func() {
				cmd, err := commands.ParseCommand("/merge after checking", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			It("should NOT match 'greenhouse' as green", func() {
				cmd, err := commands.ParseCommand("/merge when greenhouse", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeFalse())
			})

			// Whitespace variations
			It("should handle multiple spaces between words", func() {
				cmd, err := commands.ParseCommand("/merge after   CI", nil)
				Expect(err).NotTo(HaveOccurred())
				// \s+ matches one or more whitespace characters, so multiple spaces work
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should handle tabs between words", func() {
				cmd, err := commands.ParseCommand("/merge after\tCI", nil)
				Expect(err).NotTo(HaveOccurred())
				// Tab is whitespace, should match
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// Extra text around modifier
			It("should match with text before modifier", func() {
				cmd, err := commands.ParseCommand("/merge please after CI thanks", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match with text after modifier", func() {
				cmd, err := commands.ParseCommand("/merge after CI please", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// Punctuation
			It("should match modifier followed by period", func() {
				cmd, err := commands.ParseCommand("/merge after CI.", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match modifier followed by exclamation", func() {
				cmd, err := commands.ParseCommand("/merge after CI!", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// All variations of pass/passes
			It("should match 'CI pass'", func() {
				cmd, err := commands.ParseCommand("/merge when CI pass", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match 'CI passes'", func() {
				cmd, err := commands.ParseCommand("/merge when CI passes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match 'check pass'", func() {
				cmd, err := commands.ParseCommand("/merge when check pass", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match 'checks passes'", func() {
				cmd, err := commands.ParseCommand("/merge when checks passes", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// "a green" vs "are green"
			It("should match 'checks a green'", func() {
				cmd, err := commands.ParseCommand("/merge when checks a green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match 'checks are green'", func() {
				cmd, err := commands.ParseCommand("/merge when checks are green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			// After with just the keyword
			It("should match 'after green'", func() {
				cmd, err := commands.ParseCommand("/merge after green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})

			It("should match 'once green'", func() {
				cmd, err := commands.ParseCommand("/merge once green", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.WaitForCI).To(BeTrue())
			})
		})
	})
})
