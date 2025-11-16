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
	})
})
