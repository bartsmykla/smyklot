package permissions_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/smyklot/pkg/permissions"
)

// mockGitHubClient is a mock implementation of GitHubClient for testing
type mockGitHubClient struct {
	isTeamMemberFunc func(ctx context.Context, org, teamSlug, username string) (bool, error)
}

func (m *mockGitHubClient) IsTeamMember(ctx context.Context, org, teamSlug, username string) (bool, error) {
	if m.isTeamMemberFunc != nil {
		return m.isTeamMemberFunc(ctx, org, teamSlug, username)
	}
	return false, nil
}

var _ = Describe("Permission Checker [Unit]", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "smyklot-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Describe("NewChecker", func() {
		It("should create a new checker with valid repo path", func() {
			checker, err := permissions.NewChecker(tempDir, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(checker).NotTo(BeNil())
		})

		It("should return error for empty repo path", func() {
			_, err := permissions.NewChecker("", nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, permissions.ErrEmptyRepoPath)).To(BeTrue())
		})

		It("should return error for non-existent repo path", func() {
			_, err := permissions.NewChecker("/nonexistent/path", nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, permissions.ErrRepoPathNotExist)).To(BeTrue())
		})
	})

	Describe("CanApprove", func() {
		var checker *permissions.Checker

		Context("when CODEOWNERS file exists", func() {
			BeforeEach(func() {
				// Create .github/CODEOWNERS file
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `* @admin1 @admin2 @root-user`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow global owner to approve", func() {
				canApprove, err := checker.CanApprove("admin1", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should allow another global owner to approve", func() {
				canApprove, err := checker.CanApprove("root-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should deny non-owner", func() {
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

		Context("when CODEOWNERS file does not exist", func() {
			BeforeEach(func() {
				var err error
				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should deny all users when no CODEOWNERS file exists", func() {
				canApprove, err := checker.CanApprove("any-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})
		})

		Context("when CODEOWNERS file has empty owners list", func() {
			BeforeEach(func() {
				// Create CODEOWNERS with no global owners
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `/docs/ @doc-team`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should deny all users when no global owners configured", func() {
				canApprove, err := checker.CanApprove("any-user", "/any/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})
		})

		Context("when checking multiple users", func() {
			BeforeEach(func() {
				// Create CODEOWNERS file with multiple global owners
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `* @alice @bob @charlie`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow multiple global owners", func() {
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

			It("should deny non-owners", func() {
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
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `* @user-with-dash @user_with_underscore @user123`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir, nil)
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

		Context("when CODEOWNERS contains team ownership", func() {
			BeforeEach(func() {
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `* @test-org/test-team @individual-user`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should approve team member when GitHub client is provided", func() {
				mockClient := &mockGitHubClient{
					isTeamMemberFunc: func(ctx context.Context, org, teamSlug, username string) (bool, error) {
						Expect(org).To(Equal("test-org"))
						Expect(teamSlug).To(Equal("test-team"))
						Expect(username).To(Equal("team-member"))
						return true, nil
					},
				}

				checker, err := permissions.NewChecker(tempDir, mockClient)
				Expect(err).NotTo(HaveOccurred())

				canApprove, err := checker.CanApprove("team-member", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should deny non-team member when GitHub client is provided", func() {
				mockClient := &mockGitHubClient{
					isTeamMemberFunc: func(ctx context.Context, org, teamSlug, username string) (bool, error) {
						return false, nil
					},
				}

				checker, err := permissions.NewChecker(tempDir, mockClient)
				Expect(err).NotTo(HaveOccurred())

				canApprove, err := checker.CanApprove("non-member", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())
			})

			It("should approve individual user even with teams", func() {
				mockClient := &mockGitHubClient{
					isTeamMemberFunc: func(ctx context.Context, org, teamSlug, username string) (bool, error) {
						return false, nil
					},
				}

				checker, err := permissions.NewChecker(tempDir, mockClient)
				Expect(err).NotTo(HaveOccurred())

				canApprove, err := checker.CanApprove("individual-user", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should skip team check when no GitHub client is provided", func() {
				checker, err := permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())

				// Without GitHub client, team membership cannot be checked
				canApprove, err := checker.CanApprove("team-member", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeFalse())

				// But individual users still work
				canApprove, err = checker.CanApprove("individual-user", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
			})

			It("should handle GitHub API errors", func() {
				mockClient := &mockGitHubClient{
					isTeamMemberFunc: func(ctx context.Context, org, teamSlug, username string) (bool, error) {
						return false, errors.New("API error")
					},
				}

				checker, err := permissions.NewChecker(tempDir, mockClient)
				Expect(err).NotTo(HaveOccurred())

				_, err = checker.CanApprove("team-member", "/path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API error"))
			})

			It("should check multiple teams", func() {
				githubDir := filepath.Join(tempDir, ".github")
				content := `* @org1/team1 @org2/team2 @individual-user`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err := os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				callCount := 0
				mockClient := &mockGitHubClient{
					isTeamMemberFunc: func(ctx context.Context, org, teamSlug, username string) (bool, error) {
						callCount++
						if org == "org2" && teamSlug == "team2" && username == "user" {
							return true, nil
						}
						return false, nil
					},
				}

				checker, err := permissions.NewChecker(tempDir, mockClient)
				Expect(err).NotTo(HaveOccurred())

				canApprove, err := checker.CanApprove("user", "/path")
				Expect(err).NotTo(HaveOccurred())
				Expect(canApprove).To(BeTrue())
				Expect(callCount).To(Equal(2)) // Should check both teams
			})
		})
	})

	Describe("GetApprovers", func() {
		var checker *permissions.Checker

		Context("when CODEOWNERS file exists", func() {
			BeforeEach(func() {
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `* @admin1 @admin2`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return list of global owners", func() {
				approvers := checker.GetApprovers()
				Expect(approvers).To(Equal([]string{"admin1", "admin2"}))
			})
		})

		Context("when CODEOWNERS file does not exist", func() {
			BeforeEach(func() {
				var err error
				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return empty list", func() {
				approvers := checker.GetApprovers()
				Expect(approvers).To(BeEmpty())
			})
		})

		Context("when CODEOWNERS file has no global owners", func() {
			BeforeEach(func() {
				githubDir := filepath.Join(tempDir, ".github")
				err := os.MkdirAll(githubDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `/docs/ @doc-team`
				codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
				err = os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				checker, err = permissions.NewChecker(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return empty list", func() {
				approvers := checker.GetApprovers()
				Expect(approvers).To(BeEmpty())
			})
		})
	})
})
