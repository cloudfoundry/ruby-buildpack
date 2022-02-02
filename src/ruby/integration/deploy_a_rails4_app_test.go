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
			app = cutlass.New(Fixtures("rails4"))
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy [/"))
		})

		AssertNoInternetTraffic("rails4")
	})

	Context("in an online environment", func() {
		BeforeEach(SkipUnlessUncached)

		It("app has dependencies", func() {
			app = cutlass.New(Fixtures("rails4"))
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Installing node"))
			Expect(app.Stdout.String()).To(ContainSubstring("Download [https://"))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
		})

		Context("app has non vendored dependencies", func() {
			It("", func() {
				app = cutlass.New(Fixtures("rails4_not_vendored"))
				Expect(filepath.Join(app.Path, "vendor")).ToNot(BeADirectory())

				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
			})

			AssertUsesProxyDuringStagingIfPresent("rails4_not_vendored")
		})
	})
})
