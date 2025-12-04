package feedback_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/smyklot/pkg/feedback"
)

var _ = Describe("Feedback System [Unit]", func() {
	Describe("Type", func() {
		It("should have Success type", func() {
			Expect(feedback.Success).To(Equal(feedback.Type("success")))
		})

		It("should have Error type", func() {
			Expect(feedback.Error).To(Equal(feedback.Type("error")))
		})

		It("should have Warning type", func() {
			Expect(feedback.Warning).To(Equal(feedback.Type("warning")))
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
			Expect(fb.Message).To(MatchRegexp(`(?i)no approvers`))
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

	Describe("NewNoCodeownersFile", func() {
		It("should create error feedback for missing CODEOWNERS file", func() {
			fb := feedback.NewNoCodeownersFile()
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(ContainSubstring("CODEOWNERS"))
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

	Describe("CombineFeedback", func() {
		It("should return single feedback when only one provided", func() {
			fb := feedback.NewSuccess()
			combined := feedback.CombineFeedback([]*feedback.Feedback{fb}, false)
			Expect(combined).To(Equal(fb))
		})

		It("should return success feedback when empty slice", func() {
			combined := feedback.CombineFeedback([]*feedback.Feedback{}, false)
			Expect(combined.Type).To(Equal(feedback.Success))
			Expect(combined.Emoji).To(Equal("✅"))
		})

		It("should combine all success feedback with quietSuccess=false", func() {
			feedbacks := []*feedback.Feedback{
				feedback.NewApprovalSuccess("alice", false),
				feedback.NewMergeSuccess("alice", false),
			}
			combined := feedback.CombineFeedback(feedbacks, false)
			Expect(combined.Type).To(Equal(feedback.Success))
			Expect(combined.Emoji).To(Equal("✅"))
			Expect(combined.Message).To(ContainSubstring("Approved"))
			Expect(combined.Message).To(ContainSubstring("Merged"))
		})

		It("should combine all success feedback with quietSuccess=true", func() {
			feedbacks := []*feedback.Feedback{
				feedback.NewApprovalSuccess("alice", true),
				feedback.NewMergeSuccess("alice", true),
			}
			combined := feedback.CombineFeedback(feedbacks, true)
			Expect(combined.Type).To(Equal(feedback.Success))
			Expect(combined.Emoji).To(Equal("✅"))
			Expect(combined.Message).To(BeEmpty())
		})

		It("should combine all errors with error reaction", func() {
			feedbacks := []*feedback.Feedback{
				feedback.NewApprovalFailed("permission denied"),
				feedback.NewMergeFailed("not mergeable"),
			}
			combined := feedback.CombineFeedback(feedbacks, false)
			Expect(combined.Type).To(Equal(feedback.Error))
			Expect(combined.Emoji).To(Equal("❌"))
			Expect(combined.Message).To(ContainSubstring("permission denied"))
			Expect(combined.Message).To(ContainSubstring("not mergeable"))
		})

		It("should use warning reaction for mixed success and error", func() {
			feedbacks := []*feedback.Feedback{
				feedback.NewApprovalSuccess("alice", false),
				feedback.NewMergeFailed("merge conflicts"),
			}
			combined := feedback.CombineFeedback(feedbacks, false)
			Expect(combined.Type).To(Equal(feedback.Warning))
			Expect(combined.Emoji).To(Equal("😕"))
			Expect(combined.Message).To(ContainSubstring("Partial Success"))
			Expect(combined.Message).To(ContainSubstring("Approved"))
			Expect(combined.Message).To(ContainSubstring("merge conflicts"))
		})

		It("should respect quietSuccess for mixed results", func() {
			feedbacks := []*feedback.Feedback{
				feedback.NewApprovalSuccess("alice", true),
				feedback.NewMergeFailed("not mergeable"),
			}
			combined := feedback.CombineFeedback(feedbacks, true)
			Expect(combined.Type).To(Equal(feedback.Warning))
			Expect(combined.Emoji).To(Equal("😕"))
			Expect(combined.Message).To(ContainSubstring("Partial Success"))
			Expect(combined.Message).To(ContainSubstring("not mergeable"))
			Expect(combined.Message).NotTo(ContainSubstring("Approved"))
		})

		It("should include warnings in mixed results", func() {
			feedbacks := []*feedback.Feedback{
				feedback.NewApprovalSuccess("alice", false),
				feedback.NewAlreadyMerged(),
			}
			combined := feedback.CombineFeedback(feedbacks, false)
			Expect(combined.Type).To(Equal(feedback.Warning))
			Expect(combined.Emoji).To(Equal("😕"))
			Expect(combined.Message).To(ContainSubstring("Partial Success"))
		})
	})

	Describe("Pending Type", func() {
		It("should have Pending type", func() {
			Expect(feedback.Pending).To(Equal(feedback.Type("pending")))
		})
	})

	Describe("NewPendingCI", func() {
		It("should create pending feedback with hourglass emoji", func() {
			fb := feedback.NewPendingCI("alice", "merge")
			Expect(fb.Type).To(Equal(feedback.Pending))
			Expect(fb.Emoji).To(Equal("⏳"))
			Expect(fb.Message).To(ContainSubstring("Waiting for CI"))
			Expect(fb.Message).To(ContainSubstring("alice"))
			Expect(fb.Message).To(ContainSubstring("merge"))
		})

		It("should include different merge methods", func() {
			methods := []string{"merge", "squash", "rebase"}

			for _, method := range methods {
				fb := feedback.NewPendingCI("bob", method)
				Expect(fb.Message).To(ContainSubstring(method))
			}
		})

		It("should indicate automatic merge on CI success", func() {
			fb := feedback.NewPendingCI("alice", "merge")
			Expect(fb.Message).To(ContainSubstring("automatically"))
		})
	})

	Describe("NewPendingCIMerged", func() {
		It("should create success feedback with message when quietSuccess=false", func() {
			fb := feedback.NewPendingCIMerged("alice", false)
			Expect(fb.Type).To(Equal(feedback.Success))
			Expect(fb.Emoji).To(Equal("✅"))
			Expect(fb.Message).To(ContainSubstring("CI Passed"))
			Expect(fb.Message).To(ContainSubstring("PR Merged"))
			Expect(fb.Message).To(ContainSubstring("alice"))
		})

		It("should create success feedback with emoji only when quietSuccess=true", func() {
			fb := feedback.NewPendingCIMerged("alice", true)
			Expect(fb.Type).To(Equal(feedback.Success))
			Expect(fb.Emoji).To(Equal("✅"))
			Expect(fb.Message).To(BeEmpty())
		})

		It("should indicate checks passed", func() {
			fb := feedback.NewPendingCIMerged("bob", false)
			Expect(fb.Message).To(ContainSubstring("checks passed"))
		})
	})

	Describe("NewPendingCIFailed", func() {
		It("should create error feedback with reason", func() {
			fb := feedback.NewPendingCIFailed("3 checks failed")
			Expect(fb.Type).To(Equal(feedback.Error))
			Expect(fb.Emoji).To(Equal("❌"))
			Expect(fb.Message).To(ContainSubstring("CI Failed"))
			Expect(fb.Message).To(ContainSubstring("Merge Cancelled"))
			Expect(fb.Message).To(ContainSubstring("3 checks failed"))
		})

		It("should suggest fixing checks", func() {
			fb := feedback.NewPendingCIFailed("tests failing")
			Expect(fb.Message).To(ContainSubstring("fix"))
			Expect(fb.Message).To(ContainSubstring("try again"))
		})

		It("should handle different failure reasons", func() {
			reasons := []string{
				"2 of 5 checks failed",
				"CI timed out",
				"build error in workflow",
			}

			for _, reason := range reasons {
				fb := feedback.NewPendingCIFailed(reason)
				Expect(fb.Message).To(ContainSubstring(reason))
			}
		})
	})
})
