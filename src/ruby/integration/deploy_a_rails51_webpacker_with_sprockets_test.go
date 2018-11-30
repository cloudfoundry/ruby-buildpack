package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing a rails51 webpacker app with sprockets", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails51_webpacker"))
	})

	It("compiles assets with webpacker", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Webpacker is installed"))
		Expect(app.Stdout.String()).To(ContainSubstring("Asset precompilation completed"))

		Expect(app.GetBody("/")).To(ContainSubstring("Welcome to Rails51 Webpacker!"))
		Eventually(app.Stdout.String()).Should(ContainSubstring("Cleaning assets"))
	})
})
