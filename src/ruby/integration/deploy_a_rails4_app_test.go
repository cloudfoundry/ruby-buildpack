package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rails 4 App", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("in an offline environment", func() {
		BeforeEach(func() {
			SkipUnlessCached()
		})

		It("", func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails4"))
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy [/"))
		})

		AssertNoInternetTraffic("rails4")
	})

	Context("in an online environment", func() {
		BeforeEach(SkipUnlessUncached)

		It("app has dependencies", func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails4"))
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Installing node 10."))
			Expect(app.Stdout.String()).To(ContainSubstring("Download [https://"))

			Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
		})

		Context("app has non vendored dependencies", func() {
			It("", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails4_not_vendored"))
				Expect(filepath.Join(app.Path, "vendor")).ToNot(BeADirectory())

				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
			})

			AssertUsesProxyDuringStagingIfPresent("rails4_not_vendored")
		})
	})
})
