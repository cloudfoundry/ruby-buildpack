package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rails 5.1 (Webpack/Yarn) App", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails51"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("Installs node6 and runs", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Installing node 6."))

		Expect(app.GetBody("/")).To(ContainSubstring("Hello World"))
		Eventually(func() string { return app.Stdout.String() }, 10*time.Second).Should(ContainSubstring(`Started GET "/" for`))

		By("Make sure supply does not change BuildDir", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum Before Supply: 5d823d48d154ee2622e8cf8c2fb21ff7"))
			Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum After Supply: 5d823d48d154ee2622e8cf8c2fb21ff7"))
		})
	})
})
