package commands_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bartsmykla/smyklot/pkg/commands"
)

var _ = Describe("Command Parser [Unit]", func() {
	Describe("ParseCommand", func() {
		Context("when parsing slash commands", func() {
			It("should parse /approve command", func() {
				cmd, err := commands.ParseCommand("/approve")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("/approve"))
			})

			It("should parse /merge command", func() {
				cmd, err := commands.ParseCommand("/merge")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("/merge"))
			})

			It("should handle slash commands with leading whitespace", func() {
				cmd, err := commands.ParseCommand("  /approve")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle slash commands with trailing whitespace", func() {
				cmd, err := commands.ParseCommand("/merge  ")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should be case-insensitive for slash commands", func() {
				cmd, err := commands.ParseCommand("/APPROVE")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing mention commands", func() {
			It("should parse @smyklot approve command", func() {
				cmd, err := commands.ParseCommand("@smyklot approve")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("@smyklot approve"))
			})

			It("should parse @smyklot merge command", func() {
				cmd, err := commands.ParseCommand("@smyklot merge")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
				Expect(cmd.Raw).To(Equal("@smyklot merge"))
			})

			It("should handle mention commands with extra text before", func() {
				cmd, err := commands.ParseCommand("Hey @smyklot approve this please")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle mention commands with extra text after", func() {
				cmd, err := commands.ParseCommand("@smyklot merge when ready")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should be case-insensitive for mention commands", func() {
				cmd, err := commands.ParseCommand("@smyklot APPROVE")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should handle @Smyklot with capital S", func() {
				cmd, err := commands.ParseCommand("@Smyklot approve")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing invalid commands", func() {
			It("should return unknown for non-command text", func() {
				cmd, err := commands.ParseCommand("just a regular comment")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return unknown for invalid slash command", func() {
				cmd, err := commands.ParseCommand("/invalid")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should return unknown for mention without command", func() {
				cmd, err := commands.ParseCommand("@smyklot hello")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should handle empty string", func() {
				cmd, err := commands.ParseCommand("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})

			It("should handle whitespace only", func() {
				cmd, err := commands.ParseCommand("   ")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandUnknown))
				Expect(cmd.IsValid).To(BeFalse())
			})
		})

		Context("when parsing commands in multiline text", func() {
			It("should find slash command in first line", func() {
				text := `/approve
Some additional context
about why this should be approved`
				cmd, err := commands.ParseCommand(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should find mention command in multiline text", func() {
				text := `This looks good!
@smyklot approve
Thanks for the PR!`
				cmd, err := commands.ParseCommand(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})

		Context("when parsing multiple commands in same comment", func() {
			It("should prioritize the first command found", func() {
				cmd, err := commands.ParseCommand("/approve /merge")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandApprove))
				Expect(cmd.IsValid).To(BeTrue())
			})

			It("should prioritize slash command over mention command", func() {
				cmd, err := commands.ParseCommand("/merge @smyklot approve")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd.Type).To(Equal(commands.CommandMerge))
				Expect(cmd.IsValid).To(BeTrue())
			})
		})
	})
})
