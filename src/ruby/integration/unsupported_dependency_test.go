package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "unsupported_ruby"))
	})

	It("displays a nice error message when Ruby 99.99.99 is specified", func() {
		Expect(app.Push()).ToNot(Succeed())
		Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
		Eventually(app.Stdout.String).Should(ContainSubstring("No Matching versions, ruby = 99.99.99 not found in this buildpack"))
	})
})
