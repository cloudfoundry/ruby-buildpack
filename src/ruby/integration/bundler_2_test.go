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
		app = cutlass.New(Fixtures("bundler_2_0_1"))
		app.SetEnv("BP_DEBUG", "1")
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Using bundler 2"))
		Expect(app.Stdout.String()).To(ContainSubstring(`Deleting "Bundled With" from the Gemfile.lock`))
	})

	It("works with current version of bundler 2", func() {
		app = cutlass.New(Fixtures("bundler_2_0_2"))
		app.SetEnv("BP_DEBUG", "1")
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Using bundler 2"))
		Expect(app.Stdout.String()).NotTo(ContainSubstring(`Deleting "Bundled With" from the Gemfile.lock`))
	})
})
