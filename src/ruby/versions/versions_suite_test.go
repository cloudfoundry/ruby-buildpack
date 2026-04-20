package versions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVersions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Versions Suite")
}
