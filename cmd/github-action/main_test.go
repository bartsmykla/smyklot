package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/smyklot/pkg/github"
)

var _ = Describe("Main Pending CI Functions [Unit]", func() {
	Describe("getPendingCILabel", func() {
		Context("with requiredOnly=false", func() {
			It("should return merge label for merge method", func() {
				label := getPendingCILabel(github.MergeMethodMerge, false)
				Expect(label).To(Equal(github.LabelPendingCIMerge))
			})

			It("should return squash label for squash method", func() {
				label := getPendingCILabel(github.MergeMethodSquash, false)
				Expect(label).To(Equal(github.LabelPendingCISquash))
			})

			It("should return rebase label for rebase method", func() {
				label := getPendingCILabel(github.MergeMethodRebase, false)
				Expect(label).To(Equal(github.LabelPendingCIRebase))
			})

			It("should return merge label for unknown method", func() {
				label := getPendingCILabel(github.MergeMethod("unknown"), false)
				Expect(label).To(Equal(github.LabelPendingCIMerge))
			})
		})

		Context("with requiredOnly=true", func() {
			It("should return required merge label for merge method", func() {
				label := getPendingCILabel(github.MergeMethodMerge, true)
				Expect(label).To(Equal(github.LabelPendingCIMergeRequired))
			})

			It("should return required squash label for squash method", func() {
				label := getPendingCILabel(github.MergeMethodSquash, true)
				Expect(label).To(Equal(github.LabelPendingCISquashRequired))
			})

			It("should return required rebase label for rebase method", func() {
				label := getPendingCILabel(github.MergeMethodRebase, true)
				Expect(label).To(Equal(github.LabelPendingCIRebaseRequired))
			})

			It("should return required merge label for unknown method", func() {
				label := getPendingCILabel(github.MergeMethod("unknown"), true)
				Expect(label).To(Equal(github.LabelPendingCIMergeRequired))
			})
		})
	})

	Describe("getMergeMethodName", func() {
		It("should return 'merge' for merge method", func() {
			name := getMergeMethodName(github.MergeMethodMerge)
			Expect(name).To(Equal("merge"))
		})

		It("should return 'squash' for squash method", func() {
			name := getMergeMethodName(github.MergeMethodSquash)
			Expect(name).To(Equal("squash"))
		})

		It("should return 'rebase' for rebase method", func() {
			name := getMergeMethodName(github.MergeMethodRebase)
			Expect(name).To(Equal("rebase"))
		})

		It("should return 'merge' for unknown method", func() {
			name := getMergeMethodName(github.MergeMethod("unknown"))
			Expect(name).To(Equal("merge"))
		})
	})

	Describe("isBotAlreadyApproved", func() {
		It("should return true when bot has approved", func() {
			info := &github.PRInfo{
				ApprovedBy: []string{"user1", "smyklot[bot]", "user2"},
			}
			Expect(isBotAlreadyApproved(info, "smyklot[bot]")).To(BeTrue())
		})

		It("should return false when bot has not approved", func() {
			info := &github.PRInfo{
				ApprovedBy: []string{"user1", "user2"},
			}
			Expect(isBotAlreadyApproved(info, "smyklot[bot]")).To(BeFalse())
		})

		It("should return false for empty approvers list", func() {
			info := &github.PRInfo{
				ApprovedBy: []string{},
			}
			Expect(isBotAlreadyApproved(info, "smyklot[bot]")).To(BeFalse())
		})

		It("should handle different bot usernames", func() {
			info := &github.PRInfo{
				ApprovedBy: []string{"custom-bot"},
			}
			Expect(isBotAlreadyApproved(info, "custom-bot")).To(BeTrue())
			Expect(isBotAlreadyApproved(info, "smyklot[bot]")).To(BeFalse())
		})
	})

	Describe("shouldEnableAutoMerge", func() {
		It("should return true for merge queue error", func() {
			err := NewGitHubError(ErrMergePR, githubError("merge queue is required"))
			Expect(shouldEnableAutoMerge(err)).To(BeTrue())
		})

		It("should return true for status checks error", func() {
			err := NewGitHubError(ErrMergePR, githubError("required status check"))
			Expect(shouldEnableAutoMerge(err)).To(BeTrue())
		})

		It("should return true for branch protection error", func() {
			err := NewGitHubError(ErrMergePR, githubError("branch protection rules"))
			Expect(shouldEnableAutoMerge(err)).To(BeTrue())
		})

		It("should return false for other errors", func() {
			err := NewGitHubError(ErrMergePR, githubError("some other error"))
			Expect(shouldEnableAutoMerge(err)).To(BeFalse())
		})
	})
})

// githubError is a simple error type for testing
type githubError string

func (e githubError) Error() string {
	return string(e)
}
