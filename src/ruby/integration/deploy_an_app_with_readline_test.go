package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("in an online environment", func() {
		BeforeEach(func() {
			SkipUnlessUncached()
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_readline"))
		})

		It("", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
			Expect(app.Stdout.String()).ToNot(ContainSubstring("cannot open shared object file"))
		})
	})
})
