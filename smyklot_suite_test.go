package smyklot_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSmyklot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smyklot Suite")
}
