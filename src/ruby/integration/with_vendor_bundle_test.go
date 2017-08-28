package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with dependencies installed in vendor/bundle", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_vendor_bundle"))
	})

	It("", func() {
		PushAppAndConfirm(app)

		By("remove vendor/bundle directory", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Removing `vendor/bundle`"))
			Expect(app.Stdout.String()).To(ContainSubstring("Checking in `vendor/bundle` is not supported. Please remove this directory and add it to your .gitignore. To vendor your gems with Bundler, use `bundle pack` instead."))

			files, err := app.Files("app/vendor")
			Expect(err).ToNot(HaveOccurred())
			Expect(files).ToNot(ContainElement("app/vendor/bundle"))
		})

		By("has required gems at runtime", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Healthy"))
			Eventually(func() string { return app.Stdout.String() }, 5*time.Second).Should(ContainSubstring("This is red"))
			Eventually(func() string { return app.Stdout.String() }).Should(ContainSubstring("This is blue"))
		})
	})
})
