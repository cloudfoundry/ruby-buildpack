package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rails 5.2 (Webpack/Yarn) App", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("rails52_webpack_no_sprockets"))
	})

	It("Installs node and runs", func() {
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Hello, Rails!"))
		Eventually(app.Stdout.String).Should(ContainSubstring(`Started GET "/" for`))

		Eventually(app.Stdout.String).ShouldNot(ContainSubstring("Cleaning assets"))

	})
})
