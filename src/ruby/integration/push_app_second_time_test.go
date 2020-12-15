package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app a second time", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("sinatra"))
		app.Buildpacks = []string{"ruby_buildpack"}
	})

	RestoringVendorBundle := "Restoring vendor_bundle from cache"
	DownloadRegexp := `Download \[.*/bundler[\-_].*\.tgz\]`
	CopyRegexp := `Copy \[.*/bundler\-.*\.tgz\]`

	It("uses the cache and runs", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).ToNot(ContainSubstring(RestoringVendorBundle))
		if !cutlass.Cached {
			Expect(app.Stdout.String()).To(MatchRegexp(DownloadRegexp))
			Expect(app.Stdout.String()).ToNot(MatchRegexp(CopyRegexp))
		}
		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))

		app.Stdout.Reset()
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring(RestoringVendorBundle))
		if !cutlass.Cached {
			Expect(app.Stdout.String()).To(MatchRegexp(CopyRegexp))
			Expect(app.Stdout.String()).ToNot(MatchRegexp(DownloadRegexp))
		}
		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))

		app.Stdout.Reset()
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring(RestoringVendorBundle))
		if !cutlass.Cached {
			Expect(app.Stdout.String()).To(MatchRegexp(CopyRegexp))
			Expect(app.Stdout.String()).ToNot(MatchRegexp(DownloadRegexp))
		}
		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
	})
})
