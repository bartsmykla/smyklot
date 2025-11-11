package feedback_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFeedback(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Feedback Suite")
}
