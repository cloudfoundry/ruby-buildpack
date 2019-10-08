package integration_test

import (
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

	Describe("When the name of the Gemfile is specified in the manifest via BUNDLE_GEMFILE", func() {
		BeforeEach(func() {
			app = cutlass.New(Fixtures("custom_gemfile"))
		})

		It("detects the ruby buildpack and uses the version of ruby specified in Gemfile-APP", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 2.6"))
			Expect(app.Stdout.String()).To(ContainSubstring("Installing sinatra 1.4.7"))
		})
	})

	Describe("When the name of the Gemfile is improperly specified in the manifest", func() {
		BeforeEach(func() {
			app = cutlass.New(Fixtures("custom_gemfile_bad_manifest"))
		})

		It("fails to stage", func() {
			Expect(app.Push()).NotTo(Succeed())
		})
	})
})
