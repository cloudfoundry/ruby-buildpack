package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with No Gemfile", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(Fixtures("no_gemfile"))
	})

	Context("Single/Final buildpack", func() {
		BeforeEach(func() {
			app.Buildpacks = []string{"ruby_buildpack"}
		})

		It("fails in finalize", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
			Eventually(app.Stdout.String).Should(ContainSubstring("Gemfile.lock required"))
		})
	})

	Context("Supply buildpack", func() {
		BeforeEach(func() {
			if !ApiHasMultiBuildpack() {
				Skip("API does not have multi buildpack support")
			}
			app.Buildpacks = []string{"ruby_buildpack", "binary_buildpack"}
		})

		It("deploys", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby"))

			By("running with the supplied ruby version", func() {
				defaultRubyVersion := DefaultVersion("ruby")
				Expect(app.GetBody("/")).To(ContainSubstring("Ruby Version: " + defaultRubyVersion))
			})
		})
	})
})
