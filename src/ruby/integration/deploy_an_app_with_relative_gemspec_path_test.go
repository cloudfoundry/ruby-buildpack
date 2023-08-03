package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with relative gemspec path", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("relative_gemspec_path"))
	})

	It("loads the gem with the relative gemspec path", func() {
		PushAppAndConfirm(app)
		Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
	})
})
