package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JRuby App", func() {
	var app *cutlass.App

	AfterEach(func() { app = DestroyApp(app) })

	Context("without start command", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "sinatra_jruby"))
			app.Memory = "512M"
		})

		It("", func() {
			PushAppAndConfirm(app)

			By("the buildpack logged it installed a specific version of JRuby", func() {
				Expect(app.Stdout.String()).To(ContainSubstring("Installing openjdk"))
				Expect(app.Stdout.String()).To(MatchRegexp("ruby-2.3.\\d+-jruby-9.1.\\d+.0"))
				Expect(app.GetBody("/ruby")).To(MatchRegexp("jruby 2.3.\\d+"))
			})

			By("the OpenJDK runs properly", func() {
				Expect(app.Stdout.String()).ToNot(ContainSubstring("OpenJDK 64-Bit Server VM warning"))
			})
		})

		Context("a cached buildpack", func() {
			BeforeEach(SkipUnlessCached)

			AssertNoInternetTraffic("sinatra_jruby")
		})
	})
	Context("with a jruby start command", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "jruby_start_command"))
			app.Memory = "512M"
		})

		It("stages and runs successfully", func() {
			PushAppAndConfirm(app)
		})
	})
})
