package permissions_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bartsmykla/smyklot/pkg/permissions"
)

var _ = Describe("Permission Checker [Unit]", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "smyklot-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("NewChecker", func() {
		It("should create a new checker with valid repo path", func() {
			checker, err := permissions.NewChecker(tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(checker).NotTo(BeNil())
		})

		It("should return error for empty repo path", func() {
			_, err := permissions.NewChecker("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`(?i)empty.*repository.*path`))
		})

		It("should return error for non-existent repo path", func() {
			_, err := permissions.NewChecker("/nonexistent/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`(?i)repository.*path.*does not exist`))
		})
	})

	Describe("CanApprove", func() {
		var checker *permissions.Checker

		Context("when root OWNERS file exists", func() {
			BeforeEach(func() {
				// Create root OWNERS file
				content := `approvers:
  - admin1
  - admin2
  - root-user
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow root approver to approve", func() {
				canApprove, err := checker.CanApprove("admin1", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should allow another root approver to approve", func() {
				canApprove, err := checker.CanApprove("root-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should deny non-approver", func() {
				canApprove, err := checker.CanApprove("random-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})

			It("should handle empty username", func() {
				canApprove, err := checker.CanApprove("", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})

			It("should be case-sensitive for usernames", func() {
				canApprove, err := checker.CanApprove("ADMIN1", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})

			It("should allow approval for root path", func() {
				canApprove, err := checker.CanApprove("admin1", "/")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should allow approval for nested paths", func() {
				canApprove, err := checker.CanApprove("admin1", "/pkg/module/file.go")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should allow approval for empty path (defaults to root)", func() {
				canApprove, err := checker.CanApprove("admin1", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})
		})

		Context("when root OWNERS file does not exist", func() {
			BeforeEach(func() {
				var err error
				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should deny all users when no OWNERS file exists", func() {
				canApprove, err := checker.CanApprove("any-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})
		})

		Context("when root OWNERS file has empty approvers list", func() {
			BeforeEach(func() {
				// Create OWNERS file with empty approvers
				content := `approvers: []
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should deny all users when approvers list is empty", func() {
				canApprove, err := checker.CanApprove("any-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})
		})

		Context("when checking multiple users", func() {
			BeforeEach(func() {
				// Create root OWNERS file with multiple approvers
				content := `approvers:
  - alice
  - bob
  - charlie
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow multiple approvers", func() {
				canApprove, err := checker.CanApprove("alice", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())

				canApprove, err = checker.CanApprove("bob", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())

				canApprove, err = checker.CanApprove("charlie", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should deny non-approvers", func() {
				canApprove, err := checker.CanApprove("david", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())

				canApprove, err = checker.CanApprove("eve", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})
		})

		Context("when handling special characters in usernames", func() {
			BeforeEach(func() {
				content := `approvers:
  - user-with-dash
  - user_with_underscore
  - user123
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle usernames with dashes", func() {
				canApprove, err := checker.CanApprove("user-with-dash", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should handle usernames with underscores", func() {
				canApprove, err := checker.CanApprove("user_with_underscore", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should handle usernames with numbers", func() {
				canApprove, err := checker.CanApprove("user123", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})
		})
	})

	Describe("GetApprovers", func() {
		var checker *permissions.Checker

		Context("when root OWNERS file exists", func() {
			BeforeEach(func() {
				content := `approvers:
  - admin1
  - admin2
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return list of approvers", func() {
				approvers := checker.GetApprovers()
				Expect(approvers).To(Equal([]string{"admin1", "admin2"}))
			})
		})

		Context("when root OWNERS file does not exist", func() {
			BeforeEach(func() {
				var err error
				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return empty list", func() {
				approvers := checker.GetApprovers()
				Expect(approvers).To(BeEmpty())
			})
		})

		Context("when root OWNERS file has empty approvers", func() {
			BeforeEach(func() {
				content := `approvers: []
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return empty list", func() {
				approvers := checker.GetApprovers()
				Expect(approvers).To(BeEmpty())
			})
		})
	})
})
