package permissions_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/smyklot/pkg/permissions"
)

var _ = Describe("CODEOWNERS Parser [Unit]", func() {
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

	Describe("ParseCodeownersFile", func() {
		Context("when parsing valid CODEOWNERS files", func() {
			It("should parse a basic CODEOWNERS file with global owners", func() {
				content := `* @global-owner1 @global-owner2`
				codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
				err := os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(codeowners).NotTo(BeNil())
				Expect(codeowners.GetGlobalOwners()).To(Equal([]string{"global-owner1", "global-owner2"}))
			})

			It("should parse CODEOWNERS with multiple patterns", func() {
				content := `* @global-owner
/docs/ @doc-team
*.js @js-team`
				codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
				err := os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(codeowners.Entries).To(HaveLen(3))
				Expect(codeowners.GetGlobalOwners()).To(Equal([]string{"global-owner"}))
			})

			It("should skip empty lines and comments", func() {
				content := `# This is a comment

* @owner1 @owner2

# Another comment
/docs/ @doc-team`
				codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
				err := os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(codeowners.Entries).To(HaveLen(2))
			})

			It("should handle lines without @ prefix", func() {
				content := `* owner1 @owner2`
				codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
				err := os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
				Expect(err).NotTo(HaveOccurred())
				// Should only include owners with @ prefix
				Expect(codeowners.GetGlobalOwners()).To(Equal([]string{"owner2"}))
			})

			It("should handle empty CODEOWNERS file", func() {
				content := ``
				codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
				err := os.WriteFile(codeownersPath, []byte(content), 0600)
				Expect(err).NotTo(HaveOccurred())

				codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(codeowners.Entries).To(BeEmpty())
				Expect(codeowners.GetGlobalOwners()).To(BeEmpty())
			})
		})

		Context("when handling errors", func() {
			It("should return error for empty file path", func() {
				_, err := permissions.ParseCodeownersFile("")
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, permissions.ErrEmptyFilePath)).To(BeTrue())
			})

			It("should return error for non-existent file", func() {
				_, err := permissions.ParseCodeownersFile("/nonexistent/CODEOWNERS")
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, permissions.ErrReadFailed)).To(BeTrue())
			})
		})
	})

	Describe("GetAllOwners", func() {
		It("should return all unique owners from all patterns", func() {
			content := `* @owner1 @owner2
/docs/ @owner2 @owner3
*.js @owner1`
			codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
			err := os.WriteFile(codeownersPath, []byte(content), 0600)
			Expect(err).NotTo(HaveOccurred())

			codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
			Expect(err).NotTo(HaveOccurred())

			allOwners := codeowners.GetAllOwners()
			Expect(allOwners).To(ConsistOf("owner1", "owner2", "owner3"))
		})

		It("should return empty list for empty CODEOWNERS", func() {
			content := ``
			codeownersPath := filepath.Join(tempDir, "CODEOWNERS")
			err := os.WriteFile(codeownersPath, []byte(content), 0600)
			Expect(err).NotTo(HaveOccurred())

			codeowners, err := permissions.ParseCodeownersFile(codeownersPath)
			Expect(err).NotTo(HaveOccurred())

			allOwners := codeowners.GetAllOwners()
			Expect(allOwners).To(BeEmpty())
		})
	})
})
