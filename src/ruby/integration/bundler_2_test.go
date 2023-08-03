package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() { app = DestroyApp(app) })

	It("works with old version of bundler 2", func() {
		app = cutlass.New(Fixtures("bundler_2"))
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(MatchRegexp(`Your Gemfile.lock was bundled with bundler 2\.\d+\.\d+, which is incompatible with the current bundler version \(2\.\d+\.\d+\)`))
		Expect(app.Stdout.String()).To(ContainSubstring(`Deleting "Bundled With" from the Gemfile.lock`))
		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
	})
})
