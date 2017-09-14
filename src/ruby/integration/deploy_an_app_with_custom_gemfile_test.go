package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with custom Gemfile", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "custom_gemfile"))
	})

	It("uses the version of ruby specified in Gemfile-APP", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 2.2.8"))
	})
})
