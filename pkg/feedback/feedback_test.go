package feedback_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bartsmykla/smyklot/pkg/feedback"
)

var _ = Describe("Feedback System [Unit]", func() {
	Describe("FeedbackType", func() {
		It("should have Success type", func() {
			Expect(feedback.Success).To(Equal(feedback.FeedbackType("success")))
		})

		It("should have Error type", func() {
			Expect(feedback.Error).To(Equal(feedback.FeedbackType("error")))
		})

		It("should have Warning type", func() {
			Expect(feedback.Warning).To(Equal(feedback.FeedbackType("warning")))
		})
	})

	Describe("NewSuccess", func() {
		It("should create success feedback with only emoji", func() {
			fb := feedback.NewSuccess()
			Expect(fb.Type).To(Equal(feedback.Success))
			Expect(fb.Emoji).To(Equal("✅"))
			Expect(fb.Message).To(BeEmpty())
		})
	})

	Describe("NewUnauthorized", func() {
		It("should create error feedback for unauthorized user", func() {
			fb := feedback.NewUnauthorized("john", []string{"alice", "bob"})
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(ContainSubstring("not authorized"))
			Expect(fb.Message).To(ContainSubstring("john"))
			Expect(fb.Message).To(ContainSubstring("alice"))
			Expect(fb.Message).To(ContainSubstring("bob"))
		})

		It("should handle single approver", func() {
			fb := feedback.NewUnauthorized("john", []string{"alice"})
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Message).To(ContainSubstring("alice"))
		})

		It("should handle empty approvers list", func() {
			fb := feedback.NewUnauthorized("john", []string{})
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Message).To(ContainSubstring("no approvers"))
		})

		It("should handle multiple approvers", func() {
			fb := feedback.NewUnauthorized("john", []string{"alice", "bob", "charlie"})
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Message).To(ContainSubstring("alice"))
			Expect(fb.Message).To(ContainSubstring("bob"))
			Expect(fb.Message).To(ContainSubstring("charlie"))
		})
	})

	Describe("NewInvalidCommand", func() {
		It("should create error feedback for invalid command", func() {
			fb := feedback.NewInvalidCommand("/invalid")
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(MatchRegexp(`(?i)invalid`))
			Expect(fb.Message).To(ContainSubstring("/invalid"))
		})

		It("should suggest valid commands", func() {
			fb := feedback.NewInvalidCommand("/test")
			Expect(fb.Message).To(ContainSubstring("/approve"))
			Expect(fb.Message).To(ContainSubstring("/merge"))
		})
	})

	Describe("NewAlreadyApproved", func() {
		It("should create warning feedback", func() {
			fb := feedback.NewAlreadyApproved("alice")
			Expect(fb.Type).To(Equal(feedback.Warning))
			Expect(fb.Emoji).To(Equal("⚠️"))
			Expect(fb.Message).To(MatchRegexp(`(?i)approved`))
			Expect(fb.Message).To(ContainSubstring("alice"))
		})
	})

	Describe("NewAlreadyMerged", func() {
		It("should create warning feedback", func() {
			fb := feedback.NewAlreadyMerged()
			Expect(fb.Type).To(Equal(feedback.Warning))
			Expect(fb.Emoji).To(Equal("⚠️"))
			Expect(fb.Message).To(MatchRegexp(`(?i)merged`))
		})
	})

	Describe("NewPRNotReady", func() {
		It("should create error feedback with reason", func() {
			fb := feedback.NewPRNotReady("CI checks failing")
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(ContainSubstring("not ready"))
			Expect(fb.Message).To(ContainSubstring("CI checks failing"))
		})

		It("should handle different reasons", func() {
			reasons := []string{
				"required reviews not met",
				"conflicts with base branch",
				"checks have not completed",
			}

			for _, reason := range reasons {
				fb := feedback.NewPRNotReady(reason)
				Expect(fb.Message).To(ContainSubstring(reason))
			}
		})
	})

	Describe("NewMergeConflict", func() {
		It("should create error feedback for merge conflict", func() {
			fb := feedback.NewMergeConflict()
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(MatchRegexp(`(?i)conflict`))
		})
	})

	Describe("NewNoOWNERSFile", func() {
		It("should create error feedback for missing OWNERS file", func() {
			fb := feedback.NewNoOWNERSFile()
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(ContainSubstring("OWNERS"))
			Expect(fb.Message).To(ContainSubstring("not found"))
		})
	})

	Describe("Feedback.String", func() {
		It("should return emoji only for success", func() {
			fb := feedback.NewSuccess()
			Expect(fb.String()).To(Equal("✅"))
		})

		It("should return emoji + message for error", func() {
			fb := feedback.NewInvalidCommand("/test")
			str := fb.String()
			Expect(str).To(ContainSubstring("❌"))
			Expect(str).To(MatchRegexp(`(?i)invalid`))
		})

		It("should return emoji + message for warning", func() {
			fb := feedback.NewAlreadyApproved("alice")
			str := fb.String()
			Expect(str).To(ContainSubstring("⚠️"))
			Expect(str).To(MatchRegexp(`(?i)approved`))
		})
	})

	Describe("Feedback.RequiresComment", func() {
		It("should return false for success", func() {
			fb := feedback.NewSuccess()
			Expect(fb.RequiresComment()).To(BeFalse())
		})

		It("should return true for error", func() {
			fb := feedback.NewUnauthorized("john", []string{"alice"})
			Expect(fb.RequiresComment()).To(BeTrue())
		})

		It("should return true for warning", func() {
			fb := feedback.NewAlreadyApproved("alice")
			Expect(fb.RequiresComment()).To(BeTrue())
		})
	})

	Describe("Edge Cases", func() {
		Context("when creating feedback with empty values", func() {
			It("should handle empty username in unauthorized", func() {
				fb := feedback.NewUnauthorized("", []string{"alice"})
				Expect(fb.Type).To(Equal(feedback.Error))
				Expect(fb.Message).NotTo(BeEmpty())
			})

			It("should handle empty command in invalid command", func() {
				fb := feedback.NewInvalidCommand("")
				Expect(fb.Type).To(Equal(feedback.Error))
				Expect(fb.Message).NotTo(BeEmpty())
			})

			It("should handle empty reason in PR not ready", func() {
				fb := feedback.NewPRNotReady("")
				Expect(fb.Type).To(Equal(feedback.Error))
				Expect(fb.Message).NotTo(BeEmpty())
			})
		})

		Context("when creating feedback with special characters", func() {
			It("should handle usernames with special characters", func() {
				fb := feedback.NewUnauthorized("user-123", []string{"admin_1"})
				Expect(fb.Message).To(ContainSubstring("user-123"))
				Expect(fb.Message).To(ContainSubstring("admin_1"))
			})

			It("should handle commands with special characters", func() {
				fb := feedback.NewInvalidCommand("/test-command")
				Expect(fb.Message).To(ContainSubstring("/test-command"))
			})
		})
	})

	Describe("Message Formatting", func() {
		It("should format unauthorized message clearly", func() {
			fb := feedback.NewUnauthorized("john", []string{"alice", "bob"})
			Expect(fb.Message).To(MatchRegexp(`(?i)not authorized`))
			Expect(fb.Message).To(MatchRegexp(`john`))
		})

		It("should format invalid command message with suggestions", func() {
			fb := feedback.NewInvalidCommand("/test")
			Expect(fb.Message).To(MatchRegexp(`(?i)invalid`))
			Expect(fb.Message).To(MatchRegexp(`/approve|/merge`))
		})

		It("should format PR not ready message with reason", func() {
			fb := feedback.NewPRNotReady("checks failing")
			Expect(fb.Message).To(MatchRegexp(`(?i)not ready`))
			Expect(fb.Message).To(MatchRegexp(`checks failing`))
		})
	})
})
