package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing a rails webpacker app with sprockets", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("rails6"))
		app.Disk = "1G"
	})

	It("compiles assets with webpacker", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Asset precompilation completed"))

		Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		Eventually(app.Stdout.String()).Should(ContainSubstring("Cleaning assets"))
	})
})
