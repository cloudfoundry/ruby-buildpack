package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("pushing an app a second time", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "sinatra"))
	})

	It("uses the cache and runs", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).ToNot(ContainSubstring("Restoring vendor_bundle from cache"))
		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))

		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Restoring vendor_bundle from cache"))
		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
	})
})
