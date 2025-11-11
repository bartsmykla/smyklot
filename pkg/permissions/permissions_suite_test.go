package permissions_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPermissions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Permissions Suite")
}
