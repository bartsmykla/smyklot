package permissions_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bartsmykla/smyklot/pkg/permissions"
)

var _ = Describe("OWNERS File Parser [Unit]", func() {
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

	Describe("ParseOwnersFile", func() {
		Context("when parsing valid OWNERS files", func() {
			It("should parse a basic OWNERS file", func() {
				content := `approvers:
  - user1
  - user2
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"user1", "user2"}))
				Expect(owners.Path).To(Equal(tempDir))
			})

			It("should parse OWNERS file with single approver", func() {
				content := `approvers:
  - admin
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"admin"}))
			})

			It("should handle OWNERS file with YAML comments", func() {
				content := `# Root approvers
approvers:
  - user1  # Primary maintainer
  - user2  # Secondary maintainer
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"user1", "user2"}))
			})

			It("should handle OWNERS file with extra whitespace", func() {
				content := `approvers:
  - user1
  - user2

`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"user1", "user2"}))
			})

			It("should set correct path for root OWNERS file", func() {
				content := `approvers:
  - root-admin
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Path).To(Equal(tempDir))
			})

			It("should set correct path for nested OWNERS file", func() {
				subDir := filepath.Join(tempDir, "pkg", "module")
				err := os.MkdirAll(subDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				content := `approvers:
  - module-owner
`
				ownersPath := filepath.Join(subDir, "OWNERS")
				err = os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Path).To(Equal(subDir))
				Expect(owners.Approvers).To(Equal([]string{"module-owner"}))
			})
		})

		Context("when handling empty approvers", func() {
			It("should return empty list for OWNERS file with no approvers", func() {
				content := `approvers: []
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(BeEmpty())
			})

			It("should return empty list for OWNERS file with only whitespace approvers", func() {
				content := `approvers:
  - ""
  - "  "
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(BeEmpty())
			})
		})

		Context("when handling invalid YAML", func() {
			It("should return error for invalid YAML syntax", func() {
				content := `approvers:
  - user1
  - user2
  - : invalid
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				_, err = permissions.ParseOwnersFile(ownersPath)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, permissions.ErrInvalidYAML)).To(BeTrue())
			})

			It("should return error for non-YAML content", func() {
				content := `This is not YAML at all`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				_, err = permissions.ParseOwnersFile(ownersPath)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for YAML without approvers field", func() {
				content := `maintainers:
  - user1
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(BeEmpty())
			})
		})

		Context("when handling file errors", func() {
			It("should return error for non-existent file", func() {
				ownersPath := filepath.Join(tempDir, "NONEXISTENT")

				_, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, permissions.ErrReadFailed)).To(BeTrue())
			})

			It("should return error for directory instead of file", func() {
				dirPath := filepath.Join(tempDir, "OWNERS")
				err := os.Mkdir(dirPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				_, err = permissions.ParseOwnersFile(dirPath)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, permissions.ErrReadFailed)).To(BeTrue())
			})

			It("should return error for empty file path", func() {
				_, err := permissions.ParseOwnersFile("")
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, permissions.ErrEmptyFilePath)).To(BeTrue())
			})
		})

		Context("when handling special characters in usernames", func() {
			It("should handle usernames with hyphens", func() {
				content := `approvers:
  - user-name-1
  - user-name-2
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"user-name-1", "user-name-2"}))
			})

			It("should handle usernames with numbers", func() {
				content := `approvers:
  - user123
  - admin2024
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"user123", "admin2024"}))
			})
		})

		Context("when handling duplicate approvers", func() {
			It("should preserve duplicate approvers", func() {
				content := `approvers:
  - user1
  - user2
  - user1
`
				ownersPath := filepath.Join(tempDir, "OWNERS")
				err := os.WriteFile(ownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				owners, err := permissions.ParseOwnersFile(ownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(owners.Approvers).To(Equal([]string{"user1", "user2", "user1"}))
			})
		})
	})
})
